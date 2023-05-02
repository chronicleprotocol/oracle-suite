package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

// WrapperNode is a node that wraps another node with a custom meta.
// It is useful for adding additional information to the node to better
// describe a price model.
type WrapperNode struct {
	branch Node
	meta   provider.Meta
}

// NewWrapperNode creates a new WrapperNode instance.
func NewWrapperNode(node Node, meta provider.Meta) *WrapperNode {
	return &WrapperNode{branch: node, meta: meta}
}

// AddBranch implements the Node interface.
func (n *WrapperNode) AddBranch(branch ...Node) error {
	if len(branch) == 0 {
		return nil
	}
	if len(branch) > 1 {
		return fmt.Errorf("only one branch is allowed")
	}
	return nil
}

// Branches implements the Node interface.
func (n *WrapperNode) Branches() []Node {
	if n.branch == nil {
		return nil
	}
	return []Node{n.branch}
}

// Pair implements the Node interface.
func (n *WrapperNode) Pair() provider.Pair {
	return n.branch.Pair()
}

// Tick implements the Node interface.
func (n *WrapperNode) Tick() provider.Tick {
	if n.branch == nil {
		return provider.Tick{
			Error: fmt.Errorf("branch is not set (this is likely a bug)"),
		}
	}
	tick := n.branch.Tick()
	tick.SubTicks = []provider.Tick{tick}
	tick.Meta = n.meta
	return tick
}

// Meta implements the Node interface.
func (n *WrapperNode) Meta() provider.Meta {
	return n.meta
}
