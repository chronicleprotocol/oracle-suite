package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type IndirectMeta struct {
	// Ticks is a list of ticks used to calculate cross rate.
	Ticks []provider.Tick
}

func (m IndirectMeta) Meta() map[string]interface{} {
	return map[string]interface{}{
		"node":  "indirect",
		"ticks": m.Ticks,
	}
}

// IndirectNode is a node that calculates cross rate from the list
// of ticks from its branches. The cross rate is calculated from the first
// tick to the last tick hence the order of branches is important.
type IndirectNode struct {
	pair     provider.Pair
	branches []Node
}

// NewIndirectNode creates a new IndirectNode instance.
//
// The pair argument is a pair to which the cross rate must be resolved.
func NewIndirectNode(pair provider.Pair) *IndirectNode {
	return &IndirectNode{
		pair: pair,
	}
}

// AddBranch implements the Node interface.
func (i *IndirectNode) AddBranch(branch ...Node) error {
	i.branches = append(i.branches, branch...)
	return nil
}

// Branches implements the Node interface.
func (i *IndirectNode) Branches() []Node {
	return i.branches
}

// Pair implements the Node interface.
func (i *IndirectNode) Pair() provider.Pair {
	return i.pair
}

// Tick implements the Node interface.
func (i *IndirectNode) Tick() provider.Tick {
	var ticks []provider.Tick
	for _, branch := range i.branches {
		ticks = append(ticks, branch.Tick())
	}
	meta := IndirectMeta{Ticks: ticks}
	for _, tick := range ticks {
		if err := tick.Validate(); err != nil {
			return provider.Tick{
				Pair:  i.pair,
				Meta:  meta,
				Error: fmt.Errorf("invalid tick: %w", err),
			}
		}
	}
	indirect, err := crossRate(ticks)
	if err != nil {
		return provider.Tick{
			Pair:  i.pair,
			Meta:  meta,
			Error: err,
		}
	}
	if !indirect.Pair.Equal(i.pair) {
		return provider.Tick{
			Pair:  i.pair,
			Meta:  meta,
			Error: fmt.Errorf("expected pair %s, got %s", i.pair, indirect.Pair),
		}
	}
	return provider.Tick{
		Pair:  indirect.Pair,
		Price: indirect.Price,
		Time:  indirect.Time,
		Meta:  meta,
	}
}

// crossRate returns a calculated price from the list of prices. Prices order
// is important because prices are calculated from first to last.
//
//nolint:gocyclo,funlen
func crossRate(t []provider.Tick) (provider.Tick, error) {
	if len(t) == 0 {
		return provider.Tick{}, nil
	}
	for i := 0; i < len(t)-1; i++ {
		a := t[i]
		b := t[i+1]
		var (
			pair  provider.Pair
			price *bn.FloatNumber
		)
		switch {
		case a.Pair.Quote == b.Pair.Quote: // A/C, B/C
			pair.Base = a.Pair.Base
			pair.Quote = b.Pair.Base
			if b.Price.Sign() > 0 {
				price = a.Price.Div(b.Price)
			} else {
				price = bn.Float(0)
			}
		case a.Pair.Base == b.Pair.Base: // C/A, C/B
			pair.Base = a.Pair.Quote
			pair.Quote = b.Pair.Quote
			if a.Price.Sign() > 0 {
				price = b.Price.Div(a.Price)
			} else {
				price = bn.Float(0)
			}
		case a.Pair.Quote == b.Pair.Base: // A/C, C/B
			pair.Base = a.Pair.Base
			pair.Quote = b.Pair.Quote
			price = a.Price.Mul(b.Price)
		case a.Pair.Base == b.Pair.Quote: // C/A, B/C
			pair.Base = a.Pair.Quote
			pair.Quote = b.Pair.Base
			if a.Price.Sign() > 0 && b.Price.Sign() > 0 {
				price = bn.Float(1).Div(b.Price).Div(a.Price)
			} else {
				price = bn.Float(0)
			}
		default:
			return a, fmt.Errorf("unable to calculate cross rate for %s and %s", a.Pair, b.Pair)
		}
		b.Pair = pair
		b.Price = price
		if a.Time.Before(b.Time) {
			b.Time = a.Time
		}
		t[i+1] = b
	}
	resolved := t[len(t)-1]
	return provider.Tick{
		Pair:  resolved.Pair,
		Time:  resolved.Time,
		Price: resolved.Price,
	}, nil
}
