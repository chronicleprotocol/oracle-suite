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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/dump"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/interpolate"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const LoggerTag = "GRAFANA"

// Config is the configuration for the Grafana logger.
type Config struct {
	// Metrics is a list of metric definitions.
	Metrics []Metric
	// Interval specifies how often logs should be sent to the Grafana Cloud
	// server. Logs with the same name in that interval will be replaced with
	// never ones.
	Interval uint
	// Graphite server endpoint.
	GraphiteEndpoint string
	// Graphite API key.
	GraphiteAPIKey string
	// HTTPClient used to send metrics to Grafana Cloud.
	HTTPClient *http.Client
	// Logger used to log errors related to this logger, such as connection errors.
	Logger log.Logger
}

// Metric describes one Grafana metric.
type Metric struct {
	// MatchMessage is a regexp that must match the log message.
	MatchMessage *regexp.Regexp
	// MatchFields is a list of regexp's that must match the values of the
	// fields defined in the map keys.
	MatchFields map[string]*regexp.Regexp
	// Value is the dot-separated path of the field with the metric value.
	// If empty, the value 1 will be used as the metric value.
	Value string
	// Name is the name of the metric. It can contain references to log fields
	// in the format ${path}, where path is the dot-separated path to the field.
	Name string
	// Tag is a list of metric tags. They can contain references to log fields
	// in the format ${path}, where path is the dot-separated path to the field.
	Tags map[string][]string
	// OnDuplicate specifies how duplicated values in the same interval should
	// be handled.
	OnDuplicate OnDuplicate
	// TransformFunc defines the function applied to the value before setting it.
	TransformFunc func(float64) float64
	// ParserFunc is going to be applied to transform the value reflection to an actual float64 value
	ParserFunc func(reflect.Value) (float64, bool)

	parsedName interpolate.Parsed
	parsedTags map[string][]interpolate.Parsed
}

// New creates a new logger that can extract parameters from log messages and
// send them to Grafana Cloud using the Graphite endpoint. It starts
// a background goroutine that will be sending metrics to the Grafana Cloud
// server as often as described in Config.Interval parameter.
func New(level log.Level, cfg Config) (log.Logger, error) {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	// Parse names and tags in advance to improve performance.
	for n := range cfg.Metrics {
		m := &cfg.Metrics[n]
		m.parsedName = interpolate.Parse(m.Name)
		m.parsedTags = make(map[string][]interpolate.Parsed)
		for k, v := range m.Tags {
			for _, s := range v {
				m.parsedTags[k] = append(m.parsedTags[k], interpolate.Parse(s))
			}
		}
	}
	l := &logger{
		shared: &shared{
			waitCh:           make(chan error),
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
	return l, nil
}

type logger struct {
	*shared
	level  log.Level
	fields log.Fields
}

type shared struct {
	mu     sync.Mutex
	ctx    context.Context
	waitCh chan error

	logger           log.Logger
	metrics          []Metric
	interval         uint
	graphiteEndpoint string
	graphiteAPIKey   string
	httpClient       *http.Client
	metricPoints     map[metricKey]metricValue
}

type OnDuplicate int

const (
	Replace OnDuplicate = iota // Replace the current value.
	Ignore                     // Use previous value.
	Sum                        // Add to the current value.
	Sub                        // Subtract from the current value.
	Max                        // Use higher value.
	Min                        // Use lower value.
)

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

// Level implements the log.Logger interface.
func (c *logger) Level() log.Level {
	return c.level
}

// WithField implements the log.Logger interface.
func (c *logger) WithField(key string, value interface{}) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	f[key] = value
	return &logger{
		shared: c.shared,
		level:  c.level,
		fields: f,
	}
}

// WithFields implements the log.Logger interface.
func (c *logger) WithFields(fields log.Fields) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	for k, v := range fields {
		f[k] = v
	}
	return &logger{
		shared: c.shared,
		level:  c.level,
		fields: f,
	}
}

// WithError implements the log.Logger interface.
func (c *logger) WithError(err error) log.Logger {
	return c.WithField("err", err.Error())
}

// Debugf implements the log.Logger interface.
func (c *logger) Debugf(format string, args ...interface{}) {
	if c.level >= log.Debug {
		c.collect(fmt.Sprintf(format, args...), c.fields)
	}
}

