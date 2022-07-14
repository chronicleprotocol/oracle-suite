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

package spectre

import (
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// prices contains a list of messages.Price's for a single asset pair.
type prices struct {
	prices []*messages.Price
}

// newPricesList creates a new store instance.
func newPricesList(p []*messages.Price) *prices {
	return &prices{
		prices: p,
	}
}

// len returns the number of messages in the list.
func (p *prices) len() int {
	return len(p.prices)
}

// messages returns raw messages list.
func (p *prices) messages() []*messages.Price {
	return p.prices
}

// oraclePrices returns oracle prices.
func (p *prices) oraclePrices() []*oracle.Price {
	var prices []*oracle.Price
	for _, price := range p.prices {
		prices = append(prices, price.Price)
	}
	return prices
}

// truncate removes random prices until the number of remaining prices is equal
// to n. If the number of prices is less or equal to n, it does nothing.
//
// This method is used to reduce number of arguments in transaction which will
// reduce transaction costs.
func (p *prices) truncate(n int64) {
	if int64(len(p.prices)) <= n {
		return
	}

	rand.Shuffle(len(p.prices), func(i, j int) {
		p.prices[i], p.prices[j] = p.prices[j], p.prices[i]
	})

	p.prices = p.prices[0:n]
}

// median calculates the median price for all messages in the list.
func (p *prices) median() *big.Int {
	count := len(p.prices)
	if count == 0 {
		return big.NewInt(0)
	}

	sort.Slice(p.prices, func(i, j int) bool {
		return p.prices[i].Price.Val.Cmp(p.prices[j].Price.Val) < 0
	})

	if count%2 == 0 {
		m := count / 2
		x1 := p.prices[m-1].Price.Val
		x2 := p.prices[m].Price.Val
		return new(big.Int).Div(new(big.Int).Add(x1, x2), big.NewInt(2))
	}

	return p.prices[(count-1)/2].Price.Val
}

// spread calculates the spread between given price and a median price.
// The spread is returned as percentage points.
func (p *prices) spread(price *big.Int) float64 {
	if len(p.prices) == 0 || price.Cmp(big.NewInt(0)) == 0 {
		return math.Inf(1)
	}

	oldPriceF := new(big.Float).SetInt(price)
	newPriceF := new(big.Float).SetInt(p.median())

	x := new(big.Float).Sub(newPriceF, oldPriceF)
	x = new(big.Float).Quo(x, oldPriceF)
	x = new(big.Float).Mul(x, big.NewFloat(100))
	xf, _ := x.Float64()

	return math.Abs(xf)
}

// clearOlderThan deletes messages which are older than given time.
func (p *prices) clearOlderThan(t time.Time) {
	var prices []*messages.Price
	for _, price := range p.prices {
		if !price.Price.Age.Before(t) {
			prices = append(prices, price)
		}
	}
	p.prices = prices
}
