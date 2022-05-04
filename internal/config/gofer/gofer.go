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

package gofer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/internal/query"
	pkgEthereum "github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/nodes"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/origins"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/rpc"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const defaultTTL = 60 * time.Second
const maxTTL = 300 * time.Second

type ErrCyclicReference struct {
	Pair gofer.Pair
	Path []nodes.Node
}

func (e ErrCyclicReference) Error() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("a cyclic reference was detected for the %s pair: ", e.Path))
	for i, n := range e.Path {
		t := reflect.TypeOf(n).String()
		switch typedNode := n.(type) {
		case nodes.Aggregator:
			s.WriteString(fmt.Sprintf("%s(%s)", t, typedNode.Pair()))
		default:
			s.WriteString(t)
		}
		if i != len(e.Path)-1 {
			s.WriteString(" -> ")
		}
	}
	return s.String()
}

type Gofer struct {
	RPC         RPC                   `json:"rpc"`
	Origins     map[string]Origin     `json:"origins"`
	PriceModels map[string]PriceModel `json:"priceModels"`
}

type RPC struct {
	Address string `json:"address"`
}

type Origin struct {
	Type   string          `json:"type"`
	Name   string          `json:"name"`
	URL    string          `json:"url"`
	Params json.RawMessage `json:"params"`
}

type PriceModel struct {
	Method  string          `json:"method"`
	Sources [][]Source      `json:"sources"`
	Params  json.RawMessage `json:"params"`
	TTL     int             `json:"ttl"`
}

type MedianPriceModel struct {
	MinSourceSuccess int `json:"minimumSuccessfulSources"`
}

type Source struct {
	Origin string `json:"origin"`
	Pair   string `json:"pair"`
	TTL    int    `json:"ttl"`
}

