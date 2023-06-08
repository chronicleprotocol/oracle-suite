package graph

import (
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// TickIndirectNode is a node that calculates cross rate from the list of price
// ticks from its nodes.
//
// It expects that all nodes return data points with value.Tick values.
//
// The order of nodes is important because prices are calculated from first
// to last. Adjacent nodes must have one common asset.
type TickIndirectNode struct {
	nodes []Node
}

// NewTickIndirectNode creates a new TickIndirectNode instance.
func NewTickIndirectNode() *TickIndirectNode {
	return &TickIndirectNode{}
}

// AddNodes implements the Node interface.
func (n *TickIndirectNode) AddNodes(nodes ...Node) error {
	n.nodes = append(n.nodes, nodes...)
	return nil
}

// Nodes implements the Node interface.
func (n *TickIndirectNode) Nodes() []Node {
	return n.nodes
}

// DataPoint implements the Node interface.
func (n *TickIndirectNode) DataPoint() datapoint.Point {
	var points []datapoint.Point
	for _, nodes := range n.nodes {
		points = append(points, nodes.DataPoint())
	}
	meta := n.Meta()
	for _, point := range points {
		if err := point.Validate(); err != nil {
			return datapoint.Point{
				Time:  time.Now(),
				Meta:  meta,
				Error: fmt.Errorf("invalid data point: %w", err),
			}
		}
		if _, ok := point.Value.(value.Tick); !ok {
			return datapoint.Point{
				Time:  time.Now(),
				Meta:  meta,
				Error: fmt.Errorf("invalid data point value type: %T, expected value.Tick", point.Value),
			}
		}
	}
	indirect, err := crossRate(points)
	indirect.Meta = meta
	if err != nil {
		return datapoint.Point{
			Time:  time.Now(),
			Meta:  meta,
			Error: err,
		}
	}
	return indirect
}

// Meta implements the Node interface.
func (n *TickIndirectNode) Meta() map[string]any {
	return map[string]any{"type": "indirect"}
}

// crossRate returns a calculated price from the list of prices. Prices order
// is important because prices are calculated from first to last.
func crossRate(points []datapoint.Point) (datapoint.Point, error) {
	if len(points) == 0 {
		return datapoint.Point{}, nil
	}
	if len(points) == 1 {
		return points[0], nil
	}
	for i := 0; i < len(points)-1; i++ {
		ap := points[i]
		bp := points[i+1]
		at := ap.Value.(value.Tick)
		bt := bp.Value.(value.Tick)
		var (
			pair  value.Pair
			price *bn.FloatNumber
		)
		switch {
		case at.Pair.Quote == bt.Pair.Quote: // A/C, B/C
			pair.Base = at.Pair.Base
			pair.Quote = bt.Pair.Base
			if bt.Price.Sign() > 0 {
				price = at.Price.Div(bt.Price)
			} else {
				price = bn.Float(0)
			}
		case at.Pair.Base == bt.Pair.Base: // C/A, C/B
			pair.Base = at.Pair.Quote
			pair.Quote = bt.Pair.Quote
			if at.Price.Sign() > 0 {
				price = bt.Price.Div(at.Price)
			} else {
				price = bn.Float(0)
			}
		case at.Pair.Quote == bt.Pair.Base: // A/C, C/B
			pair.Base = at.Pair.Base
			pair.Quote = bt.Pair.Quote
			price = at.Price.Mul(bt.Price)
		case at.Pair.Base == bt.Pair.Quote: // C/A, B/C
			pair.Base = at.Pair.Quote
			pair.Quote = bt.Pair.Base
			if at.Price.Sign() > 0 && bt.Price.Sign() > 0 {
				price = bn.Float(1).Div(bt.Price).Div(at.Price)
			} else {
				price = bn.Float(0)
			}
		default:
			return ap, fmt.Errorf("unable to calculate cross rate for %s and %s", at.Pair, bt.Pair)
		}
		bt.Pair = pair
		bt.Price = price
		bp.Value = bt
		if ap.Time.Before(bp.Time) {
			bp.Time = ap.Time
		}
		points[i+1] = bp
	}
	resolved := points[len(points)-1]
	return datapoint.Point{
		Value:     resolved.Value,
		Time:      resolved.Time,
		SubPoints: points,
	}, nil
}
