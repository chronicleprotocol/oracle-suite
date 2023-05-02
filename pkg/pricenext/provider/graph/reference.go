package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

// ReferenceNode is a node that references another node.
type ReferenceNode struct {
	pair   provider.Pair
	branch Node
}

// NewReferenceNode creates a new ReferenceNode instance.
func NewReferenceNode(pair provider.Pair) *ReferenceNode {
	return &ReferenceNode{pair: pair}
}

// AddBranch implements the Node interface.
func (n *ReferenceNode) AddBranch(branch ...Node) error {
	if len(branch) == 0 {
		return nil
	}
	if n.branch != nil {
		return fmt.Errorf("branch already exists")
	}
	if len(branch) != 1 {
		return fmt.Errorf("expected 1 branch, got %d", len(branch))
	}
	if !branch[0].Pair().Equal(n.pair) {
		return fmt.Errorf("expected pair %s, got %s", n.pair, branch[0].Pair())
	}
	n.branch = branch[0]
	return nil
}

// Branches implements the Node interface.
func (n *ReferenceNode) Branches() []Node {
	if n.branch == nil {
		return nil
	}
	return []Node{n.branch}
}

// Pair implements the Node interface.
func (n *ReferenceNode) Pair() provider.Pair {
	return n.pair
}

// Tick implements the Node interface.
func (n *ReferenceNode) Tick() provider.Tick {
	if n.branch == nil {
		return provider.Tick{
			Pair:  n.pair,
			Error: fmt.Errorf("branch is not set (this is likely a bug)"),
		}
	}
	tick := n.branch.Tick()
	tick.SubTicks = []provider.Tick{tick}
	tick.Meta = n.Meta()
	return tick
}

// Meta implements the Node interface.
func (n *ReferenceNode) Meta() provider.Meta {
	return MapMeta{"type": "reference"}
}
