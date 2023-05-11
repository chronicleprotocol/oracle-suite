package graph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/data/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/callback"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func TestUpdater(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		g := []Node{
			NewOriginNode(
				"origin_a",
				"query_a",
				time.Minute,
				time.Minute,
			),
			NewOriginNode(
				"origin_b",
				"query_b",
				time.Minute,
				time.Minute,
			),
		}
		u := NewUpdater(
			map[string]origin.Origin{
				"origin_a": &mockOrigin{
					fetchDataPoints: func(_ context.Context, query []any) (map[any]data.Point, error) {
						points := make(map[any]data.Point, len(query))
						for _, q := range query {
							points[q] = data.Point{
								Value: stringValue(q.(string)),
								Time:  time.Now(),
							}
						}
						return points, nil
					},
				},
				"origin_b": &mockOrigin{
					fetchDataPoints: func(_ context.Context, query []any) (map[any]data.Point, error) {
						points := make(map[any]data.Point, len(query))
						for _, q := range query {
							points[q] = data.Point{
								Value: stringValue(q.(string)),
								Time:  time.Now(),
							}
						}
						return points, nil
					},
				},
			},
			null.New(),
		)
		require.NoError(t, u.Update(context.Background(), g))
		assert.Equal(t, "query_a", g[0].DataPoint().Value.Print())
		assert.Equal(t, "query_b", g[1].DataPoint().Value.Print())
	})
	t.Run("fresh tick", func(t *testing.T) {
		g := []Node{
			NewOriginNode(
				"origin_a",
				"query_a",
				time.Minute,
				time.Minute,
			),
		}
		_ = g[0].(*OriginNode).SetDataPoint(data.Point{
			Value: stringValue("this_should_not_be_overwritten"),
			Time:  time.Now(),
		})

		u := NewUpdater(
			map[string]origin.Origin{
				"origin_a": &mockOrigin{
					fetchDataPoints: func(_ context.Context, query []any) (map[any]data.Point, error) {
						points := make(map[any]data.Point, len(query))
						for _, q := range query {
							points[q] = data.Point{
								Value: stringValue(q.(string)),
								Time:  time.Now(),
							}
						}
						return points, nil
					},
				},
			},
			null.New(),
		)
		require.NoError(t, u.Update(context.Background(), g))
		assert.Equal(t, "this_should_not_be_overwritten", g[0].DataPoint().Value.Print())
	})
	t.Run("missing tick", func(t *testing.T) {
		g := []Node{
			NewOriginNode(
				"origin_a",
				"query_a",
				time.Minute,
				time.Minute,
			),
			NewOriginNode(
				"origin_b",
				"query_b",
				time.Minute,
				time.Minute,
			),
		}
		u := NewUpdater(
			map[string]origin.Origin{
				"origin_a": &mockOrigin{
					fetchDataPoints: func(_ context.Context, types []any) (map[any]data.Point, error) {
						return nil, nil
					},
				},
				"origin_b": &mockOrigin{
					fetchDataPoints: func(_ context.Context, query []any) (map[any]data.Point, error) {
						points := make(map[any]data.Point, len(query))
						for _, q := range query {
							points[q] = data.Point{
								Value: stringValue(q.(string)),
								Time:  time.Now(),
							}
						}
						return points, nil
					},
				},
			},
			null.New(),
		)
		require.NoError(t, u.Update(context.Background(), g))
		assert.Error(t, g[0].DataPoint().Validate())
		assert.Contains(t, g[0].DataPoint().Validate().Error(), "data point is not set")
		assert.Equal(t, "query_b", g[1].DataPoint().Value.Print())
	})
	t.Run("panic", func(t *testing.T) {
		var logs []string
		l := callback.New(log.Debug, func(level log.Level, fields log.Fields, log string) {
			logs = append(logs, log)
		})
		g := []Node{
			NewOriginNode(
				"origin_a",
				"query_a",
				time.Minute,
				time.Minute,
			),
			NewOriginNode(
				"origin_b",
				"query_b",
				time.Minute,
				time.Minute,
			),
		}
		u := NewUpdater(
			map[string]origin.Origin{
				"origin_a": &mockOrigin{
					fetchDataPoints: func(_ context.Context, types []any) (map[any]data.Point, error) {
						panic("panic")
					},
				},
				"origin_b": &mockOrigin{
					fetchDataPoints: func(_ context.Context, query []any) (map[any]data.Point, error) {
						points := make(map[any]data.Point, len(query))
						for _, q := range query {
							points[q] = data.Point{
								Value: stringValue(q.(string)),
								Time:  time.Now(),
							}
						}
						return points, nil
					},
				},
			},
			l,
		)
		require.NoError(t, u.Update(context.Background(), g))
		assert.Error(t, g[0].DataPoint().Validate())
		assert.Contains(t, g[0].DataPoint().Validate().Error(), "data point is not set")
		assert.Contains(t, logs, "Panic while fetching data points from the origin")
		assert.Equal(t, "query_b", g[1].DataPoint().Value.Print())
	})
}
