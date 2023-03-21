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
	"regexp"
	"runtime"
	"strings"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/chain"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/grafana"
)

type Dependencies struct {
	AppName    string
	BaseLogger log.Logger
}

type ConfigLogger struct {
	// Grafana is a configuration for a Grafana logger.
	Grafana *grafanaLogger `hcl:"grafana,block"`
}

type grafanaLogger struct {
	// Interval is a time interval in seconds between sending metrics to Grafana.
	Interval int `hcl:"interval"`

	// Endpoint is a Graphite endpoint.
	Endpoint string `hcl:"endpoint"`

	// APIKey is a Graphite API key.
	APIKey string `hcl:"api_key"`

	// Metrics is a list of metrics to send to Grafana.
	Metrics []grafanaMetric `hcl:"metric,block"`
}

type grafanaMetric struct {
	// MatchMessage is a regular expression to match a log message.
	MatchMessage string `hcl:"match_message"`

	// MatchFields is a map of regular expressions to match log fields.
	MatchFields map[string]string `hcl:"match_fields"`

	// Value is a dot-separated path of the field with the metric value.
	// If empty, the value 1 will be used as the metric value.
	Value string `hcl:"value,optional"`

	// ScaleFactor Scales the value by the specified number. If it is zero,
	// scaling is not applied.
	ScaleFactor float64 `hcl:"scale_factor,optional"`

	// Name of metric. It can contain references to log fields in the format
	// `%{path}`, where path is the dot-separated path to the field.
	Name string `hcl:"name"`

	// Tags is a list of metric tags. They can contain references to log fields
	// in the format `%{path}`, where path is the dot-separated path to the
	// field.
	Tags map[string][]string `hcl:"tags"`

	// OnDuplicate specifies how duplicated values in the same interval should
	// be handled. Possible values are:
	// - "sum" - sum the values
	// - "max" - use the maximum value
	// - "min" - use the minimum value
	// - "replace" - use the last value
	OnDuplicate string `hcl:"on_duplicate"`
}

func (c *ConfigLogger) Configure(d Dependencies) (log.Logger, error) {
	if c == nil {
		return d.BaseLogger, nil
	}

	loggers := []log.Logger{d.BaseLogger}

	if c.Grafana != nil {
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

func (c *ConfigLogger) configureGrafanaLogger(d Dependencies) (log.Logger, error) {
	var err error
	var ms []grafana.Metric
	for _, cm := range c.Grafana.Metrics {
		gm := grafana.Metric{
			Value:         cm.Value,
			Name:          cm.Name,
			Tags:          cm.Tags,
			TransformFunc: scalingFunc(cm.ScaleFactor),
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

	logger, err := grafana.New(d.BaseLogger.Level(), grafana.Config{
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
