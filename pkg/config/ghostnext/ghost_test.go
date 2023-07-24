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

package ghostnext

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		path string
		test func(*testing.T, *Config)
	}{
		{
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				services, err := cfg.Services(null.New(), false)
				require.NoError(t, err)
				require.NotNil(t, services.Feed)
				require.NotNil(t, services.Transport)
				require.NotNil(t, services.Logger)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			var cfg Config
			err := config.LoadFiles(&cfg, []string{"./testdata/" + test.path})
			require.NoError(t, err)
			test.test(t, &cfg)
		})
	}
}
