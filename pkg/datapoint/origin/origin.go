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

type ContractAddresses map[string]string

func (c ContractAddresses) ByPair(p value.Pair) (types.Address, int, int, error) {
	var baseIndex = -1
	var quoteIndex = -1
	var address types.Address
	for key, hexAddress := range c {
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
		if baseIndex >= 0 && 0 <= quoteIndex {
			address = types.MustAddressFromHex(hexAddress)
			break
		}
	}
	if baseIndex >= 0 && 0 <= quoteIndex && baseIndex != quoteIndex {
		return address, baseIndex, quoteIndex, nil
	}
	// not found the pair
	return types.ZeroAddress, -1, -1, fmt.Errorf("failed to get contract address for pair: %s", p.String())
}
