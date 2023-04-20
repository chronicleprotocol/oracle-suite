package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

type InvertMeta struct {
	// Tick is a tick used to calculate inverted price.
	Tick provider.Tick
}

func (m InvertMeta) Meta() map[string]any {
	return map[string]any{
		"node": "invert",
		"tick": m.Tick,
	}
}

// InvertNode is a node that inverts the pair and price. E.g. if the pair is
// BTC/USD and the price is 1000, then the pair will be USD/BTC and the price
// will be 0.001.
type InvertNode struct {
	pair   provider.Pair
	branch Node
}

// NewInvertNode creates a new InvertNode instance.
func NewInvertNode(pair provider.Pair) *InvertNode {
	return &InvertNode{pair: pair}
}

// AddBranch implements the Node interface.
func (i *InvertNode) AddBranch(branch ...Node) error {
	if len(branch) == 0 {
		return nil
	}
	if i.branch != nil {
		return fmt.Errorf("branch already exists")
	}
	if len(branch) != 1 {
		return fmt.Errorf("only 1 branch is allowed")
	}
	if !branch[0].Pair().Equal(i.pair.Invert()) {
		return fmt.Errorf("expected pair %s, got %s", i.pair, branch[0].Pair())
	}
	i.branch = branch[0]
	return nil
}

// Branches implements the Node interface.
func (i *InvertNode) Branches() []Node {
	if i.branch == nil {
		return nil
	}
	return []Node{i.branch}
}

// Pair implements the Node interface.
func (i *InvertNode) Pair() provider.Pair {
	return i.pair
}

// Tick implements the Node interface.
func (i *InvertNode) Tick() provider.Tick {
	if i.branch == nil {
		return provider.Tick{
			Pair:  i.pair,
			Error: fmt.Errorf("branch is not set"),
		}
	}
	tick := i.branch.Tick()
	tick.Pair = i.pair.Invert()
	tick.Price = tick.Price.Inv()
	tick.Volume24h = tick.Volume24h.Div(tick.Price)
	tick.Meta = &InvertMeta{Tick: tick}
	return tick
}
