package origin

import (
	"context"
	"fmt"
	"strings"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
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
	FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error)
}

const ether = 1e18

type ContractAddresses map[string]types.Address

// ByPair returns the contract address and the indexes of tokens, where the contract contains the given pair
// If not found base and quote token, return zero address and -1 for indexes
// For example, if we have a pool address of USDT/WBTC/WETH, and we are looking for USDT/WETH,
// then ByPair return the pool address and the indexes of 0, 2 (index is based on zero)
func (c ContractAddresses) ByPair(p value.Pair) (types.Address, int, int, error) {
	for key, address := range c {
		// key is the list of tokens that the pool contains.
		// It should be listed with the separator '/' and is sorted by ascending order.
		// i.e. `3pool` in curve is the pool of DAI, USDC and USDT,
		// so it is defined as "DAI/USDC/USDT = 0xbebc44782c7db0a1a60cb6fe97d0b483032ff1c7"
		baseIndex := -1
		quoteIndex := -1
		tokens := strings.Split(key, "/")
		for i := range tokens {
			if tokens[i] == p.Base {
				baseIndex = i
				break
			}
		}
		for i := range tokens {
			if tokens[i] == p.Quote {
				quoteIndex = i
				break
			}
		}
		if baseIndex >= 0 && 0 <= quoteIndex && baseIndex != quoteIndex {
			// if p is inverted pair, baseIndex should be greater than quoteIndex
			return address, baseIndex, quoteIndex, nil
		}
	}
	// not found the pair
	return types.ZeroAddress, -1, -1, fmt.Errorf("failed to get contract address for pair: %s", p.String())
}

func convertAddressMap(addresses map[string]string) (ContractAddresses, error) {
	typeAddresses := make(map[string]types.Address)
	for key, address := range addresses {
		tokens := strings.Split(key, "/")
		if len(tokens) < 2 {
			return nil, fmt.Errorf("not found pair: %s", key)
		}
		typeAddresses[key] = types.MustAddressFromHex(address)
	}
	return typeAddresses, nil
}

func fillDataPointsWithError(points map[any]datapoint.Point, pairs []value.Pair, err error) map[any]datapoint.Point {
	var target = points
	if target == nil {
		target = make(map[any]datapoint.Point)
	}
	for _, pair := range pairs {
		target[pair] = datapoint.Point{Error: err}
	}
	return target
}

func queryToPairs(query []any) ([]value.Pair, bool) {
	pairs := make([]value.Pair, len(query))
	for i, q := range query {
		switch q := q.(type) {
		case value.Pair:
			pairs[i] = q
		default:
			return nil, false
		}
	}
	return pairs, true
}
