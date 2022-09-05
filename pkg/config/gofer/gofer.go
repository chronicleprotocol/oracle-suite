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
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/nodes"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/origins"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const defaultTTL = 60 * time.Second
const maxTTL = 240 * time.Second

type ErrCyclicReference struct {
	Pair provider.Pair
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
	RPC           RPC                   `yaml:"rpc"` // Old configuration format, to remove in the future.
	RPCListenAddr string                `yaml:"rpcListenAddr"`
	Origins       map[string]Origin     `yaml:"origins"`
	PriceModels   map[string]PriceModel `yaml:"priceModels"`
}

type RPC struct {
	Address string `yaml:"address"`
}

type Origin struct {
	Type   string    `yaml:"type"`
	URL    string    `yaml:"url"` // TODO: Move it to the params field.
	Params yaml.Node `yaml:"params"`
}

type PriceModel struct {
	Method  string     `yaml:"method"`
	Sources [][]Source `yaml:"sources"`
	Params  yaml.Node  `yaml:"params"`
	TTL     int        `yaml:"ttl"`
}

type MedianPriceModel struct {
	MinSourceSuccess int                    `yaml:"minimumSuccessfulSources"`
	PostPriceHook    map[string]interface{} `yaml:"postPriceHook"`
}

type Source struct {
	Origin string `yaml:"origin"`
	Pair   string `yaml:"pair"`
	TTL    int    `yaml:"ttl"`
}

// ConfigureRPCAgent returns a new rpc.Agent instance.
func (c *Gofer) ConfigureRPCAgent(cli ethereum.Client, gof provider.Provider, logger log.Logger) (*rpc.Agent, error) {
	listenAddr := c.RPC.Address
	if len(c.RPCListenAddr) != 0 {
		listenAddr = c.RPCListenAddr
	}
	srv, err := rpc.NewAgent(rpc.AgentConfig{
		Provider: gof,
		Network:  "tcp",
		Address:  listenAddr,
		Logger:   logger,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize RPC agent: %w", err)
	}
	return srv, nil
}

// ConfigureAsyncGofer returns a new async gofer instance.
func (c *Gofer) ConfigureAsyncGofer(cli ethereum.Client, logger log.Logger) (provider.Provider, error) {
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
	gof, err := graph.NewAsyncProvider(gra, fed, ns, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize RPC agent: %w", err)
	}
	return gof, nil
}

func (c *Gofer) ConfigurePriceHook(ctx context.Context, cli ethereum.Client) (provider.PriceHook, error) {
	m := provider.NewHookParams()
	for name, model := range c.PriceModels {
		switch model.Method {
		case "median":
			var params MedianPriceModel
			err := model.Params.Decode(&params)
			if err != nil {
				return nil, err
			}
			if len(params.PostPriceHook) > 0 {
				m[name] = params.PostPriceHook
			}
		default:
		}
	}
	return provider.NewPostPriceHook(ctx, cli, m)
}

// ConfigureGofer returns a new async gofer instance.
func (c *Gofer) ConfigureGofer(cli ethereum.Client, logger log.Logger, noRPC bool) (provider.Provider, error) {
	listenAddr := c.RPC.Address
	if len(c.RPCListenAddr) != 0 {
		listenAddr = c.RPCListenAddr
	}
	if listenAddr == "" || noRPC {
		gra, err := c.buildGraphs()
		if err != nil {
			return nil, fmt.Errorf("unable to load price models: %w", err)
		}
		originSet, err := c.buildOrigins(cli)
		if err != nil {
			return nil, err
		}
		fed := feeder.NewFeeder(originSet, logger)
		gof := graph.NewProvider(gra, fed)
		return gof, nil
	}
	return c.configureRPCClient(listenAddr)
}

// configureRPCClient returns a new rpc.RPC instance.
func (c *Gofer) configureRPCClient(listenAddr string) (*rpc.Provider, error) {
	return rpc.NewProvider("tcp", listenAddr)
}

func (c *Gofer) buildOrigins(cli ethereum.Client) (*origins.Set, error) {
	const defaultWorkerCount = 10
	wp := query.NewHTTPWorkerPool(defaultWorkerCount)
	originSet := origins.DefaultOriginSet(wp)
	for name, origin := range c.Origins {
		handler, err := NewHandler(origin.Type, wp, cli, origin.URL, origin.Params)
		if err != nil || handler == nil {
			return nil, fmt.Errorf(
				"failed to initiate %s origin with name %s due to error: %w", origin.Type, name, err,
			)
		}
		originSet.SetHandler(name, handler)
	}
	return originSet, nil
}

func (c *Gofer) buildGraphs() (map[provider.Pair]nodes.Aggregator, error) {
	var err error

	graphs := map[provider.Pair]nodes.Aggregator{}

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

func (c *Gofer) buildRoots(graphs map[provider.Pair]nodes.Aggregator) error {
	for name, model := range c.PriceModels {
		modelPair, err := provider.NewPair(name)
		if err != nil {
			return err
		}

		switch model.Method {
		case "median":
			var params MedianPriceModel
			if err := model.Params.Decode(&params); err != nil {
				return err
			}
			graphs[modelPair] = nodes.NewMedianAggregatorNode(modelPair, params.MinSourceSuccess)
		default:
			return fmt.Errorf("unknown method %s for pair %s", model.Method, name)
		}
	}

	return nil
}

func (c *Gofer) buildBranches(graphs map[provider.Pair]nodes.Aggregator) error {
	for name, model := range c.PriceModels {
		// We can ignore error here, because it was checked already
		// in buildRoots method.
		modelPair, _ := provider.NewPair(name)

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
			// Otherwise, we can pass that nodes.OriginNode directly to
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

func (c *Gofer) reference(graphs map[provider.Pair]nodes.Aggregator, source Source) (nodes.Node, error) {
	sourcePair, err := provider.NewPair(source.Pair)
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
	sourcePair, err := provider.NewPair(source.Pair)
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

func (c *Gofer) detectCycle(graphs map[provider.Pair]nodes.Aggregator) error {
	for _, pair := range sortGraphs(graphs) {
		if path := nodes.DetectCycle(graphs[pair]); len(path) > 0 {
			return ErrCyclicReference{Pair: pair, Path: path}
		}
	}

	return nil
}

func sortGraphs(graphs map[provider.Pair]nodes.Aggregator) []provider.Pair {
	var ps []provider.Pair
	for p := range graphs {
		ps = append(ps, p)
	}
	sort.SliceStable(ps, func(i, j int) bool {
		return ps[i].String() < ps[j].String()
	})
	return ps
}
