package eventapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportevm"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportstarknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
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
				assert.Equal(t, "0.0.0.0:8000", cfg.ListenAddr)

				assert.NotNil(t, cfg.Memory)
				assert.Equal(t, uint32(86400), cfg.Memory.TTL)

				assert.NotNil(t, cfg.Redis)
				assert.Equal(t, uint32(86400), cfg.Redis.TTL)
				assert.Equal(t, "localhost:6379", cfg.Redis.Address)
				assert.Equal(t, "user", cfg.Redis.Username)
				assert.Equal(t, "password", cfg.Redis.Password)
				assert.Equal(t, 0, cfg.Redis.DB)
				assert.Equal(t, int64(1048576), cfg.Redis.MemoryLimit)
				assert.Equal(t, false, cfg.Redis.TLS)
				assert.Equal(t, "localhost", cfg.Redis.TLSServerName)
				assert.Equal(t, "./tls_cert.pem", cfg.Redis.TLSCertFile)
				assert.Equal(t, "./tls_key.pem", cfg.Redis.TLSKeyFile)
				assert.Equal(t, "./tls_root_ca.pem", cfg.Redis.TLSRootCAFile)
				assert.Equal(t, false, cfg.Redis.Cluster)
				assert.Equal(t, []string{"localhost:7000", "localhost:7001"}, cfg.Redis.ClusterAddrs)
			},
		},
		{
			name: "service",
			path: "service.hcl",
			test: func(t *testing.T, cfg *Config) {
				transport := local.New([]byte("test"), 1, nil)
				storage, err := cfg.Storage()
				require.NoError(t, err)
				logger := null.New()

				eventStore, err := store.New(store.Config{
					EventTypes: []string{teleportevm.TeleportEventType, teleportstarknet.TeleportEventType},
					Storage:    storage,
					Transport:  transport,
					Logger:     logger,
				})
				require.NoError(t, err)
				require.NotNil(t, eventStore)

				eventAPI, err := cfg.EventAPI(Dependencies{
					EventStore: eventStore,
					Transport:  transport,
					Logger:     logger,
				})
				require.NoError(t, err)
				assert.NotNil(t, eventAPI)
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
