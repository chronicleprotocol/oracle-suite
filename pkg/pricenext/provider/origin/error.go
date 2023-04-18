package origin

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

type ErrPairNotSupported struct {
	Pair provider.Pair
}

func (e ErrPairNotSupported) Error() string {
	return fmt.Sprintf("pair %s not supported", e.Pair)
}
