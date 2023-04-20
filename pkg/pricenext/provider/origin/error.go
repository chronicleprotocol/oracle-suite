package origin

import (
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

type ErrPairNotSupported struct {
	Pair provider.Pair
}

func (e ErrPairNotSupported) Error() string {
	return fmt.Sprintf("pair %s not supported", e.Pair)
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
