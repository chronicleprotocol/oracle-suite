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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

type OriginPair struct {
	Origin string
	Pair   provider.Pair
}

func (o OriginPair) String() string {
	return fmt.Sprintf("%s %s", o.Pair.String(), o.Origin)
}

type PairPrice struct {
	Pair      provider.Pair
	Price     float64
	Bid       float64
	Ask       float64
	Volume24h float64
	Time      time.Time
}

// OriginPrice represent a price which was sourced directly from an origin.
type OriginPrice struct {
	PairPrice
	// Origin is a name of Price source.
	Origin string
	// Error is a list of optional error messages which may occur during
	// calculating the price. If this string is not empty, then the price
	// value is not reliable.
	Error error
}

// AggregatorPrice represent a price which was calculated by using other prices.
type AggregatorPrice struct {
	PairPrice
	// OriginPrices is a list of all OriginPrices used to calculate Price.
	OriginPrices []OriginPrice
	// AggregatorPrices is a list of all OriginPrices used to calculate Price.
	AggregatorPrices []AggregatorPrice
	// Parameters is a custom list of optional parameters returned by an aggregator.
	Parameters map[string]string
	// Errors is a list of optional error messages which may occur during
	// fetching Price. If this list is not empty, then the price value
	// is not reliable.
	Error error
}
