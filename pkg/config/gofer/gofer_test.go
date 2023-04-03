package gofer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/marshal"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name string
		path string
		test func(*testing.T, *Config)
	}{
		{
			name: "agent",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				services, err := cfg.AgentServices(null.New())
				require.NoError(t, err)
				require.NotNil(t, services.Agent)
				require.NotNil(t, services.PriceProvider)
				require.NotNil(t, services.Logger)
			},
		},
		{
			name: "client",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				services, err := cfg.ClientServices(context.Background(), null.New(), true, marshal.Trace)
				require.NoError(t, err)
				require.NotNil(t, services.PriceProvider)
				require.NotNil(t, services.PriceHook)
				require.NotNil(t, services.Marshaller)
				require.NotNil(t, services.Logger)
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
