package graph

import (
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/data/origin"
)

// AliasNode is a node that aliases another tick's asset pair.
//
// It expects one node that returns a data point with an origin.Tick value.
type AliasNode struct {
	alias origin.Pair
	node  Node
}

// NewAliasNode creates a new AliasNode instance.
func NewAliasNode(alias origin.Pair) *AliasNode {
	return &AliasNode{alias: alias}
}

// AddNodes implements the Node interface.
//
// Only one node is allowed. If more than one node is added, an error is
// returned.
func (n *AliasNode) AddNodes(nodes ...Node) error {
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
func (n *AliasNode) Nodes() []Node {
	if n.node == nil {
		return nil
	}
	return []Node{n.node}
}

// DataPoint implements the Node interface.
func (n *AliasNode) DataPoint() data.Point {
	if n.node == nil {
		return data.Point{
			Time:  time.Now(),
			Meta:  n.Meta(),
			Error: fmt.Errorf("node is not set"),
		}
	}
	point := n.node.DataPoint()
	tick, ok := point.Value.(origin.Tick)
	if !ok {
		return data.Point{
			Time:  time.Now(),
			Meta:  n.Meta(),
			Error: fmt.Errorf("invalid data point, expected origin.Tick"),
		}
	}
	tick.Pair = n.alias
	return data.Point{
		Value:     tick,
		Time:      point.Time,
		SubPoints: []data.Point{point},
		Meta:      n.Meta(),
		Error:     point.Error,
	}
}

// Meta implements the Node interface.
func (n *AliasNode) Meta() map[string]any {
	return map[string]any{
		"type":  "alias",
		"alias": n.alias,
	}
}
