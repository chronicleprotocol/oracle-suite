package origin

import (
	"context"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
)

// Origin provides dataPoint prices for a given set of pairs from an external
// source.
type Origin interface {
	// FetchDataPoints fetches data points for the given list of queries.
	//
	// A query is an any type that can be used to query the origin for a data
	// point. For example, a query could be a pair of assets.
	//
	// Note that this method does not guarantee that data points will be
	// returned for all pairs nor in the same order as the pairs. The caller
	// must verify returned data.
	FetchDataPoints(ctx context.Context, query []any) (map[any]data.Point, error)
}
