package graph

import (
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
)

// TickAliasNode is a node that aliases another tick's asset pair.
//
// It expects one node that returns a data point with an value.Tick value.
type TickAliasNode struct {
	alias value.Pair
	node  Node
}

// NewTickAliasNode creates a new TickAliasNode instance.
func NewTickAliasNode(alias value.Pair) *TickAliasNode {
	return &TickAliasNode{alias: alias}
}

// AddNodes implements the Node interface.
//
// Only one node is allowed. If more than one node is added, an error is
// returned.
func (n *TickAliasNode) AddNodes(nodes ...Node) error {
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
func (n *TickAliasNode) Nodes() []Node {
	if n.node == nil {
		return nil
	}
	return []Node{n.node}
}

// DataPoint implements the Node interface.
func (n *TickAliasNode) DataPoint() datapoint.Point {
	if n.node == nil {
		return datapoint.Point{
			Time:  time.Now(),
			Meta:  n.Meta(),
			Error: fmt.Errorf("node is not set"),
		}
	}
	point := n.node.DataPoint()
	tick, ok := point.Value.(value.Tick)
	if !ok {
		return datapoint.Point{
			Time:  time.Now(),
			Meta:  n.Meta(),
			Error: fmt.Errorf("invalid data point, expected value.Tick"),
		}
	}
	tick.Pair = n.alias
	return datapoint.Point{
		Value:     tick,
		Time:      point.Time,
		SubPoints: []datapoint.Point{point},
		Meta:      n.Meta(),
		Error:     point.Error,
	}
}

// Meta implements the Node interface.
func (n *TickAliasNode) Meta() map[string]any {
	return map[string]any{
		"type":  "alias",
		"alias": n.alias,
	}
}
