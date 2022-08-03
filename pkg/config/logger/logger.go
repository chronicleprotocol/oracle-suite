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

package logger

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/chain"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/grafana"
)

var grafanaLoggerFactory = func(lvl log.Level, cfg grafana.Config) (log.Logger, error) {
	return grafana.New(lvl, cfg)
}

type Dependencies struct {
	AppName    string
	BaseLogger log.Logger
}

type Logger struct {
	Grafana grafanaLogger `json:"grafana"`
}

type grafanaLogger struct {
	Enable   bool            `json:"enable"`
	Interval int             `json:"interval"`
	Endpoint string          `json:"endpoint"`
	APIKey   string          `json:"apiKey"`
	Metrics  []grafanaMetric `json:"metrics"`
}

type grafanaMetric struct {
	Type         string              `json:"type"`
	MatchMessage string              `json:"matchMessage"`
	MatchFields  map[string]string   `json:"matchFields"`
	Value        string              `json:"value"`
	ScaleFactor  float64             `json:"scaleFactor"`
	Name         string              `json:"name"`
	Tags         map[string][]string `json:"tags"`
	OnDuplicate  string              `json:"onDuplicate"`
}

func (c *Logger) Configure(d Dependencies) (log.Logger, error) {
	loggers := []log.Logger{d.BaseLogger}

	if c.Grafana.Enable {
		logger, err := c.configureGrafanaLogger(d)
		if err != nil {
			return nil, fmt.Errorf("logger config: unable to create grafana logger: %s", err)
		}
		loggers = append(loggers, logger)
	}

	logger := chain.New(loggers...)
	if len(loggers) == 1 {
		logger = loggers[0]
	}

	logger = logger.
		WithFields(log.Fields{
			"x-appName":    d.AppName,
			"x-appVersion": suite.Version,
			"x-goVersion":  runtime.Version(),
			"x-goOS":       runtime.GOOS,
			"x-goArch":     runtime.GOARCH,
			"x-instanceID": fmt.Sprintf("%08x", rand.Uint64()), //nolint:gosec
		})

	return logger, nil
}

func (c *Logger) configureGrafanaLogger(d Dependencies) (log.Logger, error) {
	var err error
	var ms []grafana.Metric
	for _, cm := range c.Grafana.Metrics {
		gm := grafana.Metric{
			Value:         cm.Value,
			Name:          cm.Name,
			Tags:          cm.Tags,
			TransformFunc: scalingFunc(cm.ScaleFactor),
		}
		if cm.Type == "semver" {
			gm.ParserFunc = ToFloatVersion
		}

		// Compile the regular expression for a message:
		gm.MatchMessage, err = regexp.Compile(cm.MatchMessage)
		if err != nil {
			return nil, fmt.Errorf("unable to compile regexp: %s", cm.MatchMessage)
		}

		// Compile regular expressions for log fields:
		gm.MatchFields = map[string]*regexp.Regexp{}
		for f, p := range cm.MatchFields {
			rx, err := regexp.Compile(p)
			if err != nil {
				return nil, fmt.Errorf("unable to compile regexp: %s", p)
			}
			gm.MatchFields[f] = rx
		}

		// On duplicate:
		switch strings.ToLower(cm.OnDuplicate) {
		case "sum":
			gm.OnDuplicate = grafana.Sum
		case "sub":
			gm.OnDuplicate = grafana.Sub
		case "min":
			gm.OnDuplicate = grafana.Min
		case "max":
			gm.OnDuplicate = grafana.Max
		case "replace", "":
			gm.OnDuplicate = grafana.Replace
		default:
			return nil, fmt.Errorf("unknown onDuplicate value: %s", cm.OnDuplicate)
		}

		ms = append(ms, gm)
	}

	interval := c.Grafana.Interval
	if interval < 1 {
		interval = 1
	}

	logger, err := grafanaLoggerFactory(d.BaseLogger.Level(), grafana.Config{
		Metrics:          ms,
		Interval:         uint(interval),
		GraphiteEndpoint: c.Grafana.Endpoint,
		GraphiteAPIKey:   c.Grafana.APIKey,
		HTTPClient:       http.DefaultClient,
		Logger:           d.BaseLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create grafana logger: %s", err)
	}

	return logger, nil
}

func scalingFunc(sf float64) func(v float64) float64 {
	if sf == 0 || sf == 1 {
		return nil
	}
	return func(v float64) float64 { return v * sf }
}

func ToFloatVersion(value reflect.Value) (float64, bool) {
	if !value.IsValid() {
		return 0, false
	}
	if value.Type().Kind() == reflect.String {
		r := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(.*)$`)
		ret := r.FindStringSubmatch(value.String())
		if ret == nil {
			return 0, false
		}

		var vMajor, vMinor, vPatch float64
		var err error
		if ret[1] != "" {
			vMajor, err = strconv.ParseFloat(ret[1], 64)
			if err != nil {
				return 0, false
			}
		} else {
			vMajor = 0
		}
		if ret[2] != "" {
			vMinor, err = strconv.ParseFloat(ret[2], 64)
			if err != nil {
				return 0, false
			}
		} else {
			vMinor = 0
		}
		if ret[3] != "" {
			vPatch, err = strconv.ParseFloat(ret[3], 64)
			if err != nil {
				return 0, false
			}
		} else {
			vPatch = 0
		}
		v := vMajor*1e6 + vMinor*1e3 + vPatch
		if len(ret[4]) > 0 {
			return -v, err == nil
		}
		return v, err == nil
	}
	return 0, false
}
