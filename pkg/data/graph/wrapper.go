package graph

import (
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
)

// WrapperNode is a node that wraps another node with a custom meta.
// It is useful for adding additional information to the node to better
// describe a price model.
type WrapperNode struct {
	node Node
	meta map[string]any
}

// NewWrapperNode creates a new WrapperNode instance.
func NewWrapperNode(node Node, meta map[string]any) *WrapperNode {
	return &WrapperNode{node: node, meta: meta}
}

// AddNodes implements the Node interface.
//
// Only one node is allowed. If more than one node is added, an error is
// returned.
func (n *WrapperNode) AddNodes(nodes ...Node) error {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) > 1 {
		return fmt.Errorf("only one node is allowed")
	}
	return nil
}

// Nodes implements the Node interface.
func (n *WrapperNode) Nodes() []Node {
	if n.node == nil {
		return nil
	}
	return []Node{n.node}
}

// DataPoint implements the Node interface.
func (n *WrapperNode) DataPoint() data.Point {
	if n.node == nil {
		return data.Point{
			Time:  time.Now(),
			Error: fmt.Errorf("node is not set (this is likely a bug)"),
		}
	}
	dataPoint := n.node.DataPoint()
	dataPoint.SubPoints = []data.Point{dataPoint}
	dataPoint.Meta = n.meta
	return dataPoint
}

// Meta implements the Node interface.
func (n *WrapperNode) Meta() map[string]any {
	return n.meta
}
