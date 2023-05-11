package graph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/data/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func newTestProvider() Provider {
	models := map[string]Node{
		"model_a": NewOriginNode(
			"test",
			stringValue("query_a"),
			time.Minute,
			time.Minute,
		),
		"model_b": NewOriginNode(
			"test",
			stringValue("query_b"),
			time.Minute,
			time.Minute,
		),
	}
	updater := NewUpdater(
		map[string]origin.Origin{
			"test": &mockOrigin{
				fetchDataPoints: func(_ context.Context, query []any) (map[any]data.Point, error) {
					points := make(map[any]data.Point, len(query))
					for _, q := range query {
						points[q] = data.Point{
							Value: q.(stringValue),
							Time:  time.Now(),
						}
					}
					return points, nil
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

func TestProvider_DataPoint(t *testing.T) {
	prov := newTestProvider()
	point, err := prov.DataPoint(context.Background(), "model_a")
	require.NoError(t, err)

	// Price must be updated.
	assert.Equal(t, "query_a", point.Value.Print())
}

func TestProvider_DataPoints(t *testing.T) {
	prov := newTestProvider()
	points, err := prov.DataPoints(context.Background(), "model_a", "model_b")
	require.NoError(t, err)

	// Prices must be updated.
	assert.Equal(t, "query_a", points["model_a"].Value.Print())
	assert.Equal(t, "query_b", points["model_b"].Value.Print())
}

func TestProvider_Model(t *testing.T) {
	prov := newTestProvider()
	_, err := prov.Model(context.Background(), "model_a")
	require.NoError(t, err)
}

func TestProvider_Models(t *testing.T) {
	prov := newTestProvider()
	_, err := prov.Models(context.Background(), "model_a", "model_b")
	require.NoError(t, err)
}
