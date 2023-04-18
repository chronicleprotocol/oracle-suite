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

type Provider struct {
	models  map[string]Node
	updater *Updater
}

func NewProvider(models map[string]Node, updater *Updater) Provider {
	return Provider{
		models:  models,
		updater: updater,
	}
}

func (p Provider) ModelNames(ctx context.Context) []string {
	models := maputil.Keys(p.models)
	sort.Strings(models)
	return models
}

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

func (p Provider) Model(ctx context.Context, model string) (*provider.Model, error) {
	panic("not implemented")
}

func (p Provider) Models(ctx context.Context, models ...string) (map[string]*provider.Model, error) {
	panic("not implemented")
}
