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

const ether = 1e18

type ContractAddresses map[string]types.Address

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

func (c ContractAddresses) ByPair(p value.Pair) (types.Address, bool, error) {
	for key, address := range c {
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
			return address, baseIndex > quoteIndex, nil
		}
	}
	// not found the pair
	return types.ZeroAddress, false, fmt.Errorf("failed to get contract address for pair: %s", p.String())
}
