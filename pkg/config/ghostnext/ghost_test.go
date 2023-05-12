package ghost

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		path string
		test func(*testing.T, *Config)
	}{
		{
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				services, err := cfg.Services(null.New())
				require.NoError(t, err)
				require.NotNil(t, services.Feed)
				require.NotNil(t, services.Transport)
				require.NotNil(t, services.Logger)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			var cfg Config
			err := config.LoadFiles(&cfg, []string{"./testdata/" + test.path})
			require.NoError(t, err)
			test.test(t, &cfg)
		})
	}
}
