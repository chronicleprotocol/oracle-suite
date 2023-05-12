package graph

import (
	"context"
	"fmt"
	"sort"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
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

// NewProvider creates a new price data.
//
// Models are map of data models graphs keyed by their data model name.
//
// Updater is an optional updater which will be used to update the data models
// before returning the data point.
func NewProvider(models map[string]Node, updater *Updater) Provider {
	return Provider{
		models:  models,
		updater: updater,
	}
}

// ModelNames implements the data.Provider interface.
func (p Provider) ModelNames(_ context.Context) []string {
	return maputil.SortKeys(p.models, sort.Strings)
}

// DataPoint implements the data.Provider interface.
func (p Provider) DataPoint(ctx context.Context, model string) (data.Point, error) {
	node, ok := p.models[model]
	if !ok {
		return data.Point{}, ErrModelNotFound{model: model}
	}
	if p.updater != nil {
		if err := p.updater.Update(ctx, []Node{node}); err != nil {
			return data.Point{}, err
		}
	}
	return node.DataPoint(), nil
}

// DataPoints implements the data.Provider interface.
func (p Provider) DataPoints(ctx context.Context, models ...string) (map[string]data.Point, error) {
	nodes := make([]Node, len(models))
	for i, model := range models {
		node, ok := p.models[model]
		if !ok {
			return nil, ErrModelNotFound{model: model}
		}
		nodes[i] = node
	}
	if p.updater != nil {
		if err := p.updater.Update(ctx, nodes); err != nil {
			return nil, err
		}
	}
	points := make(map[string]data.Point, len(models))
	for i, model := range models {
		points[model] = nodes[i].DataPoint()
	}
	return points, nil
}

// Model implements the data.Provider interface.
func (p Provider) Model(_ context.Context, model string) (data.Model, error) {
	node, ok := p.models[model]
	if !ok {
		return data.Model{}, ErrModelNotFound{model: model}
	}
	return nodeToModel(node), nil
}

// Models implements the data.Provider interface.
func (p Provider) Models(_ context.Context, models ...string) (map[string]data.Model, error) {
	nodes := make([]Node, len(models))
	for i, model := range models {
		node, ok := p.models[model]
		if !ok {
			return nil, ErrModelNotFound{model: model}
		}
		nodes[i] = node
	}
	modelsMap := make(map[string]data.Model, len(models))
	for i, model := range models {
		modelsMap[model] = nodeToModel(nodes[i])
	}
	return modelsMap, nil
}

func nodeToModel(n Node) data.Model {
	m := data.Model{}
	m.Meta = n.Meta()
	for _, n := range n.Nodes() {
		m.Models = append(m.Models, nodeToModel(n))
	}
	if m.Meta == nil {
		m.Meta = map[string]any{}
	}
	return m
}
