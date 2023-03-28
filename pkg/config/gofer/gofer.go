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

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/nodes"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/origins"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/rpc"

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

type Dependencies struct {
	Clients ethereumConfig.ClientRegistry
	Logger  log.Logger
}

type AsyncDependencies struct {
	Clients ethereumConfig.ClientRegistry
	Logger  log.Logger
}

type AgentDependencies struct {
	Provider provider.Provider
	Logger   log.Logger
}

type HookDependencies struct {
	Context context.Context
	Clients ethereumConfig.ClientRegistry
}

type ConfigGofer struct {
	// RPCListenAddr is the address on which the RPC server will listen.
	RPCListenAddr string `hcl:"rpc_listen_addr,optional"`

	// RPCAgentAddr is the address of the RPC agent.
	RPCAgentAddr string `hcl:"rpc_agent_addr,optional"`

	// Origins is a configuration of price origins.
	Origins []configOrigin `hcl:"origin,block"`

	// PriceModels is a configuration of price models.
	PriceModels []configSource `hcl:"price_model,block"`

	// Hooks is a configuration of hooks.
	Hooks []configHook `hcl:"hook,block"`

	// Configured service:
	priceProvider      provider.Provider
	asyncPriceProvider provider.Provider
	priceHook          provider.PriceHook
	rpcAgent           *rpc.Agent
}

type configOrigin struct {
	// Origin is the name of the origin.
	Origin string `hcl:",label"`

	// Type is the type of the origin, e.g. "uniswap", "kraken" etc.
	Type string `hcl:"type"`

	// Params is the configuration of the origin.
	Params cty.Value `hcl:"params,optional"`
}

type configSource struct {
	// Pair is the pair of the source in the form of "base/quote".
	Pair string `hcl:",label"`

	// Type is the type of the graph node:
	// - "origin" for an origin node, that provides a price
	// - "median" for a median node, that calculates a median price from multiple sources
	// - "indirect" for an indirect node, that calculates an indirect price from multiple sources
	Type string `hcl:",label"`

	// Sources is a list of sources for "median" and "indirect" nodes.
	Sources []configSource `hcl:"source,block"`

	Body hcl.Body `hcl:",remain"` // To handle configOriginNode and configMedianNode.
}

type configOriginNode struct {
	// Origin is the name of the origin.
	Origin string `hcl:"origin"`
}

type configMedianNode struct {
	// MinSources is the minimum number of sources required to calculate a median price.
	MinSources int `hcl:"min_sources"`
}

type configHook struct {
	// Pair is the pair of the hook in the form of "base/quote".
	Pair string `hcl:",label"`

	// PostPriceHook is the configuration of the post price hook.
	PostPriceHook cty.Value `hcl:"post_price,optional"`
}

// AsyncGofer returns a new async gofer instance.
func (c *ConfigGofer) AsyncGofer(d AsyncDependencies) (provider.Provider, error) {
	if c.asyncPriceProvider != nil {
		return c.asyncPriceProvider, nil
	}
	gra, err := c.buildGraphs()
	if err != nil {
		return nil, fmt.Errorf("gofer config: unable to build graphs: %w", err)
	}
	var ns []nodes.Node
	for _, n := range gra {
		ns = append(ns, n)
	}
	originSet, err := c.buildOrigins(d.Clients)
	if err != nil {
		return nil, err
	}
	feed := feeder.NewFeeder(originSet, d.Logger)
	asyncProvider, err := graph.NewAsyncProvider(gra, feed, ns, d.Logger)
	if err != nil {
		return nil, fmt.Errorf("gofer config: unable to initialize RPC client: %w", err)
	}
	c.asyncPriceProvider = asyncProvider
	return asyncProvider, nil
}

