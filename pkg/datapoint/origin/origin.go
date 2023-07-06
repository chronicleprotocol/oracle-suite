package origin

import (
	"context"
	"fmt"

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

type ContractAddresses map[value.Pair]types.Address

func convertAddressMap(addresses map[string]string) (ContractAddresses, error) {
	typeAddresses := make(map[value.Pair]types.Address)
	for key, address := range addresses {
		pair, err := value.PairFromString(key)
		if err != nil { // return error if invalid pair
			return nil, err
		}
		typeAddresses[pair] = types.MustAddressFromHex(address)
	}
	return typeAddresses, nil
}

func (c ContractAddresses) ByPair(p value.Pair) (types.Address, bool, error) {
	contract, ok := c[p]
	invContract, okInv := c[p.Invert()]

	if ok && !okInv {
		return contract, false, nil
	} else if !ok && okInv {
		return invContract, true, nil
	} // duplicated pairs or not found pair
	return types.ZeroAddress, false, fmt.Errorf("failed to get contract address for pair: %s", p.String())
}