// Infof implements the log.Logger interface.
func (c *logger) Infof(format string, args ...interface{}) {
	if c.level >= log.Info {
		c.collect(fmt.Sprintf(format, args...), c.fields)
	}
}

// Warnf implements the log.Logger interface.
func (c *logger) Warnf(format string, args ...interface{}) {
	if c.level >= log.Warn {
		c.collect(fmt.Sprintf(format, args...), c.fields)
	}
}

// Errorf implements the log.Logger interface.
func (c *logger) Errorf(format string, args ...interface{}) {
	if c.level >= log.Error {
		c.collect(fmt.Sprintf(format, args...), c.fields)
	}
}

// Panicf implements the log.Logger interface.
func (c *logger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	c.collect(msg, c.fields)
	c.pushMetrics() // force push metrics before app crash
	panic(msg)
}

// Debug implements the log.Logger interface.
func (c *logger) Debug(args ...interface{}) {
	if c.level >= log.Debug {
		c.collect(fmt.Sprint(args...), c.fields)
	}
}

// Info implements the log.Logger interface.
func (c *logger) Info(args ...interface{}) {
	if c.level >= log.Info {
		c.collect(fmt.Sprint(args...), c.fields)
	}
}

// Warn implements the log.Logger interface.
func (c *logger) Warn(args ...interface{}) {
	if c.level >= log.Warn {
		c.collect(fmt.Sprint(args...), c.fields)
	}
}

// Error implements the log.Logger interface.
func (c *logger) Error(args ...interface{}) {
	if c.level >= log.Error {
		c.collect(fmt.Sprint(args...), c.fields)
	}
}

// Panic implements the log.Logger interface.
func (c *logger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	c.collect(msg, c.fields)
	c.pushMetrics() // force push metrics before app crash
	panic(msg)
}

// collect checks if a log matches any of predefined metrics and if so,
// extracts a metric value from it.
func (c *logger) collect(msg string, fields log.Fields) {
	c.mu.Lock()
	defer c.mu.Unlock()
	rfields := reflect.ValueOf(fields)
	for _, metric := range c.metrics {
		if !match(metric, msg, rfields) {
			continue
		}
		var ok bool
		var mk metricKey
		var mv metricValue
		mk.time = roundTime(time.Now().Unix(), c.interval)
		mv.value = 1
		mk.name, ok = replaceVars(metric.parsedName, rfields)
		if !ok {
			c.logger.
				WithField("path", metric.Name).
				Warn("Invalid path in the name field")
			continue
		}
		for t, vs := range metric.parsedTags {
			for _, v := range vs {
				rt, ok := replaceVars(v, rfields)
				if !ok {
					c.logger.
						WithField("path", v).
						Warn("Invalid path in the tag definition")
					continue
				}
				mv.tags = append(mv.tags, t+"="+rt)
			}
		}
		if len(metric.Value) > 0 {
			value := byPath(rfields, metric.Value)
			if metric.ParserFunc != nil {
				mv.value, ok = metric.ParserFunc(value)
			} else {
				mv.value, ok = toFloat(value)
			}
			if !ok {
				c.logger.
					WithField("path", metric.Value).
					Warn("Invalid path")
				continue
			}
		}
		if metric.TransformFunc != nil {
			mv.value = metric.TransformFunc(mv.value)
		}
		c.addMetricPoint(metric, mk, mv)
	}
}

func (c *logger) addMetricPoint(m Metric, mk metricKey, mv metricValue) {
	cv, exists := c.metricPoints[mk]
	switch {
	case exists && m.OnDuplicate == Replace:
		c.metricPoints[mk] = mv
	case exists && m.OnDuplicate == Sum:
		mv.value += cv.value
		c.metricPoints[mk] = mv
	case exists && m.OnDuplicate == Sub:
		mv.value -= cv.value
		c.metricPoints[mk] = mv
	case exists && m.OnDuplicate == Min:
		if mv.value < cv.value {
			c.metricPoints[mk] = mv
		}
	case exists && m.OnDuplicate == Max:
		if mv.value > cv.value {
			c.metricPoints[mk] = mv
		}
	case exists && m.OnDuplicate == Ignore:
	default:
		c.metricPoints[mk] = mv
	}
	c.logger.
		WithFields(log.Fields{
			"timestamp": mk.time,
			"name":      mk.name,
			"value":     mv.value,
			"tags":      mv.tags,
		}).
		Debug("New metric point")
}

