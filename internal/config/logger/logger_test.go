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
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/grafana"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func TestLogger_Configure(t *testing.T) {
	prevGrafanaLoggerFactory := grafanaLoggerFactory
	defer func() { grafanaLoggerFactory = prevGrafanaLoggerFactory }()

	config := Logger{
		grafanaLogger{
			Enable:   true,
			Interval: 60,
			Endpoint: "https://example.com",
			APIKey:   "test",
			Metrics: []grafanaMetric{
				{
					MatchMessage: "message",
					MatchFields: map[string]string{
						"field": "value",
					},
					Value: "value",
					Name:  "name1",
					Tags: map[string][]string{
						"tag": {"a", "b"},
					},
					OnDuplicate: "sum",
				},
				{
					Name:        "name2",
					OnDuplicate: "SUB",
				},
				{
					Name:        "name3",
					OnDuplicate: "min",
				},
				{
					Name:        "name4",
					OnDuplicate: "max",
				},
				{
					Name:        "name5",
					OnDuplicate: "replace",
				},
				{
					Name:        "name6",
					OnDuplicate: "",
				},
				{
					Name:        "name7",
					ScaleFactor: 0,
				},
				{
					Name:        "name8",
					ScaleFactor: 1,
				},
				{
					Name:        "name9",
					ScaleFactor: 1e-2,
				},
				{
					Name:        "name10",
					ScaleFactor: 1e2,
				},
			},
		},
	}

	grafanaLoggerFactory = func(lvl log.Level, cfg grafana.Config) log.Logger {
		assert.Equal(t, log.Panic, lvl)
		assert.Equal(t, uint(60), cfg.Interval)
		assert.Equal(t, http.DefaultClient, cfg.HTTPClient)
		assert.Equal(t, "https://example.com", cfg.GraphiteEndpoint)
		assert.Equal(t, "test", cfg.GraphiteAPIKey)
		assert.Equal(t, "message", cfg.Metrics[0].MatchMessage.String())
		assert.Equal(t, "value", cfg.Metrics[0].MatchFields["field"].String())
		assert.Equal(t, "value", cfg.Metrics[0].Value)
		assert.Equal(t, []string{"a", "b"}, cfg.Metrics[0].Tags["tag"])
		assert.Equal(t, grafana.Sum, cfg.Metrics[0].OnDuplicate)
		assert.Equal(t, grafana.Sub, cfg.Metrics[1].OnDuplicate)
		assert.Equal(t, grafana.Min, cfg.Metrics[2].OnDuplicate)
		assert.Equal(t, grafana.Max, cfg.Metrics[3].OnDuplicate)
		assert.Equal(t, grafana.Replace, cfg.Metrics[4].OnDuplicate)
		assert.Equal(t, grafana.Replace, cfg.Metrics[5].OnDuplicate)

		assert.Nil(t, cfg.Metrics[6].TransformFunc)
		assert.Nil(t, cfg.Metrics[7].TransformFunc)

		assert.NotNil(t, cfg.Metrics[8].TransformFunc)
		assert.Equal(t, 0.01, cfg.Metrics[8].TransformFunc(1))

		assert.NotNil(t, cfg.Metrics[9].TransformFunc)
		assert.Equal(t, 100.0, cfg.Metrics[9].TransformFunc(1))

		return cfg.Logger
	}

	l, err := config.Configure(Dependencies{
		AppName:    "app",
		BaseLogger: null.New(),
	})

	assert.NotNil(t, l)
	assert.Equal(t, "*chain.logger", reflect.TypeOf(l).String())
	assert.NoError(t, err)
}
