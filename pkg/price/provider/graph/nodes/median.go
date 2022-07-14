//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package nodes

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

type ErrNotEnoughSources struct {
	Given int
	Min   int
}

func (e ErrNotEnoughSources) Error() string {
	return fmt.Sprintf(
		"not enough sources to calculate median, %d given but at least %d required",
		e.Given,
		e.Min,
	)
}

type ErrIncompatiblePairs struct {
	Given    provider.Pair
	Expected provider.Pair
}

func (e ErrIncompatiblePairs) Error() string {
	return fmt.Sprintf(
		"unable to calculate median for different pairs, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

// MedianAggregatorNode gets Prices from all of its children and calculates
// median price.
//
//                           -- [Origin A/B]
//                          /
//  [MedianAggregatorNode] ---- [Origin A/B]       -- ...
//                          \                     /
//                           -- [AggregatorNode A/B] ---- ...
//                                                \
//                                                 -- ...
//
// All children of this node must return a Price for the same pair.
type MedianAggregatorNode struct {
	pair       provider.Pair
	minSources int
	children   []Node
}

func NewMedianAggregatorNode(pair provider.Pair, minSources int) *MedianAggregatorNode {
	return &MedianAggregatorNode{
		pair:       pair,
		minSources: minSources,
	}
}

// Children implements the Node interface.
func (n *MedianAggregatorNode) Children() []Node {
	return n.children
}

// AddChild implements the Parent interface.
func (n *MedianAggregatorNode) AddChild(node Node) {
	n.children = append(n.children, node)
}

func (n *MedianAggregatorNode) Pair() provider.Pair {
	return n.pair
}

func (n *MedianAggregatorNode) Price() AggregatorPrice {
	var ts time.Time
	var prices, bids, asks []float64
	var originPrices []OriginPrice
	var aggregatorPrices []AggregatorPrice
	var err error

	for i, c := range n.children {
		// There is no need to copy errors from prices to the MedianAggregatorNode
		// because there may be enough remaining prices to calculate median price.

		var price PairPrice
		switch typedNode := c.(type) {
		case Origin:
			originPrice := typedNode.Price()
			originPrices = append(originPrices, originPrice)
			price = originPrice.PairPrice
			if originPrice.Error != nil {
				continue
			}
		case Aggregator:
			aggregatorPrice := typedNode.Price()
			aggregatorPrices = append(aggregatorPrices, aggregatorPrice)
			price = aggregatorPrice.PairPrice
			if aggregatorPrice.Error != nil {
				continue
			}
		}

		if !n.pair.Equal(price.Pair) {
			err = multierror.Append(
				err,
				ErrIncompatiblePairs{Given: price.Pair, Expected: n.pair},
			)
			continue
		}

		if price.Price > 0 {
			prices = append(prices, price.Price)
		}
		if price.Bid > 0 {
			bids = append(bids, price.Bid)
		}
		if price.Ask > 0 {
			asks = append(asks, price.Ask)
		}
		if i == 0 || price.Time.Before(ts) {
			ts = price.Time
		}
	}

	if len(prices) < n.minSources {
		err = multierror.Append(
			err,
			ErrNotEnoughSources{Given: len(prices), Min: n.minSources},
		)
	}

	return AggregatorPrice{
		PairPrice: PairPrice{
			Pair:      n.pair,
			Price:     median(prices),
			Bid:       median(bids),
			Ask:       median(asks),
			Volume24h: 0,
			Time:      ts,
		},
		OriginPrices:     originPrices,
		AggregatorPrices: aggregatorPrices,
		Parameters:       map[string]string{"method": "median", "minimumSuccessfulSources": strconv.Itoa(n.minSources)},
		Error:            err,
	}
}

func median(xs []float64) float64 {
	count := len(xs)
	if count == 0 {
		return 0
	}

	sort.Float64s(xs)
	if count%2 == 0 {
		m := count / 2
		x1 := xs[m-1]
		x2 := xs[m]
		return (x1 + x2) / 2
	}

	return xs[(count-1)/2]
}
