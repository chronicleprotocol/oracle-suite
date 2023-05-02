package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

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
//
// Node requires one branch. If more than one branch is added, an error is
// returned.
func (n *InvertNode) AddBranch(branch ...Node) error {
	if len(branch) == 0 {
		return nil
	}
	if n.branch != nil {
		return fmt.Errorf("branch already exists")
	}
	if len(branch) != 1 {
		return fmt.Errorf("only 1 branch is allowed")
	}
	if !branch[0].Pair().Equal(n.pair.Invert()) {
		return fmt.Errorf("expected pair %s, got %s", n.pair, branch[0].Pair())
	}
	n.branch = branch[0]
	return nil
}

// Branches implements the Node interface.
func (n *InvertNode) Branches() []Node {
	if n.branch == nil {
		return nil
	}
	return []Node{n.branch}
}

// Pair implements the Node interface.
func (n *InvertNode) Pair() provider.Pair {
	return n.pair
}

// Tick implements the Node interface.
func (n *InvertNode) Tick() provider.Tick {
	if n.branch == nil {
		return provider.Tick{
			Pair:  n.pair,
			Error: fmt.Errorf("branch is not set"),
		}
	}
	tick := n.branch.Tick()
	tick.Pair = n.pair.Invert()
	tick.Price = tick.Price.Inv()
	tick.Volume24h = tick.Volume24h.Div(tick.Price)
	tick.SubTicks = []provider.Tick{tick}
	tick.Meta = n.Meta()
	return tick
}

// Meta implements the Node interface.
func (n *InvertNode) Meta() provider.Meta {
	return MapMeta{"type": "invert"}
}
