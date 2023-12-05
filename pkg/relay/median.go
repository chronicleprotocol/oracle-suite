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

package relay

import (
	"context"
	"math"
	"time"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/multicall"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type median struct {
	contract       MedianContract
	dataPointStore store.DataPointProvider
	feedAddresses  []types.Address
	dataModel      string
	spread         float64
	expiration     time.Duration
	log            log.Logger
}

type medianState struct {
	wat string
	val *bn.DecFixedPointNumber
	age time.Time
	bar int
}

func (w *median) createRelayCall(ctx context.Context) []relayCall {
	state, err := w.currentState(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to call Median contract")
		return nil
	}
	if state.wat != w.dataModel {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("This is a bug in the configuration, probably a wrong contract address is used").
			Error("Contract asset name does not match the configured asset name")
		return nil
	}

	// Load data points from the store.
	dataPoints, signatures, ok := w.findDataPoints(ctx, state.age, state.bar)
	if !ok {
		return nil
	}

	prices := dataPointsToPrices(dataPoints)
	median := calculateMedian(prices)
	spread := calculateSpread(median, state.val.DecFloatPoint())

	// Check if price on the Median contract needs to be updated.
	// The price needs to be updated if:
	// - Price is older than the interval specified in the expiration field.
	// - Price differs from the current price by more than is specified in the
	//   spread field.
	isExpired := time.Since(state.age) >= w.expiration
	isStale := math.IsInf(spread, 0) || spread >= w.spread

	// Print logs.
	w.log.
		WithFields(w.logFields()).
		WithFields(log.Fields{
			"bar":              state.bar,
			"age":              state.age,
			"val":              state.val,
			"expired":          isExpired,
			"stale":            isStale,
			"expiration":       w.expiration,
			"spread":           w.spread,
			"timeToExpiration": time.Since(state.age).String(),
			"currentSpread":    spread,
		}).
		Debug("Median")

	// If price is stale or expired, return a poke transaction.
	if isExpired || isStale {
		vals := make([]chronicle.MedianVal, len(prices))
		for i := range dataPoints {
			vals[i] = chronicle.MedianVal{
				Val: prices[i].DecFixedPoint(chronicle.MedianPricePrecision),
				Age: dataPoints[i].Time,
				V:   uint8(signatures[i].V.Uint64()),
				R:   signatures[i].R,
				S:   signatures[i].S,
			}
		}

		poke := w.contract.Poke(vals)
		gas, err := poke.Gas(ctx, types.LatestBlockNumber)
		if err != nil {
			w.log.
				WithError(err).
				WithFields(w.logFields()).
				WithAdvice("Ignore if it is related to temporary network issues").
				Error("Failed to poke the Median contract")
			return nil
		}

		return []relayCall{{
			client:      w.contract.Client(),
			address:     w.contract.Address(),
			callable:    poke,
			gasEstimate: gas,
		}}
	}

	return nil
}

func (w *median) currentState(ctx context.Context) (state medianState, err error) {
	state.val, err = w.contract.Val(ctx)
	if err != nil {
		return medianState{}, err
	}
	if err := multicall.AggregateCallables(
		w.contract.Client(),
		w.contract.Wat(),
		w.contract.Age(),
		w.contract.Bar(),
	).Call(ctx, types.LatestBlockNumber, []any{
		&state.wat,
		&state.age,
		&state.bar,
	}); err != nil {
		return medianState{}, err
	}
	return state, nil
}

func (w *median) findDataPoints(ctx context.Context, after time.Time, quorum int) ([]datapoint.Point, []types.Signature, bool) {
	// Generate slice of random indices to select data points from.
	// It is important to select data points randomly to avoid promoting
	// any particular feed.
	randIndices, err := randomInts(len(w.feedAddresses))
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("This is a bug and needs to be investigated").
			Error("Failed to generate random indices")
		return nil, nil, false
	}

	// Try to get data points from the store from the feeds in the random order
	// until we get enough data points to satisfy the quorum.
	var dataPoints []datapoint.Point
	var signatures []types.Signature
	for _, i := range randIndices {
		sdp, ok, err := w.dataPointStore.LatestFrom(ctx, w.feedAddresses[i], w.dataModel)
		if err != nil {
			w.log.
				WithError(err).
				WithFields(w.logFields()).
				WithField("feedAddress", w.feedAddresses[i]).
				WithAdvice("Ignore if occurs occasionally").
				Warn("Failed to get data point")
			continue
		}
		if !ok {
			continue
		}
		if sdp.Signature.V == nil || sdp.Signature.R == nil || sdp.Signature.S == nil {
			continue
		}
		if _, ok := sdp.DataPoint.Value.(value.Tick); !ok {
			w.log.
				WithFields(w.logFields()).
				WithField("feedAddress", w.feedAddresses[i]).
				WithAdvice("This is probably caused by setting a wrong data model for this contract").
				Error("Data point is not a tick")
			continue
		}
		if sdp.DataPoint.Time.Before(after) {
			continue
		}
		dataPoints = append(dataPoints, sdp.DataPoint)
		signatures = append(signatures, sdp.Signature)
		if len(dataPoints) == quorum {
			break
		}
	}
	if len(dataPoints) != quorum {
		w.log.
			WithFields(w.logFields()).
			WithFields(log.Fields{
				"quorum": quorum,
				"found":  len(dataPoints),
			}).
			WithAdvice("Ignore if occurs during the first few minutes after the start of the relay").
			Warn("Unable to obtain enough data points")
		return nil, nil, false
	}

	return dataPoints, signatures, true
}

func (w *median) logFields() log.Fields {
	return log.Fields{
		"address":   w.contract.Address(),
		"dataModel": w.dataModel,
	}
}
