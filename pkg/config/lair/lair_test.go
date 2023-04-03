package lair

import (
	"testing"

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
				services, err := cfg.Services(null.New())
				require.NoError(t, err)
				require.NotNil(t, services.Transport)
				require.NotNil(t, services.EventStore)
				require.NotNil(t, services.EventAPI)
				require.NotNil(t, services.Logger)
			},
		},
		{
			name: "multiple storages",
			path: "multiple-storages.hcl",
			test: func(t *testing.T, cfg *Config) {
				_, err := cfg.Services(null.New())
				require.Error(t, err)
				require.Contains(t, err.Error(), `multiple-storages.hcl:1,1-5: Validation error; "storage_memory" and "storage_redis" storage types are mutually exclusive`)
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
