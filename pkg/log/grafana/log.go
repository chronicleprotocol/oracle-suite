//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const LoggerTag = "GRAFANA"

var argRegexp = regexp.MustCompile(`\$\{[^}]+\}`)

type Config struct {
	Metrics          []Metric
	Interval         uint
	GraphiteEndpoint string
	GraphiteAPIKey   string
	HTTPClient       *http.Client
	Logger           log.Logger
}

// Logger implements the log.Logger interface.
type Logger struct {
	*shared
	level  log.Level
	fields log.Fields
}

type Metric struct {
	Pattern *regexp.Regexp
	Value   string
	Name    string
	Tags    map[string][]string
}

type shared struct {
	mu               sync.Mutex //nolint:structcheck // false-positive
	logger           log.Logger
	metrics          []Metric
	interval         uint
	graphiteEndpoint string
	graphiteAPIKey   string
	httpClient       *http.Client
	metricPoints     map[metricKey]metricValue
}

type metricKey struct {
	name string
	time int64
}

type metricValue struct {
	value float64
	tags  []string
}

type metricJSON struct {
	Name     string   `json:"name"`
	Interval uint     `json:"interval"`
	Value    float64  `json:"value"`
	Time     int64    `json:"time"`
	Tags     []string `json:"tags,omitempty"`
}

func New(level log.Level, cfg Config) *Logger {
	l := &Logger{
		shared: &shared{
			metrics:          cfg.Metrics,
			logger:           cfg.Logger.WithField("tag", LoggerTag),
			interval:         cfg.Interval,
			graphiteEndpoint: cfg.GraphiteEndpoint,
			graphiteAPIKey:   cfg.GraphiteAPIKey,
			httpClient:       cfg.HTTPClient,
			metricPoints:     make(map[metricKey]metricValue, 0),
		},
		level:  level,
		fields: log.Fields{},
	}
	go l.pushRoutine()
	return l
}

func (c *Logger) Level() log.Level {
	return c.level
}

func (c *Logger) WithField(key string, value interface{}) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	f[key] = value
	return &Logger{
		shared: c.shared,
		level:  c.level,
		fields: f,
	}
}

func (c *Logger) WithFields(fields log.Fields) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	for k, v := range fields {
		f[k] = v
	}
	return &Logger{
		shared: c.shared,
		level:  c.level,
		fields: f,
	}
}

func (c *Logger) WithError(err error) log.Logger {
	return c.WithField("err", err.Error())
}

func (c *Logger) Debugf(format string, args ...interface{}) {
	if c.level >= log.Debug {
		c.collect(fmt.Sprintf(format, args...), c.fields)
	}
}

func (c *Logger) Infof(format string, args ...interface{}) {
	if c.level >= log.Info {
		c.collect(fmt.Sprintf(format, args...), c.fields)
	}
}

func (c *Logger) Warnf(format string, args ...interface{}) {
	if c.level >= log.Warn {
		c.collect(fmt.Sprintf(format, args...), c.fields)
	}
}

func (c *Logger) Errorf(format string, args ...interface{}) {
	if c.level >= log.Error {
		c.collect(fmt.Sprintf(format, args...), c.fields)
	}
}

func (c *Logger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	c.collect(msg, c.fields)
	panic(msg)
}

func (c *Logger) Debug(args ...interface{}) {
	if c.level >= log.Debug {
		c.collect(fmt.Sprint(args...), c.fields)
	}
}

func (c *Logger) Info(args ...interface{}) {
	if c.level >= log.Info {
		c.collect(fmt.Sprint(args...), c.fields)
	}
}

func (c *Logger) Warn(args ...interface{}) {
	if c.level >= log.Warn {
		c.collect(fmt.Sprint(args...), c.fields)
	}
}

func (c *Logger) Error(args ...interface{}) {
	if c.level >= log.Error {
		c.collect(fmt.Sprint(args...), c.fields)
	}
}

func (c *Logger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	c.collect(msg, c.fields)
	panic(msg)
}

