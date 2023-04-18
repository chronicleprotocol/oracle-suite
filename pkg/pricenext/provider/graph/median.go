package graph

import (
	"fmt"
	"sort"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type NotEnoughSourcesErr struct {
	Pair     provider.Pair
	Expected int
	Given    int
}

func (e NotEnoughSourcesErr) Error() string {
	return fmt.Sprintf(
		"not enough sources to calculate %s median, %d given but at least %d required",
		e.Pair,
		e.Given,
		e.Expected,
	)
}

type IncompatiblePairsErr struct {
	Expected provider.Pair
	Given    provider.Pair
}

func (e IncompatiblePairsErr) Error() string {
	return fmt.Sprintf(
		"unable to calculate median for different pairs, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

type MedianMeta struct {
	// Min is a minimum number of sources required to calculate median.
	Min int

	// Ticks is a list of ticks used to calculate median.
	Ticks []provider.Tick
}

func (m MedianMeta) Meta() map[string]any {
	return map[string]any{
		"aggregator": "median",
		"min":        m.Min,
		"ticks":      m.Ticks,
	}
}

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
func (m *MedianNode) AddBranch(branch ...Node) error {
	m.branches = append(m.branches, branch...)
	return nil
}

func (m *MedianNode) Branches() []Node {
	return m.branches
}

func (m *MedianNode) Pair() provider.Pair {
	return m.pair
}

func (m *MedianNode) Tick() provider.Tick {
	var (
		tm     time.Time
		ticks  []provider.Tick
		prices []*bn.FloatNumber
		warns  Errors
	)
	for _, branch := range m.branches {
		tick := branch.Tick()
		if tm.IsZero() {
			tm = tick.Time
		}
		if tick.Time.Before(tm) {
			tm = tick.Time
		}
		ticks = append(ticks, tick)
		if !m.pair.Equal(tick.Pair) {
			warns = append(warns, IncompatiblePairsErr{
				Given:    tick.Pair,
				Expected: m.pair,
			})
			continue
		}
		if err := tick.Validate(); err != nil {
			warns = append(warns, err)
			continue
		}
		prices = append(prices, tick.Price)
	}
	if len(prices) < m.min {
		return provider.Tick{
			Pair:    m.pair,
			Meta:    MedianMeta{Min: m.min, Ticks: ticks},
			Warning: warns,
			Error: NotEnoughSourcesErr{
				Pair:     m.pair,
				Given:    len(prices),
				Expected: m.min,
			},
		}
	}
	return provider.Tick{
		Pair:    m.pair,
		Price:   median(prices),
		Time:    tm,
		Meta:    MedianMeta{Min: m.min, Ticks: ticks},
		Warning: warns,
	}
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
