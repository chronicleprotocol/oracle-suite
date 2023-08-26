//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package graph

import (
	"fmt"
	"sort"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// TickMedianNode is a node that calculates median value from its
// nodes.
//
// It expects that all nodes return data points with value.Tick values.
type TickMedianNode struct {
	min   int
	nodes []Node
}

// NewTickMedianNode creates a new TickMedianNode instance.
//
// The min argument is a minimum number of valid prices obtained from
// nodes required to calculate median price.
func NewTickMedianNode(min int) *TickMedianNode {
	return &TickMedianNode{
		min: min,
	}
}

// AddNodes implements the Node interface.
func (n *TickMedianNode) AddNodes(nodes ...Node) error {
	n.nodes = append(n.nodes, nodes...)
	return nil
}

// Nodes implements the Node interface.
func (n *TickMedianNode) Nodes() []Node {
	return n.nodes
}

// DataPoint implements the Node interface.
func (n *TickMedianNode) DataPoint() datapoint.Point {
	var (
		tm     time.Time
		points []datapoint.Point
		ticks  []value.Tick
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
		tick, ok := point.Value.(value.Tick)
		if !ok {
			return datapoint.Point{
				Time:  time.Now(),
				Meta:  n.Meta(),
				Error: fmt.Errorf("invalid data point value, expected value.Tick"),
			}
		}
		if len(ticks) > 0 && !ticks[len(ticks)-1].Pair.Equal(tick.Pair) {
			return datapoint.Point{
				Time:  time.Now(),
				Meta:  n.Meta(),
				Error: fmt.Errorf("invalid data point value, expected value.Tick for pair %s", ticks[len(ticks)-1].Pair),
			}
		}
		ticks = append(ticks, tick)
		prices = append(prices, tick.Price)
	}

	// Verify that we have enough valid values to calculate median.
	if len(ticks) < n.min {
		return datapoint.Point{
			Time:      time.Now(),
			SubPoints: points,
			Meta:      n.Meta(),
			Error:     fmt.Errorf("not enough values to calculate median, want %d, got %d", n.min, len(ticks)),
		}
	}

	// Return median tick.
	return datapoint.Point{
		Value: value.Tick{
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
func (n *TickMedianNode) Meta() map[string]any {
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
