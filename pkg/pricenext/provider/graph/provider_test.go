package graph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func newTestProvider() Provider {
	models := map[string]Node{
		"model_a": NewOriginNode(
			"test",
			provider.Pair{Base: "BTC", Quote: "USD"},
			provider.Pair{Base: "BTC", Quote: "USD"},
			time.Minute,
			time.Minute,
		),
		"model_b": NewOriginNode(
			"test",
			provider.Pair{Base: "BTC", Quote: "USD"},
			provider.Pair{Base: "BTC", Quote: "USD"},
			time.Minute,
			time.Minute,
		),
	}
	updater := NewUpdater(
		map[string]origin.Origin{
			"test": &mockOrigin{
				fetchTicks: func(_ context.Context, pairs []provider.Pair) []provider.Tick {
					ticks := make([]provider.Tick, len(pairs))
					for i, pair := range pairs {
						ticks[i] = provider.Tick{
							Pair:  pair,
							Price: bn.Float(42),
							Time:  time.Now(),
						}
					}
					return ticks
				},
			},
		},
		null.New(),
	)
	return NewProvider(models, updater)
}

func TestProvider_ModelNames(t *testing.T) {
	prov := newTestProvider()
	modelNames := prov.ModelNames(context.Background())

	// Model names must be returned in alphabetical order.
	assert.Equal(t, []string{"model_a", "model_b"}, modelNames)
}

func TestProvider_Tick(t *testing.T) {
	prov := newTestProvider()
	tick, err := prov.Tick(context.Background(), "model_a")
	require.NoError(t, err)

	// Price must be updated.
	assert.Equal(t, bn.Float(42), tick.Price)
}

func TestProvider_Ticks(t *testing.T) {
	prov := newTestProvider()
	tick, err := prov.Ticks(context.Background(), "model_a", "model_b")
	require.NoError(t, err)

	// Prices must be updated.
	assert.Equal(t, bn.Float(42), tick["model_a"].Price)
	assert.Equal(t, bn.Float(42), tick["model_b"].Price)
}

func TestProvider_Model(t *testing.T) {
	prov := newTestProvider()
	model, err := prov.Model(context.Background(), "model_a")
	require.NoError(t, err)

	assert.Equal(t, provider.Pair{Base: "BTC", Quote: "USD"}, model.Pair)
}

func TestProvider_Models(t *testing.T) {
	prov := newTestProvider()
	models, err := prov.Models(context.Background(), "model_a", "model_b")
	require.NoError(t, err)

	assert.Equal(t, provider.Pair{Base: "BTC", Quote: "USD"}, models["model_a"].Pair)
	assert.Equal(t, provider.Pair{Base: "BTC", Quote: "USD"}, models["model_b"].Pair)
}
