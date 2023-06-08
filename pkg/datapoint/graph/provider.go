package graph

import (
	"context"
	"fmt"
	"sort"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

type ErrModelNotFound struct {
	model string
}

func (e ErrModelNotFound) Error() string {
	return fmt.Sprintf("model %s not found", e.model)
}

// Provider is a data provider which uses a graph structure to provide data
// points.
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
func (p Provider) DataPoint(ctx context.Context, model string) (datapoint.Point, error) {
	node, ok := p.models[model]
	if !ok {
		return datapoint.Point{}, ErrModelNotFound{model: model}
	}
	if p.updater != nil {
		if err := p.updater.Update(ctx, []Node{node}); err != nil {
			return datapoint.Point{}, err
		}
	}
	return node.DataPoint(), nil
}

// DataPoints implements the data.Provider interface.
func (p Provider) DataPoints(ctx context.Context, models ...string) (map[string]datapoint.Point, error) {
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
	points := make(map[string]datapoint.Point, len(models))
	for i, model := range models {
		points[model] = nodes[i].DataPoint()
	}
	return points, nil
}

// Model implements the data.Provider interface.
func (p Provider) Model(_ context.Context, model string) (datapoint.Model, error) {
	node, ok := p.models[model]
	if !ok {
		return datapoint.Model{}, ErrModelNotFound{model: model}
	}
	return nodeToModel(node), nil
}

// Models implements the data.Provider interface.
func (p Provider) Models(_ context.Context, models ...string) (map[string]datapoint.Model, error) {
	nodes := make([]Node, len(models))
	for i, model := range models {
		node, ok := p.models[model]
		if !ok {
			return nil, ErrModelNotFound{model: model}
		}
		nodes[i] = node
	}
	modelsMap := make(map[string]datapoint.Model, len(models))
	for i, model := range models {
		modelsMap[model] = nodeToModel(nodes[i])
	}
	return modelsMap, nil
}

func nodeToModel(n Node) datapoint.Model {
	m := datapoint.Model{}
	m.Meta = n.Meta()
	for _, n := range n.Nodes() {
		m.Models = append(m.Models, nodeToModel(n))
	}
	if m.Meta == nil {
		m.Meta = map[string]any{}
	}
	return m
}
