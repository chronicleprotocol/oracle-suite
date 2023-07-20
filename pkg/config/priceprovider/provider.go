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

package priceprovider

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/feed"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/nodes"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/origins"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

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

type Config struct {
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

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured services:
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
	Params map[string]any `hcl:"params,optional"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type configSource struct {
	// Pair is the pair of the source in the form of "base/quote".
	Pair provider.Pair `hcl:",label"`

	// Type is the type of the graph node:
	// - "origin" for an origin node, that provides a price
	// - "median" for a median node, that calculates a median price from multiple sources
	// - "indirect" for an indirect node, that calculates an indirect price from multiple sources
	Type string `hcl:",label"`

	// Sources is a list of sources for "median" and "indirect" nodes.
	Sources []configSource `hcl:"source,block"`

	// Type specific configuration:
	Origin *configOriginNode
	Median *configMedianNode

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
	Body    hcl.Body        `hcl:",remain"` // To handle configOriginNode and configMedianNode.
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
	Pair provider.Pair `hcl:",label"`

	// PostPriceHook is the configuration of the post price hook.
	PostPriceHook map[string]any `hcl:"post_price,optional"`
}

// PostDecodeBlock implements the hcl.PostDecodeBlock interface.
// It is used to decode type specific configuration in source blocks.
func (c *configSource) PostDecodeBlock(
	ctx *hcl.EvalContext,
	_ *hcl.BodySchema,
	_ *hcl.Block,
	_ *hcl.BodyContent) hcl.Diagnostics {

	switch c.Type {
	case "origin":
		c.Origin = &configOriginNode{}
		return utilHCL.Decode(ctx, c.Body, c.Origin)
	case "median":
		c.Median = &configMedianNode{}
		return utilHCL.Decode(ctx, c.Body, c.Median)
	}
	return nil
}

// AsyncPriceProvider returns a new async gofer instance.
func (c *Config) AsyncPriceProvider(d AsyncDependencies) (provider.Provider, error) {
	if c.asyncPriceProvider != nil {
		return c.asyncPriceProvider, nil
	}
	graphs, err := c.buildGraphs()
	if err != nil {
		return nil, err
	}
	originSet, err := c.buildOrigins(d.Clients)
	if err != nil {
		return nil, err
	}
	feed := feed.NewFeed(originSet, d.Logger)
	asyncProvider, err := graph.NewAsyncProvider(graphs, feed, d.Logger)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Async Price Provider service: %v", err),
			Subject:  c.Range.Ptr(),
		}
	}
	c.asyncPriceProvider = asyncProvider
	return asyncProvider, nil
}

type Provider struct {
	p provider.Provider
}

func (p *Provider) ModelNames(_ context.Context) []string {
	models, err := p.p.Models()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(models))
	for name := range models {
		names = append(names, name.String())
	}
	return names
}

func (p *Provider) DataPoint(_ context.Context, model string) (datapoint.Point, error) {
	pair, err := provider.NewPair(model)
	if err != nil {
		return datapoint.Point{}, err
	}

	price, err := p.p.Price(pair)
	if err != nil {
		return datapoint.Point{}, err
	}

	return p2p(*price), nil
}

func p2p(price provider.Price) datapoint.Point {
	meta := make(map[string]interface{}, 0)
	for k, v := range price.Parameters {
		meta[k] = v
	}
	points := make([]datapoint.Point, 0, len(price.Prices))
	for _, p := range price.Prices {
		points = append(points, p2p(*p))
	}
	return datapoint.Point{
		Value: value.Tick{
			Pair: value.Pair{
				Base:  price.Pair.Base,
				Quote: price.Pair.Quote,
			},
			Price: bn.Float(price.Price),
		},
		Time:      price.Time,
		SubPoints: points,
		Meta:      meta,
		Error:     fmt.Errorf("cannot create datapoint: %s", price.Error),
	}
}

func (p *Provider) DataPoints(_ context.Context, models ...string) (map[string]datapoint.Point, error) {
	pairs := make([]provider.Pair, 0, len(models))
	for _, model := range models {
		pair, err := provider.NewPair(model)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, pair)
	}

	prices, err := p.p.Prices(pairs...)
	if err != nil {
		return nil, err
	}

	points := make(map[string]datapoint.Point, len(prices))
	for _, price := range prices {
		points[price.Pair.String()] = p2p(*price)
	}
	return points, nil
}

func (p *Provider) Model(_ context.Context, model string) (datapoint.Model, error) {
	pair, err := provider.NewPair(model)
	if err != nil {
		return datapoint.Model{}, err
	}
	models, err := p.p.Models(pair)
	if err != nil {
		return datapoint.Model{}, err
	}
	return m2m(*models[pair]), nil
}

func m2m(m provider.Model) datapoint.Model {
	meta := map[string]interface{}{
		"type": m.Type,
	}
	for k, v := range m.Parameters {
		meta[k] = v
	}
	mds := make([]datapoint.Model, 0, len(m.Models))
	for _, mm := range m.Models {
		mds = append(mds, m2m(*mm))
	}
	return datapoint.Model{
		Meta:   meta,
		Models: mds,
	}
}

func (p *Provider) Models(_ context.Context, models ...string) (map[string]datapoint.Model, error) {
	pairs := make([]provider.Pair, 0, len(models))
	for _, model := range models {
		pair, err := provider.NewPair(model)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, pair)
	}

	mds, err := p.p.Models(pairs...)
	if err != nil {
		return nil, err
	}

	r := make(map[string]datapoint.Model, len(mds))
	for k, v := range mds {
		r[k.String()] = m2m(*v)
	}
	return r, nil
}

func (c *Config) ConfigureDataProvider(d Dependencies) (datapoint.Provider, error) {
	pp, err := c.PriceProvider(d, true)
	if err != nil {
		return nil, err
	}
	return &Provider{p: pp}, nil
}

// PriceProvider returns a new async gofer instance.
func (c *Config) PriceProvider(d Dependencies, noRPC bool) (provider.Provider, error) {
	if c.priceProvider != nil {
		return c.priceProvider, nil
	}
	var err error
	if c.RPCAgentAddr == "" || noRPC {
		pricesGraph, err := c.buildGraphs()
		if err != nil {
			return nil, err
		}
		originSet, err := c.buildOrigins(d.Clients)
		if err != nil {
			return nil, err
		}
		f := feed.NewFeed(originSet, d.Logger)
		c.priceProvider = graph.NewProvider(pricesGraph, f)
	} else {
		c.priceProvider, err = rpc.NewProvider("tcp", c.RPCAgentAddr)
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create the RPC client: %v", err),
				Subject:  c.Range.Ptr(),
			}
		}
	}
	return c.priceProvider, nil
}

// RPCAgent returns a new rpc.Agent instance.
func (c *Config) RPCAgent(d AgentDependencies) (*rpc.Agent, error) {
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
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the RPC agent: %v", err),
			Subject:  c.Range.Ptr(),
		}
	}
	c.rpcAgent = rpcAgent
	return rpcAgent, nil
}

func (c *Config) PriceHook(d HookDependencies) (provider.PriceHook, error) {
	if c.priceHook != nil {
		return c.priceHook, nil
	}
	params := provider.NewHookParams()
	for _, hook := range c.Hooks {
		if len(hook.PostPriceHook) > 0 {
			params[hook.Pair.String()] = hook.PostPriceHook
		}
	}
	priceHook, err := provider.NewPostPriceHook(d.Context, d.Clients, params)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create post price hook: %v", err),
			Subject:  c.Range.Ptr(),
		}
	}
	return priceHook, nil
}

func (c *Config) buildOrigins(clients ethereumConfig.ClientRegistry) (*origins.Set, error) {
	const defaultWorkerCount = 10
	wp := query.NewHTTPWorkerPool(defaultWorkerCount)
	originSet := origins.DefaultOriginSet(wp)
	for _, origin := range c.Origins {
		handler, err := NewHandler(origin.Type, wp, clients, origin.Params)
		if err != nil || handler == nil {
			return nil, fmt.Errorf("failed to create handler for origin %s: %w", origin.Origin, err)
		}
		originSet.SetHandler(origin.Origin, handler)
	}
	return originSet, nil
}

func (c *Config) buildGraphs() (map[provider.Pair]nodes.Node, error) {
	var err error
	graphs := map[provider.Pair]nodes.Node{}
	// It is important to create root nodes before branches, because branches
	// may refer to another root nodes instances.
	c.buildRoots(graphs)
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

func (c *Config) buildRoots(graphs map[provider.Pair]nodes.Node) {
	for _, model := range c.PriceModels {
		graphs[model.Pair] = nodes.NewReferenceNode()
	}
}

func (c *Config) buildBranches(graphs map[provider.Pair]nodes.Node) error {
	for _, model := range c.PriceModels {
		node, err := c.buildNodes(model, graphs)
		if err != nil {
			return err
		}
		graphs[model.Pair].(*nodes.ReferenceNode).SetReference(node)
	}
	return nil
}

func (c *Config) buildNodes(config configSource, graphs map[provider.Pair]nodes.Node) (nodes.Node, error) {
	switch config.Type {
	case "origin":
		return c.originNode(config, graphs)
	case "median":
		return c.medianNode(config, graphs)
	case "indirect":
		return c.indirectNode(config, graphs)
	}
	return nil, fmt.Errorf("unknown node type: %s", config.Type)
}

func (c *Config) childNodes(sources []configSource, graphs map[provider.Pair]nodes.Node) ([]nodes.Node, error) {
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

func (c *Config) originNode(config configSource, graphs graph.Graphs) (nodes.Node, error) {
	if config.Origin.Origin == "." {
		ref, ok := graphs[config.Pair]
		if !ok {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Reference to unknown pair %s", config.Pair),
				Subject:  config.Range.Ptr(),
			}
		}
		return ref, nil
	}
	originPair := nodes.OriginPair{
		Origin: config.Origin.Origin,
		Pair:   config.Pair,
	}
	return nodes.NewOriginNode(
		originPair,
		1*time.Minute, // TTL to update price
		5*time.Minute, // TTL to mark price as expired
	), nil
}

func (c *Config) medianNode(config configSource, graphs graph.Graphs) (nodes.Node, error) {
	child, err := c.childNodes(config.Sources, graphs)
	if err != nil {
		return nil, err
	}
	switch len(child) {
	case 0:
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "Median aggregator must have at least one child",
			Subject:  config.Range.Ptr(),
		}
	case 1:
		return child[0], nil
	default:
		aggregator := nodes.NewMedianAggregatorNode(config.Pair, config.Median.MinSources)
		for _, c := range child {
			aggregator.AddChild(c)
		}
		return aggregator, nil
	}
}

func (c *Config) indirectNode(config configSource, graphs graph.Graphs) (nodes.Node, error) {
	child, err := c.childNodes(config.Sources, graphs)
	if err != nil {
		return nil, err
	}
	switch len(child) {
	case 0:
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "Indirect aggregator must have at least one child",
			Subject:  config.Range.Ptr(),
		}
	case 1:
		return child[0], nil
	default:
		aggregator := nodes.NewIndirectAggregatorNode(config.Pair)
		for _, c := range child {
			aggregator.AddChild(c)
		}
		return aggregator, nil
	}
}

func (c *Config) detectCycle(graphs map[provider.Pair]nodes.Node) error {
	for _, pair := range sortGraphs(graphs) {
		if path := nodes.DetectCycle(graphs[pair]); len(path) > 0 {
			stringPath := make([]string, len(path))
			for i, n := range path {
				switch typedNode := n.(type) {
				case nodes.Aggregator:
					stringPath[i] = typedNode.Pair().String()
				case nodes.Origin:
					stringPath[i] = typedNode.OriginPair().String()
				}
			}
			pos := c.Range.Ptr()
			for _, model := range c.Content.Blocks.OfType("price_model") {
				if model.Labels[0] == pair.String() {
					pos = model.DefRange.Ptr()
					break
				}
			}
			return &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Cyclic reference in %s: %s", pair, strings.Join(stringPath, " -> ")),
				Subject:  pos,
			}
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
