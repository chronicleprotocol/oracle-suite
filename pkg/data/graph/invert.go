package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/data/origin"
)

// InvertNode is a node that inverts a price. E.g. if the asset query is BTC/USD
// and the price is 1000, then the asset pair will be USD/BTC and the price
// will be 0.001.
//
// It expects one node that returns a data point with an origin.Tick value.
type InvertNode struct {
	node Node
}

// NewInvertNode creates a new InvertNode instance.
func NewInvertNode() *InvertNode {
	return &InvertNode{}
}

// AddNodes implements the Node interface.
//
// Only one node is allowed. If more than one node is added, an error is
// returned.
func (n *InvertNode) AddNodes(nodes ...Node) error {
	if len(nodes) == 0 {
		return nil
	}
	if n.node != nil {
		return fmt.Errorf("node is already set")
	}
	if len(nodes) != 1 {
		return fmt.Errorf("only 1 node is allowed")
	}
	n.node = nodes[0]
	return nil
}

// Nodes implements the Node interface.
func (n *InvertNode) Nodes() []Node {
	if n.node == nil {
		return nil
	}
	return []Node{n.node}
}

// DataPoint implements the Node interface.
func (n *InvertNode) DataPoint() data.Point {
	if n.node == nil {
		return data.Point{
			Error: fmt.Errorf("node is not set"),
		}
	}
	point := n.node.DataPoint()
	tick, ok := point.Value.(origin.Tick)
	if !ok {
		return data.Point{
			Error: fmt.Errorf("invalid data point, expected origin.Tick"),
		}
	}
	tick.Pair = tick.Pair.Invert()
	if tick.Price.Sign() != 0 {
		tick.Price = tick.Price.Inv()
		tick.Volume24h = tick.Volume24h.Div(tick.Price)
	}
	point.Value = tick
	return point
}

// Meta implements the Node interface.
func (n *InvertNode) Meta() map[string]any {
	return map[string]any{"type": "invert"}
}