func (c *Logger) collect(msg string, fields log.Fields) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, metric := range c.metrics {
		if metric.Pattern.MatchString(msg) {
			var tags []string
			value := float64(1)
			timestamp := roundTime(time.Now().Unix(), c.interval)
			if len(metric.Value) > 0 {
				var ok bool
				value, ok = toFloat(byPath(reflect.ValueOf(fields), metric.Value))
				if !ok {
					c.logger.WithField("path", metric.Value).Warn("Invalid path")
					continue
				}
			}
			name := argRegexp.ReplaceAllStringFunc(metric.Name, func(s string) string {
				path := s[2 : len(s)-1]
				value, ok := toString(byPath(reflect.ValueOf(fields), path))
				if !ok {
					c.logger.WithField("path", path).Warn("Invalid path")
					return ""
				}
				return value
			})
			for k, vs := range metric.Tags {
				for _, v := range vs {
					tags = append(tags, k+"="+v)
				}
			}
			c.metricPoints[metricKey{time: timestamp, name: name}] = metricValue{value: value, tags: tags}
			c.logger.
				WithFields(log.Fields{"timestamp": timestamp, "name": name, "value": value, "tags": tags}).
				Debug("New metric point")
		}
	}
}

func (c *Logger) pushRoutine() {
	ticker := time.NewTicker(time.Duration(c.interval) * time.Second)
	for {
		<-ticker.C
		c.pushMetrics()
	}
}

func (c *Logger) pushMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.metricPoints) == 0 {
		return
	}
	var ps []metricJSON
	for k, v := range c.metricPoints {
		ps = append(ps, metricJSON{
			Name:     k.name,
			Interval: c.interval,
			Value:    v.value,
			Time:     k.time,
			Tags:     v.tags,
		})
		delete(c.metricPoints, k)
	}
	j, err := json.Marshal(ps)
	if err != nil {
		c.logger.WithError(err).Warn("Unable to marshall metric points")
		return
	}
	req, err := http.NewRequest("POST", c.graphiteEndpoint, bytes.NewReader(j))
	if err != nil {
		c.logger.WithError(err).Warn("Invalid request")
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.graphiteAPIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Warn("Invalid request")
		return
	}
	err = res.Body.Close()
	if err != nil {
		c.logger.WithError(err).Warn("Unable to close request body")
		return
	}
}

func byPath(v reflect.Value, path string) reflect.Value {
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		return byPath(v.Elem(), path)
	}
	if len(path) == 0 {
		return v
	}
	switch v.Kind() {
	case reflect.Slice:
		elem, path := splitPath(path)
		i, err := strconv.Atoi(elem)
		if err != nil {
			return reflect.Value{}
		}
		f := v.Index(i)
		if !f.IsValid() {
			return reflect.Value{}
		}
		return byPath(f, path)
	case reflect.Map:
		elem, path := splitPath(path)
		if v.Type().Key().Kind() != reflect.String {
			return reflect.Value{}
		}
		f := v.MapIndex(reflect.ValueOf(elem))
		if !f.IsValid() {
			return reflect.Value{}
		}
		return byPath(f, path)
	case reflect.Struct:
		elem, path := splitPath(path)
		f := v.FieldByName(elem)
		if !f.IsValid() {
			return reflect.Value{}
		}
		return byPath(f, path)
	}
	return reflect.Value{}
}

func splitPath(path string) (a, b string) {
	p := strings.SplitN(path, ".", 2)
	switch len(p) {
	case 1:
		return p[0], ""
	case 2:
		return p[0], p[1]
	default:
		return "", ""
	}
}

func toFloat(value reflect.Value) (float64, bool) {
	if !value.IsValid() {
		return 0, false
	}
	switch value.Type().Kind() {
	case reflect.String:
		f, err := strconv.ParseFloat(value.String(), 64)
		return f, err == nil
	case reflect.Float32, reflect.Float64:
		return value.Float(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint()), true
	}
	return 0, false
}

func toString(value reflect.Value) (string, bool) {
	if !value.IsValid() {
		return "", false
	}
	switch value.Type().Kind() {
	case reflect.String:
		return value.String(), true
	case reflect.Float32, reflect.Float64:
		return fmt.Sprint(value.Float()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprint(value.Uint()), true
	}
	return fmt.Sprint(value.Interface()), true
}

func roundTime(time int64, interval uint) int64 {
	return time - (time % int64(interval))
}