// ConfigureRPCAgent returns a new rpc.Agent instance.
func (c *Gofer) ConfigureRPCAgent(cli pkgEthereum.Client, gof gofer.Gofer, logger log.Logger) (*rpc.Agent, error) {
	srv, err := rpc.NewAgent(rpc.AgentConfig{
		Gofer:   gof,
		Network: "tcp",
		Address: c.RPC.Address,
		Logger:  logger,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize RPC agent: %w", err)
	}
	return srv, nil
}

// ConfigureAsyncGofer returns a new async gofer instance.
func (c *Gofer) ConfigureAsyncGofer(cli pkgEthereum.Client, logger log.Logger) (gofer.StartableGofer, error) {
	gra, err := c.buildGraphs()
	if err != nil {
		return nil, fmt.Errorf("unable to load price models: %w", err)
	}
	var ns []nodes.Node
	for _, n := range gra {
		ns = append(ns, n)
	}
	originSet, err := c.buildOrigins(cli)
	if err != nil {
		return nil, err
	}
	fed := feeder.NewFeeder(originSet, logger)
	gof, err := graph.NewAsyncGofer(gra, fed, ns, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize RPC agent: %w", err)
	}
	return gof, nil
}

// ConfigureGofer returns a new async gofer instance.
func (c *Gofer) ConfigureGofer(cli pkgEthereum.Client, logger log.Logger, noRPC bool) (gofer.Gofer, error) {
	if c.RPC.Address == "" || noRPC {
		gra, err := c.buildGraphs()
		if err != nil {
			return nil, fmt.Errorf("unable to load price models: %w", err)
		}
		originSet, err := c.buildOrigins(cli)
		if err != nil {
			return nil, err
		}
		fed := feeder.NewFeeder(originSet, logger)
		gof := graph.NewGofer(gra, fed)
		return gof, nil
	}
	return c.configureRPCClient()
}

// configureRPCClient returns a new rpc.RPC instance.
func (c *Gofer) configureRPCClient() (*rpc.Gofer, error) {
	return rpc.NewGofer("tcp", c.RPC.Address)
}

func (c *Gofer) buildOrigins(cli pkgEthereum.Client) (*origins.Set, error) {
	const defaultWorkerCount = 5
	wp := query.NewHTTPWorkerPool(defaultWorkerCount)
	originSet := origins.DefaultOriginSet(wp, defaultWorkerCount)
	for name, origin := range c.Origins {
		handler, err := NewHandler(origin.Type, wp, cli, origin.URL, origin.Params)
		if err != nil || handler == nil {
			return nil, fmt.Errorf("failed to initiate %s origin with name %s due to error: %w",
				origin.Type, origin.Name, err)
		}
		originSet.SetHandler(name, handler)
	}
	return originSet, nil
}

func (c *Gofer) buildGraphs() (map[gofer.Pair]nodes.Aggregator, error) {
	var err error

	graphs := map[gofer.Pair]nodes.Aggregator{}

	// It's important to create root nodes before branches, because branches
	// may refer to another root nodes instances.
	err = c.buildRoots(graphs)
	if err != nil {
		return nil, err
	}

	err = c.buildBranches(graphs)
	if err != nil {
		return nil, err
	}

	err = c.detectCycle(graphs)
	if err != nil {
		return nil, err
	}

	return graphs, nil
}

func (c *Gofer) buildRoots(graphs map[gofer.Pair]nodes.Aggregator) error {
	for name, model := range c.PriceModels {
		modelPair, err := gofer.NewPair(name)
		if err != nil {
			return err
		}

		switch model.Method {
		case "median":
			var params MedianPriceModel
			if model.Params != nil {
				err := json.Unmarshal(model.Params, &params)
				if err != nil {
					return err
				}
			}
			graphs[modelPair] = nodes.NewMedianAggregatorNode(modelPair, params.MinSourceSuccess)
		default:
			return fmt.Errorf("unknown method %s for pair %s", model.Method, name)
		}
	}

	return nil
}

func (c *Gofer) buildBranches(graphs map[gofer.Pair]nodes.Aggregator) error {
	for name, model := range c.PriceModels {
		// We can ignore error here, because it was checked already
		// in buildRoots method.
		modelPair, _ := gofer.NewPair(name)

		var parent nodes.Parent
		if typedNode, ok := graphs[modelPair].(nodes.Parent); ok {
			parent = typedNode
		} else {
			return fmt.Errorf(
				"%s must implement the nodes.Parent interface",
				reflect.TypeOf(graphs[modelPair]).Elem().String(),
			)
		}

		for _, sources := range model.Sources {
			var children []nodes.Node
			for _, source := range sources {
				var err error
				var node nodes.Node

				if source.Origin == "." {
					node, err = c.reference(graphs, source)
					if err != nil {
						return err
					}
				} else {
					node, err = c.originNode(model, source)
					if err != nil {
						return err
					}
				}

				children = append(children, node)
			}

			// If there are provided multiple sources it means, that the price
			// have to be calculated by using the nodes.IndirectAggregatorNode.
			// Otherwise we can pass that nodes.OriginNode directly to
			// the parent node.
			var node nodes.Node
			if len(children) == 1 {
				node = children[0]
			} else {
				indirectAggregator := nodes.NewIndirectAggregatorNode(modelPair)
				for _, c := range children {
					indirectAggregator.AddChild(c)
				}
				node = indirectAggregator
			}

			parent.AddChild(node)
		}
	}

	return nil
}

func (c *Gofer) reference(graphs map[gofer.Pair]nodes.Aggregator, source Source) (nodes.Node, error) {
	sourcePair, err := gofer.NewPair(source.Pair)
	if err != nil {
		return nil, err
	}

	if _, ok := graphs[sourcePair]; !ok {
		return nil, fmt.Errorf(
			"unable to find price model for the %s pair",
			sourcePair,
		)
	}

	return graphs[sourcePair].(nodes.Node), nil
}

func (c *Gofer) originNode(model PriceModel, source Source) (nodes.Node, error) {
	sourcePair, err := gofer.NewPair(source.Pair)
	if err != nil {
		return nil, err
	}

	originPair := nodes.OriginPair{
		Origin: source.Origin,
		Pair:   sourcePair,
	}

	ttl := defaultTTL
	if model.TTL > 0 {
		ttl = time.Second * time.Duration(model.TTL)
	}
	if source.TTL > 0 {
		ttl = time.Second * time.Duration(source.TTL)
	}

	return nodes.NewOriginNode(originPair, ttl, ttl+maxTTL), nil
}

func (c *Gofer) detectCycle(graphs map[gofer.Pair]nodes.Aggregator) error {
	for _, pair := range sortGraphs(graphs) {
		if path := nodes.DetectCycle(graphs[pair]); len(path) > 0 {
			return ErrCyclicReference{Pair: pair, Path: path}
		}
	}

	return nil
}

func sortGraphs(graphs map[gofer.Pair]nodes.Aggregator) []gofer.Pair {
	var ps []gofer.Pair
	for p := range graphs {
		ps = append(ps, p)
	}
	sort.SliceStable(ps, func(i, j int) bool {
		return ps[i].String() < ps[j].String()
	})
	return ps
}
