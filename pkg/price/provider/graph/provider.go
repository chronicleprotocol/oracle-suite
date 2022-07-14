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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/nodes"
)

type ErrPairNotFound struct {
	Pair provider.Pair
}

func (e ErrPairNotFound) Error() string {
	return fmt.Sprintf("unable to find the %s pair", e.Pair)
}

// Provider implements the provider.Provider interface. It uses a graph
// structure to calculate pairs prices.
type Provider struct {
	graphs map[provider.Pair]nodes.Aggregator
	feeder *feeder.Feeder
}

// NewProvider returns a new Provider instance. If the GetByFeeder is not nil,
// then prices are automatically updated when the Price or Prices methods are
// called. Otherwise, prices have to be updated externally.
func NewProvider(graph map[provider.Pair]nodes.Aggregator, feeder *feeder.Feeder) *Provider {
	return &Provider{graphs: graph, feeder: feeder}
}

// Models implements the provider.Provider interface.
func (g *Provider) Models(pairs ...provider.Pair) (map[provider.Pair]*provider.Model, error) {
	ns, err := g.findNodes(pairs...)
	if err != nil {
		return nil, err
	}
	res := make(map[provider.Pair]*provider.Model)
	for _, n := range ns {
		if n, ok := n.(nodes.Aggregator); ok {
			res[n.Pair()] = mapGraphNodes(n)
		}
	}
	return res, nil
}

// Price implements the provider.Provider interface.
func (g *Provider) Price(pair provider.Pair) (*provider.Price, error) {
	n, ok := g.graphs[pair]
	if !ok {
		return nil, ErrPairNotFound{Pair: pair}
	}
	if g.feeder != nil {
		g.feeder.Feed([]nodes.Node{n}, time.Now())
	}
	return mapGraphPrice(n.Price()), nil
}

// Prices implements the provider.Providerinterface.
func (g *Provider) Prices(pairs ...provider.Pair) (map[provider.Pair]*provider.Price, error) {
	ns, err := g.findNodes(pairs...)
	if err != nil {
		return nil, err
	}
	if g.feeder != nil {
		g.feeder.Feed(ns, time.Now())
	}
	res := make(map[provider.Pair]*provider.Price)
	for _, n := range ns {
		if n, ok := n.(nodes.Aggregator); ok {
			res[n.Pair()] = mapGraphPrice(n.Price())
		}
	}
	return res, nil
}

// Pairs implements the provider.Provider interface.
func (g *Provider) Pairs() ([]provider.Pair, error) {
	var ps []provider.Pair
	for p := range g.graphs {
		ps = append(ps, p)
	}
	return ps, nil
}

// findNodes return root nodes for given pairs. If no nodes are specified,
// then all root nodes are returned.
func (g *Provider) findNodes(pairs ...provider.Pair) ([]nodes.Node, error) {
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

func mapGraphNodes(n nodes.Node) *provider.Model {
	gn := &provider.Model{
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

func mapGraphPrice(t interface{}) *provider.Price {
	gt := &provider.Price{
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
