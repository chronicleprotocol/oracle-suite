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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name string
		path string
		test func(*testing.T, *Config)
	}{
		{
			name: "valid",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				assert.NotNil(t, cfg.Grafana)
				assert.Equal(t, 60, cfg.Grafana.Interval)
				assert.Equal(t, "https://graphite.example.com", cfg.Grafana.Endpoint.String())
				assert.Equal(t, "your_api_key", cfg.Grafana.APIKey)

				require.Len(t, cfg.Grafana.Metrics, 1)
				metric := cfg.Grafana.Metrics[0]
				assert.Equal(t, "message", metric.MatchMessage)
				assert.Equal(t, map[string]string{"type": "sell"}, metric.MatchFields)
				assert.Equal(t, "message.path", metric.Value)
				assert.Equal(t, 0.5, metric.ScaleFactor)
				assert.Equal(t, "example.message", metric.Name)
				assert.Equal(t, map[string][]string{"environment": {"production"}}, metric.Tags)
				assert.Equal(t, "sum", metric.OnDuplicate)
			},
		},
		{
			name: "service",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				service, err := cfg.Logger(Dependencies{
					AppName:    "app",
					AppVersion: suite.Version,
					BaseLogger: null.New(),
				})
				require.NoError(t, err)
				assert.NotNil(t, service)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var cfg Config
			err := config.LoadFiles(&cfg, []string{"./testdata/" + test.path})
			require.NoError(t, err)
			test.test(t, &cfg)
		})
	}
}
