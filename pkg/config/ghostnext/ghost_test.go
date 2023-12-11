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
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		path string
		test func(*testing.T, *Config)
	}{
		{
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				services, err := cfg.Services(null.New(), "", "")
				require.NoError(t, err)
				require.NotNil(t, services)
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

func TestDefaults(t *testing.T) {
	cfg := &Config{}
	require.NoError(t, config.LoadEmbeds(cfg, cfg.DefaultEmbeds()))

	block := &hcl.Block{}
	hcl.Encode(cfg, block)

	expectedConfig, err := os.ReadFile("./testdata/default.hcl")
	require.NoError(t, err)

	loadedConfig, diags := block.Bytes()
	require.False(t, diags.HasErrors())

	require.Equal(t,
		strings.Trim(string(expectedConfig), "\n"),
		strings.Trim(string(loadedConfig), "\n"),
	)
}

func TestDefaultsForStage(t *testing.T) {
	require.NoError(t, os.Setenv("CFG_ENVIRONMENT", "stage"))
	require.NoError(t, os.Setenv("CFG_CHAIN_NAME", "sep"))

	cfg := &Config{}
	require.NoError(t, config.LoadEmbeds(cfg, cfg.DefaultEmbeds()))

	block := &hcl.Block{}
	hcl.Encode(cfg, block)

	expectedConfig, err := os.ReadFile("./testdata/default-stage-sep.hcl")
	require.NoError(t, err)

	loadedConfig, diags := block.Bytes()
	require.False(t, diags.HasErrors())

	require.Equal(t,
		strings.Trim(string(expectedConfig), "\n"),
		strings.Trim(string(loadedConfig), "\n"),
	)
}
