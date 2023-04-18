package origin

import (
	"context"
	"time"

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

// withError is a helper function which returns a list of ticks for the given
// pairs with the given error.
func withError(pairs []provider.Pair, err error) []provider.Tick {
	var ticks []provider.Tick
	for _, pair := range pairs {
		ticks = append(ticks, provider.Tick{
			Pair:  pair,
			Time:  time.Now(),
			Error: err,
		})
	}
	return ticks
}
