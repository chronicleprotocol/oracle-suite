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
	"net/http"
	"regexp"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/chain"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/grafana"
)

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
	Pattern string              `json:"pattern"`
	Value   string              `json:"value"`
	Name    string              `json:"name"`
	Tags    map[string][]string `json:"tags"`
}

func (c *Logger) Configure(d Dependencies) (log.Logger, error) {
	logger := d.BaseLogger.WithField("x-appName", d.AppName)
	if c.Grafana.Enable {
		var m []grafana.Metric
		for _, cm := range c.Grafana.Metrics {
			p, err := regexp.Compile(cm.Pattern)
			if err != nil {
				return nil, fmt.Errorf("logger config: unable to compile regexp: %s", cm.Pattern)
			}
			m = append(m, grafana.Metric{
				Pattern: p,
				Value:   cm.Value,
				Name:    cm.Name,
				Tags:    cm.Tags,
			})
		}
		interval := c.Grafana.Interval
		if interval < 1 {
			interval = 1
		}
		logger = chain.New(logger, grafana.New(logger.Level(), grafana.Config{
			Metrics:          m,
			Interval:         uint(interval),
			GraphiteEndpoint: c.Grafana.Endpoint,
			GraphiteAPIKey:   c.Grafana.APIKey,
			HTTPClient:       http.DefaultClient,
			Logger:           logger,
		}))
	}
	return logger, nil
}
