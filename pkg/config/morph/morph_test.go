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

package morph

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name string
		path string
		test func(t2 *testing.T, config *Config)
	}{
		{
			name: "valid",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "config-cache.hcl", cfg.MorphFile)
				assert.Equal(t, "default", cfg.EthereumClient)
				assert.Equal(t, "0x1234567890123456789012345678901234567890", cfg.ConfigRegistryAddress.String())
				assert.Equal(t, uint32(3600), cfg.Interval)
			},
		},
		{
			name: "service",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				clients := ethereum.ClientRegistry{
					"default": new(ethereumMocks.RPC),
				}
				var base Config
				morph, err := cfg.ConfigureMorph(Dependencies{
					Clients: clients,
					Base:    reflect.ValueOf(base),
					Logger:  null.New(),
				})
				require.NoError(t, err)
				assert.NotNil(t, morph)
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
