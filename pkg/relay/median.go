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
	"strings"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/multicall"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type medianWorker struct {
	contract       MedianContract
	dataPointStore store.DataPointProvider
	feedAddresses  []types.Address
	dataModel      string
	spread         float64
	expiration     time.Duration
	ticker         *timeutil.Ticker
	log            log.Logger
}

func (w *medianWorker) client() rpc.RPC {
	return w.contract.Client()
}

func (w *medianWorker) address() types.Address {
	return w.contract.Address()
}

func (w *medianWorker) createRelayCall(ctx context.Context) (gasEstimate uint64, call contract.Callable) {
	wat, val, age, bar, err := w.currentState(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to call Median contract")
		return 0, nil
	}
	if wat != w.dataModel {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("This is a bug in the configuration, probably a wrong contract address is used").
			Error("Contract asset name does not match the configured asset name")
		return 0, nil
	}

	// Load data points from the store.
	dataPoints, signatures, ok := w.findDataPoints(ctx, age, bar)
	if !ok {
		return 0, nil
	}

	prices := dataPointsToPrices(dataPoints)
	median := calculateMedian(prices)
	spread := calculateSpread(median, val.DecFloatPoint())

	// Check if price on the Median contract needs to be updated.
	// The price needs to be updated if:
	// - Price is older than the interval specified in the expiration field.
	// - Price differs from the current price by more than is specified in the
	//   Spread field.
	isExpired := time.Since(age) >= w.expiration
	isStale := math.IsInf(spread, 0) || spread >= w.spread

	// Print logs.
	w.log.
		WithFields(w.logFields()).
		WithFields(log.Fields{
			"bar":              bar,
			"age":              age,
			"val":              val,
			"expired":          isExpired,
			"stale":            isStale,
			"expiration":       w.expiration,
			"spread":           w.spread,
			"timeToExpiration": time.Since(age).String(),
			"currentSpread":    spread,
		}).
		Debug("Median worker")

	// If price is stale or expired, send update.
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
			w.handlePokeErr(err)
			return 0, nil
		}

		return gas, poke
	}

	return 0, nil
}

func (w *medianWorker) currentState(ctx context.Context) (wat string, val *bn.DecFixedPointNumber, age time.Time, bar int, err error) {
	val, err = w.contract.Val(ctx)
	if err != nil {
		return "", nil, time.Time{}, 0, err
	}
	if err := multicall.AggregateCallables(
		w.contract.Client(),
		w.contract.Wat(),
		w.contract.Age(),
		w.contract.Bar(),
	).Call(ctx, types.LatestBlockNumber, []any{
		&wat,
		&age,
		&bar,
	}); err != nil {
		return "", nil, time.Time{}, 0, err
	}
	return wat, val, age, bar, nil
}

func (w *medianWorker) findDataPoints(ctx context.Context, after time.Time, quorum int) ([]datapoint.Point, []types.Signature, bool) {
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

func (w *medianWorker) handlePokeErr(err error) {
	if strings.Contains(err.Error(), "replacement transaction underpriced") {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("This is expected during large price movements; the relay tries to update multiple contracts at once").
			Warn("Failed to poke the Median contract; previous transaction is still pending")
		return
	}
	/*
		if contract.IsRevert(err) {
			w.log.
				WithError(err).
				WithFields(w.logFields()).
				WithAdvice("Probably caused by a race condition between multiple relays; if this is a case, no action is required").
				Error("Failed to poke the Median contract")
			return
		}
	*/
	w.log.
		WithError(err).
		WithFields(w.logFields()).
		WithAdvice("Ignore if it is related to temporary network issues").
		Error("Failed to poke the Median contract")
}

func (w *medianWorker) logFields() log.Fields {
	return log.Fields{
		"address":   w.contract.Address(),
		"dataModel": w.dataModel,
	}
}
