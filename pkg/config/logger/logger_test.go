package logger

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
				assert.NotNil(t, cfg.Grafana)
				assert.Equal(t, 60, cfg.Grafana.Interval)
				assert.Equal(t, "https://graphite.example.com", cfg.Grafana.Endpoint.String())
				assert.Equal(t, "your_api_key", cfg.Grafana.APIKey)

				require.Len(t, cfg.Grafana.Metrics, 1)
				metric := cfg.Grafana.Metrics[0]
				assert.Equal(t, "message", metric.MatchMessage)
				assert.Equal(t, map[string]string{"type": "sell"}, metric.MatchFields)
				assert.Equal(t, "message.path", metric.Value)
				assert.Equal(t, 0.5, metric.ScaleFactor)
				assert.Equal(t, "example.message", metric.Name)
				assert.Equal(t, map[string][]string{"environment": {"production"}}, metric.Tags)
				assert.Equal(t, "sum", metric.OnDuplicate)
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
