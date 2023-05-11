package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/treerender"
)

// Provider provides prices for asset pairs.
type Provider interface {
	// ModelNames returns a list of supported price models.
	ModelNames(ctx context.Context) []string

	// DataPoint returns a data point for the given model.
	DataPoint(ctx context.Context, model string) (Point, error)

	// DataPoints returns a map of data points for the given models.
	DataPoints(ctx context.Context, models ...string) (map[string]Point, error)

	// Model returns a price model for the given asset pair.
	Model(ctx context.Context, model string) (Model, error)

	// Models describes price models which are used to calculate prices.
	// If no pairs are specified, models for all pairs are returned.
	Models(ctx context.Context, models ...string) (map[string]Model, error)
}

// Model is a simplified representation of a model which is used to calculate
// asset pair prices. The main purpose of this structure is to help the end
// user to understand how prices are derived and calculated.
//
// This structure is purely informational. The way it is used depends on
// a specific implementation.
type Model struct {
	// Meta contains metadata for the model. It should contain information
	// about the model and its parameters.
	//
	// Following keys are reserved:
	// - models: a list of sub models
	Meta map[string]any

	// Models is a list of sub models used to calculate price.
	Models []Model
}

// MarshalJSON implements the json.Marshaler interface.
func (m Model) MarshalJSON() ([]byte, error) {
	meta := m.Meta
	meta["models"] = m.Models
	return json.Marshal(meta)
}

// MarshalTrace returns a human-readable representation of the model.
func (m Model) MarshalTrace() ([]byte, error) {
	return treerender.RenderTree(func(node any) treerender.NodeData {
		meta := map[string]any{}
		model := node.(Model)
		typ := "node"
		if model.Meta != nil {
			meta = model.Meta
		}
		if n, ok := meta["type"].(string); ok {
			typ = n
			delete(meta, "type")
		}
		var models []any
		for _, m := range model.Models {
			models = append(models, m)
		}
		return treerender.NodeData{
			Name:      typ,
			Params:    meta,
			Ancestors: models,
			Error:     nil,
		}
	}, []any{m}, 0), nil
}

// Point represents a data point. It can represent any value obtained from
// an origin. It can be a price, a volume, a market cap, etc.
//
// Before using this data, you should check if it is valid by calling
// Point.Validate() method.
type Point struct {
	// Value is the value of the data point.
	Value Value

	// Time is the time when the data point was obtained.
	Time time.Time

	// SubPoints is a list of sub data points that are used to obtain this
	// data point.
	SubPoints []Point

	// Meta contains metadata for the data point. It may contain additional
	// information about the data point, such as the origin it was obtained
	// from, etc.
	//
	// Following keys are reserved:
	// - value: the value of the data point
	// - time: the time when the data point was obtained
	// - ticks: a list of sub data points
	// - error: an error which occurred during obtaining the data point
	Meta map[string]any

	// Error is an optional error which occurred during obtaining the price.
	// If error is not nil, then the price is invalid and should not be used.
	//
	// Point may be invalid for other reasons, hence you should always check
	// the data point for validity by calling Point.Validate() method.
	Error error
}

// Validate returns an error if the data point is invalid.
func (p Point) Validate() error {
	if p.Error != nil {
		return p.Error
	}
	if p.Value == nil {
		return fmt.Errorf("value is not set")
	}
	if v, ok := p.Value.(ValidatableValue); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	if p.Time.IsZero() {
		return fmt.Errorf("time is not set")
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (p Point) MarshalJSON() ([]byte, error) {
	meta := p.Meta
	meta["value"] = p.Value
	meta["time"] = p.Time.In(time.UTC).Format(time.RFC3339Nano)
	var ticks []any
	for _, t := range p.SubPoints {
		ticks = append(ticks, t)
	}
	if len(ticks) > 0 {
		meta["ticks"] = ticks
	}
	if err := p.Validate(); err != nil {
		meta["error"] = err.Error()
	}
	return json.Marshal(meta)
}

// MarshalTrace returns a human-readable representation of the tick.
func (p Point) MarshalTrace() ([]byte, error) {
	return treerender.RenderTree(func(node any) treerender.NodeData {
		meta := make(map[string]any)
		point := node.(Point)
		err := point.Validate()
		if point.Meta != nil {
			meta = point.Meta
		}
		typ := "data_point"
		if n, ok := meta["type"].(string); ok {
			typ = n
			delete(meta, "type")
		}
		meta["time"] = point.Time.In(time.UTC).Format(time.RFC3339Nano)
		if point.Value != nil {
			meta["value"] = point.Value.Print()
		}
		var ticks []any
		for _, t := range point.SubPoints {
			ticks = append(ticks, t)
		}
		return treerender.NodeData{
			Name:      typ,
			Params:    meta,
			Ancestors: ticks,
			Error:     err,
		}
	}, []any{p}, 0), nil
}

// Value is a data point value.
type Value interface {
	// Print returns a human-readable representation of the value.
	Print() string
}

// ValidatableValue is a data point value which can be validated.
type ValidatableValue interface {
	Validate() error
}

// NumericValue is a data point value which is a number.
type NumericValue interface {
	Number() *bn.FloatNumber
}
