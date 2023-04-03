package ethereum

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
				assert.Equal(t, []string{"key"}, cfg.RandKeys)
				assert.Equal(t, "key1", cfg.Keys[0].Name)
				assert.Equal(t, "0x1234567890123456789012345678901234567890", cfg.Keys[0].Address.String())
				assert.Equal(t, "./keystore", cfg.Keys[0].KeystorePath)

				assert.Equal(t, "key2", cfg.Keys[1].Name)
				assert.Equal(t, "0x2345678901234567890123456789012345678901", cfg.Keys[1].Address.String())
				assert.Equal(t, "./keystore2", cfg.Keys[1].KeystorePath)
				assert.Equal(t, "./passphrase", cfg.Keys[1].PassphraseFile)

				assert.Equal(t, "client1", cfg.Clients[0].Name)
				assert.Equal(t, "https://rpc1.example", cfg.Clients[0].RPCURLs[0].String())
				assert.Equal(t, uint64(1), cfg.Clients[0].ChainID)
				assert.Equal(t, "key1", cfg.Clients[0].EthereumKey)

				assert.Equal(t, "client2", cfg.Clients[1].Name)
				assert.Equal(t, "https://rpc2.example", cfg.Clients[1].RPCURLs[0].String())
				assert.Equal(t, uint32(10), cfg.Clients[1].Timeout)
				assert.Equal(t, uint32(5), cfg.Clients[1].GracefulTimeout)
				assert.Equal(t, uint64(100), cfg.Clients[1].MaxBlocksBehind)
				assert.Equal(t, "key2", cfg.Clients[1].EthereumKey)
				assert.Equal(t, uint64(1), cfg.Clients[1].ChainID)
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