// pushRoutine pushes metrics in interval defined in c.interval.
func (c *logger) pushRoutine() {
	defer c.logger.Info("Stopped")
	defer close(c.waitCh)
	defer c.pushMetrics()
	ticker := time.NewTicker(time.Duration(c.interval) * time.Second)
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.pushMetrics()
		}
	}
}

// pushMetrics pushes metrics to the Grafana Cloud server.
func (c *logger) pushMetrics() {
	var once sync.Once
	c.mu.Lock()
	defer once.Do(c.mu.Unlock)
	if len(c.metricPoints) == 0 {
		return
	}
	var metrics []metricJSON
	for k, v := range c.metricPoints {
		metrics = append(metrics, metricJSON{
			Name:     k.name,
			Interval: c.interval,
			Value:    v.value,
			Time:     k.time,
			Tags:     v.tags,
		})
		delete(c.metricPoints, k)
	}
	// After this line, the mutex must be unlocked, otherwise calling
	// the logger will cause the code to block until the request to
	// Grafana is complete.
	once.Do(c.mu.Unlock)
	reqBody, err := json.Marshal(metrics)
	if err != nil {
		c.logger.
			WithError(err).
			Warn("Unable to marshall metric points")
		return
	}
	c.logger.
		WithField("metrics", len(metrics)).
		Info("Pushing metrics")
	req, err := http.NewRequest("POST", c.graphiteEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		c.logger.
			WithError(err).
			Warn("Invalid request")
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.graphiteAPIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.
			WithError(err).
			Warn("Invalid request")
		return
	}
	_, _ = io.ReadAll(res.Body)
	err = res.Body.Close()
	if err != nil {
		c.logger.
			WithError(err).
			Warn("Unable to close request body")
		return
	}
}

// Start implements the supervisor.Service interface.
func (c *logger) Start(ctx context.Context) error {
	c.logger.Info("Starting")
	if c.ctx != nil {
		return fmt.Errorf("service can be started only once")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	c.ctx = ctx
	go c.pushRoutine()
	return nil
}

// Wait implements the supervisor.Service interface.
func (c *logger) Wait() chan error {
	return c.waitCh
}

// match checks if log message and log fields matches metric definition.
func match(metric Metric, msg string, fields reflect.Value) bool {
	for path, rx := range metric.MatchFields {
		field, ok := toString(byPath(fields, path))
		if !ok || !rx.MatchString(field) {
			return false
		}
	}
	return metric.MatchMessage == nil || metric.MatchMessage.MatchString(msg)
}

// replaceVars replaces vars provided as ${field} with values from log fields.
func replaceVars(s interpolate.Parsed, fields reflect.Value) (string, bool) {
	valid := true
	return s.Interpolate(func(v interpolate.Variable) string {
		name, ok := toString(byPath(fields, v.Name))
		if !ok {
			if v.HasDefault {
				return v.Default
			}
			valid = false
			return ""
		}
		return name
	}), valid
}

func byPath(value reflect.Value, path string) reflect.Value {
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		return byPath(value.Elem(), path)
	}
	if len(path) == 0 {
		return value
	}
	switch value.Kind() {
	case reflect.Slice:
		elem, path := splitPath(path)
		i, err := strconv.Atoi(elem)
		if err != nil {
			return reflect.Value{}
		}
		f := value.Index(i)
		if !f.IsValid() {
			return reflect.Value{}
		}
		return byPath(f, path)
	case reflect.Map:
		elem, path := splitPath(path)
		if value.Type().Key().Kind() != reflect.String {
			return reflect.Value{}
		}
		f := value.MapIndex(reflect.ValueOf(elem))
		if !f.IsValid() {
			return reflect.Value{}
		}
		return byPath(f, path)
	case reflect.Struct:
		elem, path := splitPath(path)
		f := value.FieldByName(elem)
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
	if t, ok := value.Interface().(time.Duration); ok {
		return t.Seconds(), true
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
	return fmt.Sprint(dump.Dump(value.Interface())), true
}

func roundTime(time int64, interval uint) int64 {
	return time - (time % int64(interval))
}
