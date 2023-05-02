package graph

import (
	"fmt"
	"sort"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// MedianNode is a node that calculates median price from its
// branches.
type MedianNode struct {
	pair     provider.Pair
	min      int
	branches []Node
}

// NewMedianNode creates a new MedianNode instance.
//
// The min argument is a minimum number of valid prices obtained from
// branches required to calculate median.
func NewMedianNode(pair provider.Pair, min int) *MedianNode {
	return &MedianNode{
		pair: pair,
		min:  min,
	}
}

// AddBranch implements the Node interface.
func (n *MedianNode) AddBranch(branch ...Node) error {
	n.branches = append(n.branches, branch...)
	return nil
}

// Branches implements the Node interface.
func (n *MedianNode) Branches() []Node {
	return n.branches
}

// Pair implements the Node interface.
func (n *MedianNode) Pair() provider.Pair {
	return n.pair
}

// Tick implements the Node interface.
func (n *MedianNode) Tick() provider.Tick {
	var (
		tm     time.Time
		ticks  []provider.Tick
		prices []*bn.FloatNumber
	)

	meta := n.Meta()

	// Collect all ticks from branches and prices from ticks
	// that can be used to calculate median.
	for _, branch := range n.branches {
		tick := branch.Tick()
		if tm.IsZero() {
			tm = tick.Time
		}
		if tick.Time.Before(tm) {
			tm = tick.Time
		}
		ticks = append(ticks, tick)
		if !n.pair.Equal(tick.Pair) {
			continue
		}
		if err := tick.Validate(); err != nil {
			continue
		}
		prices = append(prices, tick.Price)
	}

	// Verify that we have enough valid prices to calculate median.
	if len(prices) < n.min {
		return provider.Tick{
			Pair:     n.pair,
			Meta:     meta,
			SubTicks: ticks,
			Error:    fmt.Errorf("not enough prices to calculate median"),
		}
	}

	// Return median tick.
	return provider.Tick{
		Pair:     n.pair,
		Price:    median(prices),
		Time:     tm,
		SubTicks: ticks,
		Meta:     meta,
	}
}

// Meta implements the Node interface.
func (n *MedianNode) Meta() provider.Meta {
	return MapMeta{"type": "median", "min_sources": n.min}
}

func median(xs []*bn.FloatNumber) *bn.FloatNumber {
	count := len(xs)
	if count == 0 {
		return nil
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i].Cmp(xs[j]) < 0
	})
	if count%2 == 0 {
		m := count / 2
		x1 := xs[m-1]
		x2 := xs[m]
		return x1.Add(x2).Div(bn.Float(2))
	}
	return xs[(count-1)/2]
}
