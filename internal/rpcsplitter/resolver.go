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

package rpcsplitter

import (
	"errors"
	"math/big"
	"sort"
)

var errNotEnoughResponses = errors.New("not enough responses from RPC servers")
var errDifferentResponses = errors.New("RPC servers returned different responses")

// resolver takes responses from different endpoints and returns a single
// response.
type resolver interface {
	resolve([]interface{}) (interface{}, error)
}

// defaultResolver compares responses with each other and returns
// the most common one.
type defaultResolver struct {
	minResponses int // specifies minimum number of occurrences of the most common response
}

// resolve implements resolver interface.
func (c *defaultResolver) resolve(resps []interface{}) (interface{}, error) {
	if len(resps) == 1 {
		return resps[0], nil
	}
	// Count the number of occurrences of each item by comparing each item
	// in the slice with every other item. The result is stored in a map,
	// where the key is the item itself and the value is the number of
	// occurrences.
	occurs := map[interface{}]int{}
	maxOccurs := 0
	for _, a := range resps {
		// Check if similar item exists already in the occurs map.
		f := false
		for b := range occurs {
			if compare(a, b) {
				f = true
				break
			}
		}
		if f {
			continue
		}
		// Count occurrences.
		for _, b := range resps {
			if compare(a, b) {
				occurs[a]++
				if occurs[a] > maxOccurs {
					maxOccurs = occurs[a]
				}
			}
		}
	}
	// Check if there are enough occurrences of the most common item.
	if maxOccurs < c.minResponses {
		return nil, addError(errDifferentResponses, collectErrors(resps)...)
	}
	// Find the item with the maximum number of occurrences.
	var res interface{}
	for cr, o := range occurs {
		if o == maxOccurs {
			if res != nil {
				// If res is not nil it means, that there are multiple items
				// that occurred maxOccurs times. In this case, we cannot
				// determine which one should be chosen.
				return nil, addError(errNotEnoughResponses, collectErrors(resps)...)
			}
			res = cr
		}
	}
	return res, nil
}

// gasValueResolver is designed to handle responses from methods returning a
// gas value. The way how the response is calculated depends on the number of
// responses:
// One response: returns value as is.
// Two responses: returns the lowest one.
// Three responses: returns the median value.
type gasValueResolver struct {
	minResponses int // specifies minimum number of valid responses
}

// resolve implements resolver interface.
func (c *gasValueResolver) resolve(resps []interface{}) (interface{}, error) {
	ns := filterByNumberType(resps)
	if len(ns) < c.minResponses {
		return nil, addError(errNotEnoughResponses, collectErrors(resps)...)
	}
	if len(ns) == 1 {
		return resps[0], nil
	}
	if len(ns) == 2 {
		// With two correct answers, it is safer to return the lower value.
		// Otherwise, the compromised endpoint may return a very high gas
		// price. If this price is used to determine transaction fees, it
		// could cause clients to lose money on transaction fees.
		a := ns[0].Big()
		b := ns[1].Big()
		if a.Cmp(b) > 0 {
			return (*numberType)(b), nil
		}
		return (*numberType)(a), nil
	}
	// Calculate the median.
	sort.Slice(ns, func(i, j int) bool {
		return ns[i].Big().Cmp(ns[j].Big()) < 0
	})
	if len(ns)%2 == 0 {
		m := len(ns) / 2
		bx := ns[m-1].Big()
		by := ns[m].Big()
		return (*numberType)(new(big.Int).Div(new(big.Int).Add(bx, by), big.NewInt(2))), nil
	}
	return ns[len(ns)/2], nil
}

// blockNumberResolver is designed to handle responses from eth_blockNumber method.
//
// Because some RPC endpoints may be behind others, the blockNumberResolver
// uses the lowest block number of all responses, but the difference from the
// last known cannot be less than specified in the maxBlocksBehind parameter.
type blockNumberResolver struct {
	minResponses    int // specifies minimum number of valid responses
	maxBlocksBehind int // specifies how far behind the last known block the returned block can be
}

// resolve implements resolver interface.
func (c *blockNumberResolver) resolve(resps []interface{}) (interface{}, error) {
	ns := filterByNumberType(resps)
	if len(ns) < c.minResponses {
		return nil, addError(errNotEnoughResponses, collectErrors(resps)...)
	}
	if len(ns) == 1 {
		return ns[0], nil
	}
	high := ns[0].Big()
	for _, n := range ns {
		nb := n.Big()
		if high.Cmp(nb) < 0 {
			high = nb
		}
	}
	block := high
	for _, n := range ns {
		nb := n.Big()
		if new(big.Int).Sub(high, nb).Cmp(big.NewInt(int64(c.maxBlocksBehind))) <= 0 && nb.Cmp(block) < 0 {
			block = nb
		}
	}
	return (*numberType)(block), nil
}

func filterByNumberType(resps []interface{}) (s []*numberType) {
	for _, r := range resps {
		if t, ok := r.(*numberType); ok {
			s = append(s, t)
		}
	}
	return
}

func collectErrors(resps []interface{}) (errs []error) {
	for _, r := range resps {
		if t, ok := r.(error); ok {
			errs = append(errs, t)
		}
	}
	return
}