// RPCAgent returns a new rpc.Agent instance.
func (c *ConfigGofer) RPCAgent(d AgentDependencies) (*rpc.Agent, error) {
	if c.rpcAgent != nil {
		return c.rpcAgent, nil
	}
	rpcAgent, err := rpc.NewAgent(rpc.AgentConfig{
		Provider: d.Provider,
		Network:  "tcp",
		Address:  c.RPCListenAddr,
		Logger:   d.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("gofer config: unable to initialize RPC agent: %w", err)
	}
	c.rpcAgent = rpcAgent
	return rpcAgent, nil
}

func (c *ConfigGofer) PriceHook(d HookDependencies) (provider.PriceHook, error) {
	if c.priceHook != nil {
		return c.priceHook, nil
	}
	params := provider.NewHookParams()
	for _, hook := range c.Hooks {
		v, err := ctyToAny(hook.PostPriceHook)
		if err != nil {
			return nil, fmt.Errorf("gofer config: invalid hook params: %w", err)
		}
		m, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("gofer config: invalid hook params: %v", v)
		}
		if len(m) > 0 {
			params[hook.Pair] = m
		}
	}
	priceHook, err := provider.NewPostPriceHook(d.Context, d.Clients, params)
	if err != nil {
		return nil, fmt.Errorf("gofer config: unable to initialize price hook: %w", err)
	}
	return priceHook, nil
}

// Gofer returns a new async gofer instance.
func (c *ConfigGofer) Gofer(d Dependencies, noRPC bool) (provider.Provider, error) {
	if c.priceProvider != nil {
		return c.priceProvider, nil
	}
	var err error
	if c.RPCAgentAddr == "" || noRPC {
		pricesGraph, err := c.buildGraphs()
		if err != nil {
			return nil, fmt.Errorf("gofer config: unable to load price models: %w", err)
		}
		originSet, err := c.buildOrigins(d.Clients)
		if err != nil {
			return nil, err
		}
		feed := feeder.NewFeeder(originSet, d.Logger)
		c.priceProvider = graph.NewProvider(pricesGraph, feed)
	} else {
		c.priceProvider, err = c.rpcClient(c.RPCAgentAddr)
		if err != nil {
			return nil, fmt.Errorf("gofer config: unable to initialize RPC client: %w", err)
		}
	}
	return c.priceProvider, nil
}

// configureRPCClient returns a new rpc.RPC instance.
func (c *ConfigGofer) rpcClient(listenAddr string) (*rpc.Provider, error) {
	return rpc.NewProvider("tcp", listenAddr)
}

func (c *ConfigGofer) buildOrigins(clients ethereumConfig.ClientRegistry) (*origins.Set, error) {
	const defaultWorkerCount = 10
	wp := query.NewHTTPWorkerPool(defaultWorkerCount)
	originSet := origins.DefaultOriginSet(wp)
	for _, origin := range c.Origins {
		params, err := ctyToAny(origin.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to parse params for origin %s: %w", origin.Origin, err)
		}
		handler, err := NewHandler(origin.Type, wp, clients, params)
		if err != nil || handler == nil {
			return nil, fmt.Errorf("failed to create handler for origin %s: %w", origin.Origin, err)
		}
		originSet.SetHandler(origin.Origin, handler)
	}
	return originSet, nil
}

