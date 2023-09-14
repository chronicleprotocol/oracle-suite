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

// TickPrecisionHook is a hook that limits the precision of the price and volume
// of a tick.
//
// This is done for two reasons:
//   - To decrease the size of the tick, which is broadcast to all oracles.
//   - To mitigate the risk of rounding errors that might occur during
//     generating a hash for signature verification if the precision is too
//     high. This isn't a concern as long as signatures are verified using the
//     same library that calculated the hash. However, discrepancies may arise
//     if a different library is used.
type TickPrecisionHook struct {
	maxPricePrec  uint8
	maxVolumePrec uint8
}

// NewTickPrecisionHook creates a new TickPrecisionHook with the specified
// price and volume precisions.
func NewTickPrecisionHook(maxPricePrec, maxVolumePrec uint8) *TickPrecisionHook {
	return &TickPrecisionHook{
		maxPricePrec:  maxPricePrec,
		maxVolumePrec: maxVolumePrec,
	}
}

// BeforeSign implements the Hook interface.
func (t *TickPrecisionHook) BeforeSign(_ context.Context, dp *datapoint.Point) error {
	*dp = adjustPrec(*dp, t.maxPricePrec, t.maxVolumePrec)
	return nil
}

func adjustPrec(dp datapoint.Point, pricePrec, volumePrec uint8) datapoint.Point {
	tick, ok := dp.Value.(value.Tick)
	if !ok {
		return dp
	}
	if tick.Price != nil && tick.Price.Prec() > pricePrec {
		tick.Price = tick.Price.SetPrec(pricePrec)
	}
	if tick.Volume24h != nil && tick.Volume24h.Prec() > volumePrec {
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

// TickTraceHook is a hook that adds a trace meta field, which stores the price of
// the tick at each origin.
//
// The trace field is a map that associates each pair and origin with a price.
// This allows to easily verify by others that the broadcasted tick is accurately
// calculated by the feed.
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
	// Find all origin data points.
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

	// Build trace map.
	// Format: map[<pair>@<origin>] = <price>
	trace := make(map[string]string)
	for _, point := range recur(dp) {
		tick, ok := point.Value.(value.Tick)
		if !ok || point.Meta == nil || tick.Price == nil {
			continue
		}
		if tick.Validate() == nil {
			trace[fmt.Sprintf("%s@%s", tick.Pair.String(), point.Meta["origin"])] = tick.Price.String()
		}
	}

	return trace
}
