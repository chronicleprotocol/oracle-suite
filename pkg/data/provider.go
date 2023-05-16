package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/treerender"
)

// Provider provides data points.
//
// A data point is a value obtained from an origin. Data point can be anything
// from an asset price to a string value.
//
// A model describes how a data point is calculated and obtained. For example,
// a model can describe from which origins data points are obtained and how
// they are combined to calculate a final value. Details of how models work
// depend on a specific implementation.
type Provider interface {
	// ModelNames returns a list of supported data models.
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

// Model is a simplified representation of a model which is used to obtain
// a data point. The main purpose of this structure is to help the end
// user to understand how data points values are calculated and obtained.
//
// This structure is purely informational. The way it is used depends on
// a specific implementation.
type Model struct {
	// Meta contains metadata for the model. It should contain information
	// about the model and its parameters.
	//
	// The "type" metadata field is used to determine the type of the model.
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
		for k, v := range model.Meta {
			if k == "type" {
				typ, _ = v.(string)
				continue
			}
			meta[k] = v
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
// an origin. It can be a price, a volume, a market cap, etc. The value
// itself is represented by the Value interface and can be anything, a number,
// a string, or even a complex structure.
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
	meta := make(map[string]any)
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
	for k, v := range p.Meta {
		meta["meta."+k] = v
	}
	return json.Marshal(meta)
}

// MarshalTrace returns a human-readable representation of the tick.
func (p Point) MarshalTrace() ([]byte, error) {
	return treerender.RenderTree(func(node any) treerender.NodeData {
		meta := make(map[string]any)
		point := node.(Point)
		typ := "data_point"
		meta["value"] = point.Value
		meta["time"] = point.Time.In(time.UTC).Format(time.RFC3339Nano)
		var ticks []any
		for _, t := range point.SubPoints {
			ticks = append(ticks, t)
		}
		for k, v := range point.Meta {
			if k == "type" {
				typ, _ = v.(string)
				continue
			}
			meta["meta."+k] = v
		}
		return treerender.NodeData{
			Name:      typ,
			Params:    meta,
			Ancestors: ticks,
			Error:     point.Validate(),
		}
	}, []any{p}, 0), nil
}

// Value is a interface for data point values.
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
