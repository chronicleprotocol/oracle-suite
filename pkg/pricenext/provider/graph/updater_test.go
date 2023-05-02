package graph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/callback"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestUpdater(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		g := []Node{
			NewOriginNode(
				"origin_a",
				provider.Pair{Base: "BTC", Quote: "USD"},
				provider.Pair{Base: "BTC", Quote: "USD"},
				time.Minute,
				time.Minute,
			),
			NewOriginNode(
				"origin_b",
				provider.Pair{Base: "BTC", Quote: "USD"},
				provider.Pair{Base: "BTC", Quote: "USD"},
				time.Minute,
				time.Minute,
			),
		}
		u := NewUpdater(
			map[string]origin.Origin{
				"origin_a": &mockOrigin{
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
				"origin_b": &mockOrigin{
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
		require.NoError(t, u.Update(context.Background(), g))
		assert.Equal(t, bn.Float(42), g[0].Tick().Price)
		assert.Equal(t, bn.Float(42), g[1].Tick().Price)
	})
	t.Run("fresh tick", func(t *testing.T) {
		g := []Node{
			NewOriginNode(
				"origin_a",
				provider.Pair{Base: "BTC", Quote: "USD"},
				provider.Pair{Base: "BTC", Quote: "USD"},
				time.Minute,
				time.Minute,
			),
		}
		g[0].(*OriginNode).SetTick(provider.Tick{
			Pair:  provider.Pair{Base: "BTC", Quote: "USD"},
			Price: bn.Float(42),
			Time:  time.Now(),
		})

		u := NewUpdater(
			map[string]origin.Origin{
				"origin_a": &mockOrigin{
					fetchTicks: func(_ context.Context, pairs []provider.Pair) []provider.Tick {
						ticks := make([]provider.Tick, len(pairs))
						for i, pair := range pairs {
							ticks[i] = provider.Tick{
								Pair:  pair,
								Price: bn.Float(3.14),
								Time:  time.Now(),
							}
						}
						return ticks
					},
				},
			},
			null.New(),
		)
		require.NoError(t, u.Update(context.Background(), g))
		assert.Equal(t, bn.Float(42), g[0].Tick().Price) // tick should not be updated with a 3.14 price
	})
	t.Run("missing tick", func(t *testing.T) {
		g := []Node{
			NewOriginNode(
				"origin_a",
				provider.Pair{Base: "BTC", Quote: "USD"},
				provider.Pair{Base: "BTC", Quote: "USD"},
				time.Minute,
				time.Minute,
			),
			NewOriginNode(
				"origin_b",
				provider.Pair{Base: "BTC", Quote: "USD"},
				provider.Pair{Base: "BTC", Quote: "USD"},
				time.Minute,
				time.Minute,
			),
		}
		u := NewUpdater(
			map[string]origin.Origin{
				"origin_a": &mockOrigin{
					fetchTicks: func(_ context.Context, pairs []provider.Pair) []provider.Tick {
						return nil
					},
				},
				"origin_b": &mockOrigin{
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
		require.NoError(t, u.Update(context.Background(), g))
		assert.Error(t, g[0].Tick().Validate())
		assert.Contains(t, g[0].Tick().Validate().Error(), "tick is not set")
		assert.Equal(t, bn.Float(42), g[1].Tick().Price)
	})
	t.Run("panic", func(t *testing.T) {
		var logs []string
		l := callback.New(log.Debug, func(level log.Level, fields log.Fields, log string) {
			logs = append(logs, log)
		})
		g := []Node{
			NewOriginNode(
				"origin_a",
				provider.Pair{Base: "BTC", Quote: "USD"},
				provider.Pair{Base: "BTC", Quote: "USD"},
				time.Minute,
				time.Minute,
			),
			NewOriginNode(
				"origin_b",
				provider.Pair{Base: "BTC", Quote: "USD"},
				provider.Pair{Base: "BTC", Quote: "USD"},
				time.Minute,
				time.Minute,
			),
		}
		u := NewUpdater(
			map[string]origin.Origin{
				"origin_a": &mockOrigin{
					fetchTicks: func(_ context.Context, pairs []provider.Pair) []provider.Tick {
						panic("panic")
					},
				},
				"origin_b": &mockOrigin{
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
			l,
		)
		require.NoError(t, u.Update(context.Background(), g))
		assert.Error(t, g[0].Tick().Validate())
		assert.Contains(t, g[0].Tick().Validate().Error(), "tick is not set")
		assert.Contains(t, logs, "Panic while fetching ticks")
		assert.Equal(t, bn.Float(42), g[1].Tick().Price)
	})
}
