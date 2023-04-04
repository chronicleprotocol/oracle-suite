package ethereum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
				assert.Equal(t, []string{"rand_key"}, cfg.RandKeys)
				assert.Equal(t, "key1", cfg.Keys[0].Name)
				assert.Equal(t, "0xd18d7f6d9e349d1d6bf33702192019f166a7201e", cfg.Keys[0].Address.String())
				assert.Equal(t, "./testdata/keystore", cfg.Keys[0].KeystorePath)

				assert.Equal(t, "key2", cfg.Keys[1].Name)
				assert.Equal(t, "0x2d800d93b065ce011af83f316cef9f0d005b0aa4", cfg.Keys[1].Address.String())
				assert.Equal(t, "./testdata/keystore", cfg.Keys[1].KeystorePath)
				assert.Equal(t, "./testdata/keystore/passphrase", cfg.Keys[1].PassphraseFile)

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
		{
			name: "key registry",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				keys, diags := cfg.KeyRegistry(Dependencies{Logger: null.New()})
				require.NoError(t, diags)

				require.Len(t, keys, 3)
				assert.NotNil(t, keys["rand_key"])
				assert.Equal(t, "0xd18d7f6d9e349d1d6bf33702192019f166a7201e", keys["key1"].Address().String())
				assert.Equal(t, "0x2d800d93b065ce011af83f316cef9f0d005b0aa4", keys["key2"].Address().String())
			},
		},
		{
			name: "client registry",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				clients, diags := cfg.ClientRegistry(Dependencies{Logger: null.New()})
				require.NoError(t, diags)

				require.Len(t, clients, 2)
				assert.NotNil(t, clients["client1"])
				assert.NotNil(t, clients["client2"])
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
