package relay

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
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
				assert.Equal(t, uint32(60), cfg.Interval)

				assert.Equal(t, "client1", cfg.Median[0].EthereumClient)
				assert.Equal(t, "0x1234567890123456789012345678901234567890", cfg.Median[0].ContractAddr.String())
				assert.Equal(t, "BTCUSD", cfg.Median[0].Pair)
				assert.Equal(t, float64(1), cfg.Median[0].Spread)
				assert.Equal(t, uint32(300), cfg.Median[0].Expiration)

				assert.Equal(t, "client2", cfg.Median[1].EthereumClient)
				assert.Equal(t, "0x2345678901234567890123456789012345678901", cfg.Median[1].ContractAddr.String())
				assert.Equal(t, "ETHUSD", cfg.Median[1].Pair)
				assert.Equal(t, float64(3), cfg.Median[1].Spread)
				assert.Equal(t, uint32(400), cfg.Median[1].Expiration)
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
