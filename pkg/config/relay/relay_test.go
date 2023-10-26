package relay

import (
	"testing"

	"github.com/defiweb/go-eth/types"
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
				assert.Equal(t, "client1", cfg.Median[0].EthereumClient)
				assert.Equal(t, "0x1234567890123456789012345678901234567890", cfg.Median[0].ContractAddr.String())
				assert.Equal(t, "ETH/USD", cfg.Median[0].DataModel)
				assert.Equal(t, float64(1), cfg.Median[0].Spread)
				assert.Equal(t, uint32(300), cfg.Median[0].Expiration)
				assert.Equal(t, []types.Address{
					types.MustAddressFromHex("0x0011223344556677889900112233445566778899"),
					types.MustAddressFromHex("0x1122334455667788990011223344556677889900"),
				}, cfg.Median[0].Feeds)

				assert.Equal(t, "client2", cfg.Scribe[0].EthereumClient)
				assert.Equal(t, "0x2345678901234567890123456789012345678901", cfg.Scribe[0].ContractAddr.String())
				assert.Equal(t, "BTC/USD", cfg.Scribe[0].DataModel)
				assert.Equal(t, float64(2), cfg.Scribe[0].Spread)
				assert.Equal(t, uint32(400), cfg.Scribe[0].Expiration)
				assert.Equal(t, []types.Address{
					types.MustAddressFromHex("0x2233445566778899001122334455667788990011"),
					types.MustAddressFromHex("0x3344556677889900112233445566778899001122"),
				}, cfg.Scribe[0].Feeds)

				assert.Equal(t, "client3", cfg.OptimisticScribe[0].EthereumClient)
				assert.Equal(t, "0x3456789012345678901234567890123456789012", cfg.OptimisticScribe[0].ContractAddr.String())
				assert.Equal(t, "MKR/USD", cfg.OptimisticScribe[0].DataModel)
				assert.Equal(t, float64(3), cfg.OptimisticScribe[0].Spread)
				assert.Equal(t, uint32(500), cfg.OptimisticScribe[0].Expiration)
				assert.Equal(t, float64(4), cfg.OptimisticScribe[0].OptimisticSpread)
				assert.Equal(t, uint32(600), cfg.OptimisticScribe[0].OptimisticExpiration)
				assert.Equal(t, []types.Address{
					types.MustAddressFromHex("0x4455667788990011223344556677889900112233"),
					types.MustAddressFromHex("0x5566778899001122334455667788990011223344"),
				}, cfg.OptimisticScribe[0].Feeds)
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
