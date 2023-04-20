package origin

import (
	"context"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

// Origin provides tick prices for a given set of pairs from an external
// source.
type Origin interface {
	// FetchTicks fetches ticks for the given pairs.
	//
	// Note that this method does not guarantee that ticks will be returned
	// for all pairs nor in the same order as the pairs. The caller must
	// verify returned data.
	FetchTicks(ctx context.Context, pairs []provider.Pair) []provider.Tick
}
