//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package nodes

import (
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

// Node represents generics node in a graph.
type Node interface {
	Children() []Node
}

// Parent represents a node to which you can add a child node.
type Parent interface {
	Node
	AddChild(node Node)
}

// Aggregator represents a node which can aggregate prices from its children.
type Aggregator interface {
	Node
	Pair() provider.Pair
	Price() AggregatorPrice
}

// Origin represents a node which provides price directly from an origin.
type Origin interface {
	Node
	OriginPair() OriginPair
	Price() OriginPrice
}

func Walk(fn func(Node), nodes ...Node) {
	r := map[Node]struct{}{}

	for _, node := range nodes {
		var recur func(Node)
		recur = func(node Node) {
			if _, ok := r[node]; ok {
				return
			}

			r[node] = struct{}{}
			for _, n := range node.Children() {
				recur(n)
			}
		}
		recur(node)
	}

	for n := range r {
		fn(n)
	}
}

// DetectCycle detects cycle in given graph. If cycle is
// detected then path is returned, otherwise empty slice.
func DetectCycle(node Node) []Node {
	visited := map[Node]struct{}{}

	var recur func(Node, []Node) []Node
	recur = func(node Node, parents []Node) []Node {
		// If node already appeared in the parents list, it means that given
		// graph is cyclic.
		for _, p := range parents {
			if p == node {
				return parents
			}
		}
		// Skip checking for already visited nodes.
		if _, ok := visited[node]; ok {
			return nil
		}
		visited[node] = struct{}{}
		parents = append(parents, node)
		for _, n := range node.Children() {
			// We have to copy list for each child, because each node
			// have different list of parents.
			parentsCpy := make([]Node, len(parents))
			copy(parentsCpy, parents)
			if p := recur(n, parentsCpy); p != nil {
				return p
			}
		}
		return nil
	}

	return recur(node, nil)
}
