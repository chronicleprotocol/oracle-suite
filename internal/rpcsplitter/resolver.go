package rpcsplitter

import (
	"errors"
	"math/big"
	"sort"
)

var notEnoughResponsesErr = errors.New("not enough responses from RPC servers")
var differentResponsesErr = errors.New("RPC servers returned different responses")

type resolver interface {
	resolve([]interface{}) (interface{}, error)
}

type defaultResolver struct {
	minResponses int
}

func (c *defaultResolver) resolve(crs []interface{}) (interface{}, error) {
	if len(crs) < c.minResponses {
		return nil, addError(notEnoughResponsesErr, collectErrors(crs)...)
	}
	if len(crs) == 1 {
		return crs[0], nil
	}
	// Count the number of occurrences of each item by comparing each item
	// in the slice with every other item. The result is stored in a map,
	// where the key is the item itself and the value is the number of
	// occurrences.
	occurs := map[interface{}]int{}
	maxOccurs := 0
	for _, a := range crs {
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
		for _, b := range crs {
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
		return nil, addError(differentResponsesErr, collectErrors(crs)...)
	}
	// Find the item with the maximum number of occurrences.
	var res interface{}
	for cr, o := range occurs {
		if o == maxOccurs {
			if res != nil {
				// If res is not nil it means, that there are multiple items
				// that occurred maxOccurs times. In this case, we cannot
				// determine which one should be chosen.
				return nil, addError(notEnoughResponsesErr, collectErrors(crs)...)
			}
			res = cr
		}
	}
	return res, nil
}

type gasValueResolver struct {
	minResponses int
}

func (c *gasValueResolver) resolve(crs []interface{}) (interface{}, error) {
	ns := filterByNumberType(crs)
	if len(ns) < c.minResponses {
		return nil, addError(notEnoughResponsesErr, collectErrors(crs)...)
	}
	if len(ns) == 1 {
		return crs[0], nil
	}
	if len(ns) == 2 {
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

type blockNumberResolver struct {
	minResponses    int
	maxBlocksBehind int
}

func (c *blockNumberResolver) resolve(crs []interface{}) (interface{}, error) {
	ns := filterByNumberType(crs)
	if len(ns) < c.minResponses {
		return nil, addError(notEnoughResponsesErr, collectErrors(crs)...)
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

func filterByNumberType(crs []interface{}) (s []*numberType) {
	for _, cr := range crs {
		if t, ok := cr.(*numberType); ok {
			s = append(s, t)
		}
	}
	return
}

func collectErrors(crs []interface{}) (errs []error) {
	for _, cr := range crs {
		if t, ok := cr.(error); ok {
			errs = append(errs, t)
		}
	}
	return
}
