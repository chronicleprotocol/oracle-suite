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
	"context"
	"fmt"
	"net/http"
	"regexp"
	"runtime"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/chain"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/grafana"
)

var grafanaLoggerFactory = func(lvl log.Level, cfg grafana.Config) log.Logger {
	return grafana.New(context.Background(), lvl, cfg)
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
	MatchMessage string              `json:"matchMessage"`
	MatchFields  map[string]string   `json:"matchFields"`
	Value        string              `json:"value"`
	Name         string              `json:"name"`
	Tags         map[string][]string `json:"tags"`
}

func (c *Logger) Configure(d Dependencies) (log.Logger, error) {
	loggers := []log.Logger{d.BaseLogger}

	if c.Grafana.Enable {
		var m []grafana.Metric
		for _, cm := range c.Grafana.Metrics {
			// Compile a regular expression for a message:
			mrx, err := regexp.Compile(cm.MatchMessage)
			if err != nil {
				return nil, fmt.Errorf("logger config: unable to compile regexp: %s", cm.MatchMessage)
			}
			// Compile regular expressions for log fields:
			frx := map[string]*regexp.Regexp{}
			for f, p := range cm.MatchFields {
				rx, err := regexp.Compile(p)
				if err != nil {
					return nil, fmt.Errorf("logger config: unable to compile regexp: %s", p)
				}
				frx[f] = rx
			}
			m = append(m, grafana.Metric{
				MatchMessage: mrx,
				MatchFields:  frx,
				Value:        cm.Value,
				Name:         cm.Name,
				Tags:         cm.Tags,
			})
		}
		interval := c.Grafana.Interval
		if interval < 1 {
			interval = 1
		}
		loggers = append(loggers, grafanaLoggerFactory(d.BaseLogger.Level(), grafana.Config{
			Metrics:          m,
			Interval:         uint(interval),
			GraphiteEndpoint: c.Grafana.Endpoint,
			GraphiteAPIKey:   c.Grafana.APIKey,
			HTTPClient:       http.DefaultClient,
			Logger:           d.BaseLogger,
		}))
	}

	var logger log.Logger
	switch len(loggers) {
	case 1:
		logger = loggers[0]
	default:
		logger = chain.New(loggers...)
	}

	logger = logger.
		WithFields(log.Fields{
			"x-appName":    d.AppName,
			"x-appVersion": suite.Version,
			"x-goVersion":  runtime.Version(),
			"x-goOS":       runtime.GOOS,
			"x-goArch":     runtime.GOARCH,
		})

	return logger, nil
}
