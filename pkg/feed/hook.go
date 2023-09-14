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

package feed

import (
	"context"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
)

// TickPrecisionHook is a hook that adjusts the precision of price and volume
// values in data points that hold a value.Tick.
type TickPrecisionHook struct {
	pricePrec  uint8
	volumePrec uint8
}

// NewTickPrecisionHook creates a new TickPrecisionHook with the specified
// price and volume precisions.
func NewTickPrecisionHook(pricePrec, volumePrec uint8) *TickPrecisionHook {
	return &TickPrecisionHook{
		pricePrec:  pricePrec,
		volumePrec: volumePrec,
	}
}

// BeforeSign implements the Hook interface.
func (t *TickPrecisionHook) BeforeSign(_ context.Context, dp *datapoint.Point) error {
	*dp = adjustPrec(*dp, t.pricePrec, t.volumePrec)
	return nil
}

func adjustPrec(dp datapoint.Point, pricePrec, volumePrec uint8) datapoint.Point {
	tick, ok := dp.Value.(value.Tick)
	if !ok {
		return dp
	}
	if tick.Price != nil {
		tick.Price = tick.Price.SetPrec(pricePrec)
	}
	if tick.Volume24h != nil {
		tick.Volume24h = tick.Volume24h.SetPrec(volumePrec)
	}
	dp.Value = tick
	for i, subPoint := range dp.SubPoints {
		dp.SubPoints[i] = adjustPrec(subPoint, pricePrec, volumePrec)
	}
	return dp
}

// BeforeBroadcast implements the Hook interface.
func (t *TickPrecisionHook) BeforeBroadcast(_ context.Context, _ *datapoint.Point) error {
	return nil
}

// TickTraceHook is a hook that adds a trace meta field that holds the price of
// the tick at each origin.
type TickTraceHook struct{}

// NewTickTraceHook creates a new TickTraceHook instance.
func NewTickTraceHook() *TickTraceHook {
	return &TickTraceHook{}
}

// BeforeSign implements the Hook interface.
func (t *TickTraceHook) BeforeSign(_ context.Context, _ *datapoint.Point) error {
	return nil
}

// BeforeBroadcast implements the Hook interface.
func (t *TickTraceHook) BeforeBroadcast(_ context.Context, dp *datapoint.Point) error {
	trace := buildTraceMap(*dp)
	if len(trace) > 0 {
		dp.Meta["trace"] = trace
	}
	return nil
}

func buildTraceMap(dp datapoint.Point) map[string]string {
	trace := make(map[string]string)
	var recur func(dp datapoint.Point) []datapoint.Point
	recur = func(dp datapoint.Point) []datapoint.Point {
		var points []datapoint.Point
		if dp.Meta["type"] == "origin" {
			points = append(points, dp)
		}
		for _, subPoint := range dp.SubPoints {
			points = append(points, recur(subPoint)...)
		}
		return points
	}

	for _, point := range recur(dp) {
		tick, ok := point.Value.(value.Tick)
		if !ok || point.Meta == nil || tick.Price == nil {
			continue
		}
		trace[fmt.Sprintf("%s@%s", tick.Pair.String(), point.Meta["origin"])] = tick.Price.String()
	}
	return trace
}
