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

package graph

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/nodes"
)

type ErrPairNotFound struct {
	Pair gofer.Pair
}

func (e ErrPairNotFound) Error() string {
	return fmt.Sprintf("unable to find the %s pair", e.Pair)
}

// Gofer implements the gofer.Gofer interface. It uses a graph structure
// to calculate pairs prices.
type Gofer struct {
	graphs map[gofer.Pair]nodes.Aggregator
	feeder *feeder.Feeder
}

// NewGofer returns a new Gofer instance. If the Feeder is not nil,
// then prices are automatically updated when the Price or Prices methods are
// called. Otherwise prices have to be updated externally.
func NewGofer(g map[gofer.Pair]nodes.Aggregator, f *feeder.Feeder) *Gofer {
	return &Gofer{graphs: g, feeder: f}
}

// Models implements the gofer.Gofer interface.
func (g *Gofer) Models(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Model, error) {
	ns, err := g.findNodes(pairs...)
	if err != nil {
		return nil, err
	}
	res := make(map[gofer.Pair]*gofer.Model)
	for _, n := range ns {
		if n, ok := n.(nodes.Aggregator); ok {
			res[n.Pair()] = mapGraphNodes(n)
		}
	}
	return res, nil
}

// Price implements the gofer.Gofer interface.
func (g *Gofer) Price(pair gofer.Pair) (*gofer.Price, error) {
	n, ok := g.graphs[pair]
	if !ok {
		return nil, ErrPairNotFound{Pair: pair}
	}
	if g.feeder != nil {
		g.feeder.Feed(n)
	}
	return mapGraphPrice(n.Price()), nil
}

// Prices implements the gofer.Gofer interface.
func (g *Gofer) Prices(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Price, error) {
	ns, err := g.findNodes(pairs...)
	if err != nil {
		return nil, err
	}
	if g.feeder != nil {
		g.feeder.Feed(ns...)
	}
	res := make(map[gofer.Pair]*gofer.Price)
	for _, n := range ns {
		if n, ok := n.(nodes.Aggregator); ok {
			res[n.Pair()] = mapGraphPrice(n.Price())
		}
	}
	return res, nil
}

// Pairs implements the gofer.Gofer interface.
func (g *Gofer) Pairs() ([]gofer.Pair, error) {
	var ps []gofer.Pair
	for p := range g.graphs {
		ps = append(ps, p)
	}
	return ps, nil
}

// findNodes return root nodes for given pairs. If no nodes are specified,
// then all root nodes are returned.
func (g *Gofer) findNodes(pairs ...gofer.Pair) ([]nodes.Node, error) {
	var ns []nodes.Node
	if len(pairs) == 0 { // Return all:
		for _, n := range g.graphs {
			ns = append(ns, n)
		}
	} else { // Return for given pairs:
		for _, p := range pairs {
			n, ok := g.graphs[p]
			if !ok {
				return nil, ErrPairNotFound{Pair: p}
			}
			ns = append(ns, n)
		}
	}
	return ns, nil
}

func mapGraphNodes(n nodes.Node) *gofer.Model {
	gn := &gofer.Model{
		Type:       strings.TrimLeft(reflect.TypeOf(n).String(), "*"),
		Parameters: make(map[string]string),
	}

	switch typedNode := n.(type) {
	case *nodes.IndirectAggregatorNode:
		gn.Type = "indirect"
		gn.Pair = typedNode.Pair()
	case *nodes.MedianAggregatorNode:
		gn.Type = "median"
		gn.Pair = typedNode.Pair()
	case *nodes.OriginNode:
		gn.Type = "origin"
		gn.Pair = typedNode.OriginPair().Pair
		gn.Parameters["origin"] = typedNode.OriginPair().Origin
	default:
		panic("unsupported node")
	}

	for _, cn := range n.Children() {
		gn.Models = append(gn.Models, mapGraphNodes(cn))
	}

	return gn
}

func mapGraphPrice(t interface{}) *gofer.Price {
	gt := &gofer.Price{
		Parameters: make(map[string]string),
	}

	switch typedPrice := t.(type) {
	case nodes.AggregatorPrice:
		gt.Type = "aggregator"
		gt.Pair = typedPrice.Pair
		gt.Price = typedPrice.Price
		gt.Bid = typedPrice.Bid
		gt.Ask = typedPrice.Ask
		gt.Volume24h = typedPrice.Volume24h
		gt.Time = typedPrice.Time
		if typedPrice.Error != nil {
			gt.Error = typedPrice.Error.Error()
		}
		gt.Parameters = typedPrice.Parameters
		for _, ct := range typedPrice.OriginPrices {
			gt.Prices = append(gt.Prices, mapGraphPrice(ct))
		}
		for _, ct := range typedPrice.AggregatorPrices {
			gt.Prices = append(gt.Prices, mapGraphPrice(ct))
		}
	case nodes.OriginPrice:
		gt.Type = "origin"
		gt.Pair = typedPrice.Pair
		gt.Price = typedPrice.Price
		gt.Bid = typedPrice.Bid
		gt.Ask = typedPrice.Ask
		gt.Volume24h = typedPrice.Volume24h
		gt.Time = typedPrice.Time
		if typedPrice.Error != nil {
			gt.Error = typedPrice.Error.Error()
		}
		gt.Parameters["origin"] = typedPrice.Origin
	default:
		panic("unsupported object")
	}

	return gt
}