func (c *ConfigGofer) buildGraphs() (map[provider.Pair]nodes.Node, error) {
	var err error
	graphs := map[provider.Pair]nodes.Node{}
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

func (c *ConfigGofer) buildRoots(graphs map[provider.Pair]nodes.Node) error {
	for _, model := range c.PriceModels {
		modelPair, err := provider.NewPair(model.Pair)
		if err != nil {
			return err
		}
		graphs[modelPair] = nodes.NewReferenceNode()
	}
	return nil
}

func (c *ConfigGofer) buildBranches(graphs map[provider.Pair]nodes.Node) error {
	for _, model := range c.PriceModels {
		modelPair, err := provider.NewPair(model.Pair)
		if err != nil {
			return err
		}
		node, err := c.buildNodes(model, graphs)
		if err != nil {
			return err
		}
		graphs[modelPair].(*nodes.ReferenceNode).SetReference(node)
	}
	return nil
}

func (c *ConfigGofer) buildNodes(config configSource, graphs map[provider.Pair]nodes.Node) (nodes.Node, error) {
	pair, err := provider.NewPair(config.Pair)
	if err != nil {
		return nil, err
	}
	switch config.Type {
	case "origin":
		return c.originNode(pair, config, graphs)
	case "median":
		return c.medianNode(pair, config, graphs)
	case "indirect":
		return c.indirectNode(pair, config, graphs)
	}
	return nil, fmt.Errorf("unknown node type: %s", config.Type)
}

func (c *ConfigGofer) childNodes(sources []configSource, graphs map[provider.Pair]nodes.Node) ([]nodes.Node, error) {
	var child []nodes.Node
	for _, source := range sources {
		node, err := c.buildNodes(source, graphs)
		if err != nil {
			return nil, err
		}
		child = append(child, node)
	}
	return child, nil
}

func (c *ConfigGofer) originNode(pair provider.Pair, config configSource, graphs graph.Graphs) (nodes.Node, error) {
	var params configOriginNode
	if err := decodeRemain(config.Body, &params); err != nil {
		return nil, err
	}
	if params.Origin == "." {
		return c.reference(pair, graphs)
	}
	originPair := nodes.OriginPair{
		Origin: params.Origin,
		Pair:   pair,
	}
	return nodes.NewOriginNode(originPair, defaultTTL, defaultTTL+maxTTL), nil
}

func (c *ConfigGofer) medianNode(pair provider.Pair, config configSource, graphs graph.Graphs) (nodes.Node, error) {
	child, err := c.childNodes(config.Sources, graphs)
	if err != nil {
		return nil, err
	}
	switch len(child) {
	case 0:
		return nil, fmt.Errorf("median aggregator must have at least one child")
	case 1:
		return child[0], nil
	default:
		var params configMedianNode
		if err := decodeRemain(config.Body, &params); err != nil {
			return nil, err
		}
		aggregator := nodes.NewMedianAggregatorNode(pair, params.MinSources)
		for _, c := range child {
			aggregator.AddChild(c)
		}
		return aggregator, nil
	}
}

func (c *ConfigGofer) indirectNode(pair provider.Pair, config configSource, graphs graph.Graphs) (nodes.Node, error) {
	child, err := c.childNodes(config.Sources, graphs)
	if err != nil {
		return nil, err
	}
	switch len(child) {
	case 0:
		return nil, fmt.Errorf("indirect aggregator must have at least one child")
	case 1:
		return child[0], nil
	default:
		aggregator := nodes.NewIndirectAggregatorNode(pair)
		for _, c := range child {
			aggregator.AddChild(c)
		}
		return aggregator, nil
	}
}

func (c *ConfigGofer) reference(pair provider.Pair, graphs graph.Graphs) (nodes.Node, error) {
	if _, ok := graphs[pair]; !ok {
		return nil, fmt.Errorf(
			"unable to find price model for the %s pair",
			pair,
		)
	}
	return graphs[pair], nil
}

func (c *ConfigGofer) detectCycle(graphs map[provider.Pair]nodes.Node) error {
	for _, pair := range sortGraphs(graphs) {
		if path := nodes.DetectCycle(graphs[pair]); len(path) > 0 {
			return ErrCyclicReference{Pair: pair, Path: path}
		}
	}
	return nil
}

func sortGraphs(graphs map[provider.Pair]nodes.Node) []provider.Pair {
	var ps []provider.Pair
	for p := range graphs {
		ps = append(ps, p)
	}
	sort.SliceStable(ps, func(i, j int) bool {
		return ps[i].String() < ps[j].String()
	})
	return ps
}

func decodeRemain(body hcl.Body, target any) error {
	diag := gohcl.DecodeBody(body, config.HCLContext, target)
	if diag.HasErrors() {
		return diag
	}
	return nil
}

func ctyToAny(v cty.Value) (any, error) {
	var err error
	typ := v.Type()
	switch {
	case typ.IsMapType() || typ.IsObjectType():
		m := make(map[string]any)
		for it := v.ElementIterator(); it.Next(); {
			ctyKey, ctyVal := it.Element()
			if ctyKey.Type() != cty.String {
				return nil, fmt.Errorf("unsupported type: %s", ctyKey.Type().FriendlyName())
			}
			m[ctyKey.AsString()], err = ctyToAny(ctyVal)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	case typ == cty.String:
		return v.AsString(), nil
	case typ == cty.Number:
		return v.AsBigFloat(), nil
	case typ == cty.Bool:
		return v.True(), nil
	case typ == cty.NilType:
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported type: %s", typ.FriendlyName())
}
