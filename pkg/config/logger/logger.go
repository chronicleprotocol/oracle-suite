//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/chain"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/grafana"
)

type Dependencies struct {
	AppName    string
	BaseLogger log.Logger
}

type Config struct {
	// Grafana is a configuration for a Grafana logger.
	Grafana *grafanaLogger `hcl:"grafana,block,optional"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured service:
	logger log.Logger
}

type grafanaLogger struct {
	// Interval is a time interval in seconds between sending metrics to Grafana.
	Interval int `hcl:"interval"`

	// Endpoint is a Graphite endpoint.
	Endpoint config.URL `hcl:"endpoint"`

	// APIKey is a Graphite API key.
	APIKey string `hcl:"api_key"`

	// Metrics is a list of metrics to send to Grafana.
	Metrics []grafanaMetric `hcl:"metric,block"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
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

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

func (c *Config) Logger(d Dependencies) (log.Logger, error) {
	if c == nil {
		return d.BaseLogger, nil
	}
	if c.logger != nil {
		return c.logger, nil
	}
	loggers := []log.Logger{d.BaseLogger}
	if c.Grafana != nil {
		logger, err := c.grafanaLogger(d)
		if err != nil {
			return nil, err
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
			"x-appVersion": cmd.Version,
			"x-goVersion":  runtime.Version(),
			"x-goOS":       runtime.GOOS,
			"x-goArch":     runtime.GOARCH,
			"x-instanceID": fmt.Sprintf("%08x", rand.Uint64()), //nolint:gosec
		})
	c.logger = logger
	return logger, nil
}

func (c *Config) grafanaLogger(d Dependencies) (log.Logger, error) {
	var err error
	var metrics []grafana.Metric
	for _, cfg := range c.Grafana.Metrics {
		metric := grafana.Metric{
			Value:         cfg.Value,
			Name:          cfg.Name,
			Tags:          cfg.Tags,
			TransformFunc: scalingFunc(cfg.ScaleFactor),
		}

		// Compile the regular expression for a message:
		metric.MatchMessage, err = regexp.Compile(cfg.MatchMessage)
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid regular expression: %s", cfg.MatchMessage),
				Subject:  cfg.Content.Attributes["match_message"].NameRange.Ptr(),
			}
		}

		// Compile regular expressions for log fields:
		metric.MatchFields = map[string]*regexp.Regexp{}
		for f, p := range cfg.MatchFields {
			rx, err := regexp.Compile(p)
			if err != nil {
				return nil, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Validation error",
					Detail:   fmt.Sprintf("Invalid regular expression: %s", p),
					Subject:  cfg.Content.Attributes["match_fields"].NameRange.Ptr(),
				}
			}
			metric.MatchFields[f] = rx
		}

		// On duplicate:
		switch strings.ToLower(cfg.OnDuplicate) {
		case "sum":
			metric.OnDuplicate = grafana.Sum
		case "min":
			metric.OnDuplicate = grafana.Min
		case "max":
			metric.OnDuplicate = grafana.Max
		case "replace", "":
			metric.OnDuplicate = grafana.Replace
		default:
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail: fmt.Sprintf(
					"Invalid value for on_duplicate: %s. Possible values are: sum, min, max, replace",
					cfg.OnDuplicate,
				),
				Subject: cfg.Content.Attributes["on_duplicate"].NameRange.Ptr(),
			}
		}

		metrics = append(metrics, metric)
	}
	interval := c.Grafana.Interval
	if interval < 1 {
		interval = 1
	}
	logger, err := grafana.New(d.BaseLogger.Level(), grafana.Config{
		Metrics:          metrics,
		Interval:         uint(interval),
		GraphiteEndpoint: c.Grafana.Endpoint.String(),
		GraphiteAPIKey:   c.Grafana.APIKey,
		HTTPClient:       http.DefaultClient,
		Logger:           d.BaseLogger,
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Grafana logger: %s", err),
			Subject:  c.Range.Ptr(),
		}
	}
	return logger, nil
}

func scalingFunc(sf float64) func(v float64) float64 {
	if sf == 0 || sf == 1 {
		return nil
	}
	return func(v float64) float64 { return v * sf }
}
