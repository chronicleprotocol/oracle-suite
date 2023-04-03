package eventpublisher

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
				assert.Equal(t, "key", cfg.EthereumKey)

				assert.Equal(t, "client1", cfg.TeleportEVM[0].EthereumClient)
				assert.Equal(t, uint32(60), cfg.TeleportEVM[0].Interval)
				assert.Equal(t, uint64(120), cfg.TeleportEVM[0].PrefetchPeriod)
				assert.Equal(t, uint64(3), cfg.TeleportEVM[0].BlockConfirmations)
				assert.Equal(t, uint64(100), cfg.TeleportEVM[0].BlockLimit)
				assert.Equal(t, []uint64{600, 1200}, cfg.TeleportEVM[0].ReplayAfter)
				assert.Equal(t, "0x1234567890123456789012345678901234567890", cfg.TeleportEVM[0].ContractAddrs[0].String())
				assert.Equal(t, "0x2345678901234567890123456789012345678901", cfg.TeleportEVM[0].ContractAddrs[1].String())

				assert.Equal(t, "http://localhost:8080", cfg.TeleportStarknet[0].Sequencer.String())
				assert.Equal(t, uint32(60), cfg.TeleportStarknet[0].Interval)
				assert.Equal(t, uint32(120), cfg.TeleportStarknet[0].PrefetchPeriod)
				assert.Equal(t, []uint32{600, 1200}, cfg.TeleportStarknet[0].ReplayAfter)
				assert.Equal(t, "3456789012345678901234567890123456789012", cfg.TeleportStarknet[0].ContractAddrs[0].Text(16))
				assert.Equal(t, "4567890123456789012345678901234567890123", cfg.TeleportStarknet[0].ContractAddrs[1].Text(16))
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
