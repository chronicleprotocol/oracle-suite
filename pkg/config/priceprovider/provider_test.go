package priceprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
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
				// Check RPC server configuration.
				assert.Equal(t, "localhost:8080", cfg.RPCListenAddr)
				assert.Equal(t, "localhost:8081", cfg.RPCAgentAddr)

				// Check origins configuration.
				require.Len(t, cfg.Origins, 1)
				assert.Equal(t, "origin", cfg.Origins[0].Origin)
				assert.Equal(t, "origin", cfg.Origins[0].Type)
				assert.Equal(t, map[string]any{
					"contract_address": "0x1234567890123456789012345678901234567890",
				}, cfg.Origins[0].Params)

				// Check price models configuration.
				require.Len(t, cfg.PriceModels, 1)
				assert.Equal(t, provider.Pair{"AAA", "BBB"}, cfg.PriceModels[0].Pair)
				assert.Equal(t, "median", cfg.PriceModels[0].Type)

				// Check price model sources configuration.
				require.Len(t, cfg.PriceModels[0].Sources, 2)
				assert.Equal(t, 3, cfg.PriceModels[0].Median.MinSources)
				assert.Equal(t, provider.Pair{"AAA", "BBB"}, cfg.PriceModels[0].Sources[0].Pair)
				assert.Equal(t, "origin", cfg.PriceModels[0].Sources[0].Type)
				assert.Equal(t, "origin1", cfg.PriceModels[0].Sources[0].Origin.Origin)

				assert.Equal(t, provider.Pair{"AAA", "BBB"}, cfg.PriceModels[0].Sources[1].Pair)
				assert.Equal(t, "indirect", cfg.PriceModels[0].Sources[1].Type)

				// Check indirect sources configuration.
				require.Len(t, cfg.PriceModels[0].Sources[1].Sources, 2)
				assert.Equal(t, provider.Pair{"AAA", "XXX"}, cfg.PriceModels[0].Sources[1].Sources[0].Pair)
				assert.Equal(t, "origin", cfg.PriceModels[0].Sources[1].Sources[0].Type)
				assert.Equal(t, "origin2", cfg.PriceModels[0].Sources[1].Sources[0].Origin.Origin)

				assert.Equal(t, provider.Pair{"XXX", "BBB"}, cfg.PriceModels[0].Sources[1].Sources[1].Pair)
				assert.Equal(t, "origin", cfg.PriceModels[0].Sources[1].Sources[1].Type)
				assert.Equal(t, "origin3", cfg.PriceModels[0].Sources[1].Sources[1].Origin.Origin)
			},
		},
		{
			name: "service",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				clientRegistry := ethereum.ClientRegistry{
					"client": &mocks.RPC{},
				}
				priceProvider, err := cfg.PriceProvider(Dependencies{
					Clients: clientRegistry,
					Logger:  null.New(),
				}, false)
				require.NoError(t, err)
				assert.NotNil(t, priceProvider)
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
