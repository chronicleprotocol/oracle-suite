package graph

import (
	"context"
	"fmt"
	"sort"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

type ErrModelNotFound struct {
	model string
}

func (e ErrModelNotFound) Error() string {
	return fmt.Sprintf("model %s not found", e.model)
}

// Provider is a price provider which uses a graph to calculate prices.
type Provider struct {
	models  map[string]Node
	updater *Updater
}

// NewProvider creates a new price provider.
func NewProvider(models map[string]Node, updater *Updater) Provider {
	return Provider{
		models:  models,
		updater: updater,
	}
}

// ModelNames implements the provider.Provider interface.
func (p Provider) ModelNames(_ context.Context) []string {
	return maputil.SortKeys(p.models, sort.Strings)
}

// Tick implements the provider.Provider interface.
func (p Provider) Tick(ctx context.Context, model string) (provider.Tick, error) {
	node, ok := p.models[model]
	if !ok {
		return provider.Tick{}, ErrModelNotFound{model: model}
	}
	if err := p.updater.Update(ctx, []Node{node}); err != nil {
		return provider.Tick{}, err
	}
	return node.Tick(), nil
}

// Ticks implements the provider.Provider interface.
func (p Provider) Ticks(ctx context.Context, models ...string) (map[string]provider.Tick, error) {
	nodes := make([]Node, len(models))
	for i, model := range models {
		node, ok := p.models[model]
		if !ok {
			return nil, ErrModelNotFound{model: model}
		}
		nodes[i] = node
	}
	if err := p.updater.Update(ctx, nodes); err != nil {
		return nil, err
	}
	ticks := make(map[string]provider.Tick, len(models))
	for i, model := range models {
		ticks[model] = nodes[i].Tick()
	}
	return ticks, nil
}

// Model implements the provider.Provider interface.
func (p Provider) Model(_ context.Context, model string) (provider.Model, error) {
	node, ok := p.models[model]
	if !ok {
		return provider.Model{}, ErrModelNotFound{model: model}
	}
	return nodeToModel(node), nil
}

// Models implements the provider.Provider interface.
func (p Provider) Models(_ context.Context, models ...string) (map[string]provider.Model, error) {
	nodes := make([]Node, len(models))
	for i, model := range models {
		node, ok := p.models[model]
		if !ok {
			return nil, ErrModelNotFound{model: model}
		}
		nodes[i] = node
	}
	modelsMap := make(map[string]provider.Model, len(models))
	for i, model := range models {
		modelsMap[model] = nodeToModel(nodes[i])
	}
	return modelsMap, nil
}

func nodeToModel(n Node) provider.Model {
	m := provider.Model{}
	m.Pair = n.Pair()
	m.Meta = n.Meta()
	for _, n := range n.Branches() {
		m.Models = append(m.Models, nodeToModel(n))
	}
	if m.Meta == nil {
		m.Meta = MapMeta{}
	}
	return m
}
