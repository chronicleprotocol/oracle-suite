package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		path            string
		asserts         func(*testing.T, *Config)
		wantErrContains string
	}{
		{
			path: "config.hcl",
			asserts: func(t *testing.T, cfg *Config) {
				assert.NotNil(t, cfg.LibP2P)
				assert.NotNil(t, cfg.WebAPI)

				// LibP2P
				assert.Equal(t, "0x1234567890123456789012345678901234567890", cfg.LibP2P.Feeds[0].String())
				assert.Equal(t, "0x2345678901234567890123456789012345678901", cfg.LibP2P.Feeds[1].String())
				assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/6000"}, cfg.LibP2P.ListenAddrs)
				assert.Equal(t, "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2", cfg.LibP2P.PrivKeySeed)
				assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/7000"}, cfg.LibP2P.BootstrapAddrs)
				assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/8000"}, cfg.LibP2P.DirectPeersAddrs)
				assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/9000"}, cfg.LibP2P.BlockedAddrs)
				assert.Equal(t, true, cfg.LibP2P.DisableDiscovery)
				assert.Equal(t, "eth_key", cfg.LibP2P.EthereumKey)

				// WebAPI
				assert.Equal(t, "0x3456789012345678901234567890123456789012", cfg.WebAPI.Feeds[0].String())
				assert.Equal(t, "0x4567890123456789012345678901234567890123", cfg.WebAPI.Feeds[1].String())
				assert.Equal(t, "localhost:8080", cfg.WebAPI.ListenAddr)
				assert.Equal(t, "localhost:9050", cfg.WebAPI.Socks5ProxyAddr)
				assert.Equal(t, "eth_key", cfg.WebAPI.EthereumKey)
				assert.NotNil(t, cfg.WebAPI.EthereumAddressBook)
				assert.NotNil(t, cfg.WebAPI.StaticAddressBook)

				// EthereumAddressBook
				assert.Equal(t, "0x5678901234567890123456789012345678901234", cfg.WebAPI.EthereumAddressBook.ContractAddr.String())
				assert.Equal(t, "default", cfg.WebAPI.EthereumAddressBook.EthereumClient)

				// StaticAddressBook
				assert.Equal(t, []string{"https://example.com/api/v1/endpoint"}, cfg.WebAPI.StaticAddressBook.Addresses)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			var cfg Config
			err := config.LoadFiles(&cfg, []string{"./testdata/" + test.path})
			if test.wantErrContains != "" {
				require.Containsf(t, err.Error(), test.wantErrContains, "unexpected error: %v", err)
				return
			}
			require.NoError(t, err)
			test.asserts(t, &cfg)
		})
	}
}
