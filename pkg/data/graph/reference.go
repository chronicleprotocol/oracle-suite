package graph

import (
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
)

// ReferenceNode is a node that references another node.
type ReferenceNode struct {
	node Node
}

// NewReferenceNode creates a new ReferenceNode instance.
func NewReferenceNode() *ReferenceNode {
	return &ReferenceNode{}
}

// AddNodes implements the Node interface.
func (n *ReferenceNode) AddNodes(nodes ...Node) error {
	if len(nodes) == 0 {
		return nil
	}
	if n.node != nil {
		return fmt.Errorf("node already exists")
	}
	if len(nodes) > 1 {
		return fmt.Errorf("only one node is allowed")
	}
	n.node = nodes[0]
	return nil
}

// Nodes implements the Node interface.
func (n *ReferenceNode) Nodes() []Node {
	if n.node == nil {
		return nil
	}
	return []Node{n.node}
}

// DataPoint implements the Node interface.
func (n *ReferenceNode) DataPoint() data.Point {
	if n.node == nil {
		return data.Point{
			Time:  time.Now(),
			Error: fmt.Errorf("node is not set (this is likely a bug)"),
		}
	}
	dataPoint := n.node.DataPoint()
	dataPoint.SubPoints = []data.Point{dataPoint}
	dataPoint.Meta = n.Meta()
	return dataPoint
}

// Meta implements the Node interface.
func (n *ReferenceNode) Meta() map[string]any {
	return map[string]any{"type": "reference"}
}
