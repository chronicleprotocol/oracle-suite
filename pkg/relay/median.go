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
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type medianWorker struct {
	log            log.Logger
	dataPointStore *store.Store
	feedAddresses  []types.Address
	contract       MedianContract
	dataModel      string
	spread         float64
	expiration     time.Duration
	ticker         *timeutil.Ticker
}

func (w *medianWorker) workerRoutine(ctx context.Context) {
	w.ticker.Start(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.ticker.TickCh():
			if err := w.tryUpdate(ctx); err != nil {
				w.log.
					WithError(err).
					Error("Failed to update Median contract")
			}
		}
	}
}

func (w *medianWorker) tryUpdate(ctx context.Context) error {
	// Current median price.
	val, err := w.contract.Val(ctx)
	if err != nil {
		return err
	}

	// Time of the last update.
	age, err := w.contract.Age(ctx)
	if err != nil {
		return err
	}

	// Quorum.
	bar, err := w.contract.Bar(ctx)
	if err != nil {
		return err
	}

	// Load data points from the store.
	dataPoints, signatures, err := w.getDataPoints(ctx, age, bar)
	if err != nil {
		return err
	}

	prices := dataPointsToPrices(dataPoints)
	median := calculateMedian(prices)
	spread := calculateSpread(median, val)

	// Check if price on the Median contract needs to be updated.
	// The price needs to be updated if:
	// - Price is older than the interval specified in the expiration
	//   field.
	// - Price differs from the current price by more than is specified in the
	//   OracleSpread field.
	isExpired := time.Since(age) >= w.expiration
	isStale := math.IsInf(spread, 0) || spread >= w.spread

	// Print logs.
	w.log.
		WithFields(log.Fields{
			"dataModel":        w.dataModel,
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
		Info("Trying to update Median contract")

	// If price is stale or expired, send update.
	if isExpired || isStale {
		vals := make([]contract.MedianVal, len(prices))
		for i := range dataPoints {
			vals[i] = contract.MedianVal{
				Val: prices[i],
				Age: dataPoints[i].Time,
				V:   uint8(signatures[i].V.Uint64()),
				R:   signatures[i].R,
				S:   signatures[i].S,
			}
		}

		// Send *actual* transaction.
		txHash, tx, err := w.contract.Poke(ctx, vals)
		if err != nil {
			return err
		}

		w.log.
			WithFields(log.Fields{
				"dataModel":              w.dataModel,
				"txHash":                 txHash,
				"txType":                 tx.Type,
				"txFrom":                 tx.From,
				"txTo":                   tx.To,
				"txChainId":              tx.ChainID,
				"txNonce":                tx.Nonce,
				"txGasPrice":             tx.GasPrice,
				"txGasLimit":             tx.GasLimit,
				"txMaxFeePerGas":         tx.MaxFeePerGas,
				"txMaxPriorityFeePerGas": tx.MaxPriorityFeePerGas,
				"txInput":                hexutil.BytesToHex(tx.Input),
			}).
			Info("Sent update to the Median contract")
	}

	return nil
}

func (w *medianWorker) getDataPoints(ctx context.Context, after time.Time, quorum int) ([]datapoint.Point, []types.Signature, error) {
	// Generate slice of random indices to select data points from.
	// It is important to select data points randomly to avoid promoting
	// any particular feed.
	randIndices, err := randomInts(len(w.feedAddresses))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate random indices: %w", err)
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
				WithFields(log.Fields{
					"contract":    w.contract,
					"dataModel":   w.dataModel,
					"feedAddress": w.feedAddresses[i],
				}).
				Warn("Failed to get data point")
			continue
		}
		if !ok {
			continue
		}
		if _, ok := sdp.DataPoint.Value.(value.Tick); !ok {
			w.log.
				WithFields(log.Fields{
					"contract":    w.contract,
					"dataModel":   w.dataModel,
					"feedAddress": w.feedAddresses[i],
				}).
				Warn("Data point is not a tick")
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
		return nil, nil, fmt.Errorf("unable to obtain enough data points, want %d, got %d", quorum, len(dataPoints))
	}

	return dataPoints, signatures, nil
}

// dataPointsToPrices extracts prices from data points.
func dataPointsToPrices(dataPoints []datapoint.Point) []*bn.DecFixedPointNumber {
	prices := make([]*bn.DecFixedPointNumber, len(dataPoints))
	for i, dp := range dataPoints {
		prices[i] = dp.Value.(value.Tick).Price.SetPrecision(contract.MedianPricePrecision)
	}
	return prices
}

// calculateMedian calculates the median price.
func calculateMedian(prices []*bn.DecFixedPointNumber) *bn.DecFixedPointNumber {
	count := len(prices)
	if count == 0 {
		return bn.DecFixedPoint(0, contract.MedianPricePrecision)
	}
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Cmp(prices[j]) < 0
	})
	if count%2 == 0 {
		m := count / 2
		a := prices[m-1]
		b := prices[m]
		return a.Add(b).Div(2)
	}
	return prices[(count-1)/2]
}
