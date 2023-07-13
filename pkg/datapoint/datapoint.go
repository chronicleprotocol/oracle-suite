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

package datapoint

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/defiweb/go-eth/types"
	"google.golang.org/protobuf/proto"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/treerender"
)

// Provider provides data points.
//
// A data point is a value obtained from a source. For example, a data
// point can be a price of an asset at a specific time obtained from
// an exchange.
//
// A model describes how a data point is calculated and obtained. For example,
// a model can describe from which sources data points are obtained and how
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

// Signer is responsible for signing data points.
type Signer interface {
	// Supports returns true if the signer supports the given data point.
	Supports(ctx context.Context, data Point) bool

	// Sign signs a data point using the given key.
	Sign(ctx context.Context, model string, data Point) (*types.Signature, error)
}

// Recoverer is responsible for recovering addresses from signatures.
type Recoverer interface {
	// Supports returns true if the recoverer supports the given data point.
	Supports(ctx context.Context, data Point) bool

	// Recover recovers the address from the given signature.
	Recover(ctx context.Context, model string, data Point, signature types.Signature) (*types.Address, error)
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
	//
	// Meta values must be marshalable to JSON.
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
	Value value.Value

	// Time is the time when the data point was obtained.
	Time time.Time

	// SubPoints is a list of sub data points that are used to obtain this
	// data point.
	SubPoints []Point

	// Meta contains metadata for the data point. It may contain additional
	// information about the data point, such as the origin it was obtained
	// from, etc.
	//
	// Meta values must be marshalable to JSON.
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
	if v, ok := p.Value.(value.ValidatableValue); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	if p.Time.IsZero() {
		return fmt.Errorf("time is not set")
	}
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p Point) MarshalBinary() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	value, err := value.MarshalBinary(p.Value)
	if err != nil {
		return nil, err
	}
	subPoints := make([][]byte, len(p.SubPoints))
	for i, t := range p.SubPoints {
		subPoints[i], err = t.MarshalBinary()
		if err != nil {
			return nil, err
		}
	}
	meta := make(map[string][]byte)
	for k, v := range p.Meta {
		meta[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}
	return proto.Marshal(&pb.DataPoint{
		Value:     value,
		Timestamp: p.Time.Unix(),
		SubPoints: subPoints,
		Meta:      meta,
	})
}

func (p *Point) UnmarshalBinary(data []byte) error {
	var pbDataPoint pb.DataPoint
	if err := proto.Unmarshal(data, &pbDataPoint); err != nil {
		return err
	}
	value, err := value.UnmarshalBinary(pbDataPoint.Value)
	if err != nil {
		return err
	}
	p.Value = value
	p.Time = time.Unix(pbDataPoint.Timestamp, 0)
	p.SubPoints = make([]Point, len(pbDataPoint.SubPoints))
	for i, t := range pbDataPoint.SubPoints {
		if err := p.SubPoints[i].UnmarshalBinary(t); err != nil {
			return err
		}
	}
	p.Meta = make(map[string]any)
	for k, v := range pbDataPoint.Meta {
		var value any
		if err := json.Unmarshal(v, &value); err != nil {
			return err
		}
		p.Meta[k] = value
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
