package graph

import (
	"fmt"
	"sort"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/data/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// MedianNode is a node that calculates median value from its
// nodes.
//
// It expects that all nodes return data points with origin.Tick values.
type MedianNode struct {
	min   int
	nodes []Node
}

// NewMedianNode creates a new MedianNode instance.
//
// The min argument is a minimum number of valid prices obtained from
// nodes required to calculate median price.
func NewMedianNode(min int) *MedianNode {
	return &MedianNode{
		min: min,
	}
}

// AddNodes implements the Node interface.
func (n *MedianNode) AddNodes(nodes ...Node) error {
	n.nodes = append(n.nodes, nodes...)
	return nil
}

// Nodes implements the Node interface.
func (n *MedianNode) Nodes() []Node {
	return n.nodes
}

// DataPoint implements the Node interface.
func (n *MedianNode) DataPoint() data.Point {
	var (
		tm     time.Time
		points []data.Point
		ticks  []origin.Tick
		prices []*bn.FloatNumber
	)

	// Collect all data points from nodes and that can be used to calculate
	// median.
	for _, node := range n.nodes {
		point := node.DataPoint()
		if tm.IsZero() {
			tm = point.Time
		}
		if point.Time.Before(tm) {
			tm = point.Time
		}
		points = append(points, point)
		if err := point.Validate(); err != nil {
			continue
		}
		tick, ok := point.Value.(origin.Tick)
		if !ok {
			return data.Point{
				Time:  time.Now(),
				Meta:  n.Meta(),
				Error: fmt.Errorf("invalid data point value, expected origin.Tick"),
			}
		}
		if len(ticks) > 0 && !ticks[len(ticks)-1].Pair.Equal(tick.Pair) {
			return data.Point{
				Time:  time.Now(),
				Meta:  n.Meta(),
				Error: fmt.Errorf("invalid data point value, expected origin.Tick for pair %s", ticks[len(ticks)-1].Pair),
			}
		}
		ticks = append(ticks, tick)
		prices = append(prices, tick.Price)
	}

	// Verify that we have enough valid values to calculate median.
	if len(ticks) < n.min {
		return data.Point{
			Time:      time.Now(),
			SubPoints: points,
			Meta:      n.Meta(),
			Error:     fmt.Errorf("not enough values to calculate median"),
		}
	}

	// Return median tick.
	return data.Point{
		Value: origin.Tick{
			Pair:      ticks[0].Pair,
			Price:     median(prices),
			Volume24h: bn.Float(0),
		},
		Time:      tm,
		SubPoints: points,
		Meta:      n.Meta(),
	}
}

// Meta implements the Node interface.
func (n *MedianNode) Meta() map[string]any {
	return map[string]any{
		"type":       "median",
		"min_values": n.min,
	}
}

func median(xs []*bn.FloatNumber) *bn.FloatNumber {
	count := len(xs)
	if count == 0 {
		return nil
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i].Cmp(xs[j]) < 0
	})
	if count%2 == 0 {
		m := count / 2
		x1 := xs[m-1]
		x2 := xs[m]
		return x1.Add(x2).Div(bn.Float(2))
	}
	return xs[(count-1)/2]
}
