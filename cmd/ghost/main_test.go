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

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	ghost "github.com/chronicleprotocol/oracle-suite/pkg/config/ghostnext"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

func TestConfig_Ghost_Run(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		config   supervisor.Config
		services supervisor.Service
		envVars  map[string]string
		wantErr  bool
	}{
		{
			name:     "ghost-run-test",
			args:     []string{},
			config:   &ghost.Config{},
			services: &ghost.Services{},
			envVars: map[string]string{
				"CFG_LIBP2P_EXTERNAL_ADDR": "1.2.3.4",
			},
		},
	}

	for _, tt := range tests {
		os.Clearenv()
		for k, v := range tt.envVars {
			require.NoError(t, os.Setenv(k, v))
		}

		t.Run(tt.name, func(t *testing.T) {
			var cf = cmd.ConfigFlagsForConfig(tt.config.(config.HasDefaults))
			require.NoError(t, cf.FlagSet().Parse(tt.args))
			err := cf.Load(tt.config)
			require.NoError(t, err)

			s, err := tt.config.Services(null.New(), tt.name, suite.Version)
			require.NoError(t, err)

			assert.IsType(t, tt.services, s)
		})
	}
}
