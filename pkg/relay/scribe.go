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

	"github.com/defiweb/go-eth/hexutil"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/musig/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type scribeWorker struct {
	log            log.Logger
	muSigStore     store.SignatureProvider
	contract       ScribeContract
	dataModel      string
	spread         float64
	expiration     time.Duration
	delay          time.Duration
	shouldUpdateAt time.Time
	ticker         *timeutil.Ticker
}

func (w *scribeWorker) workerRoutine(ctx context.Context) {
	w.ticker.Start(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.ticker.TickCh():
			w.tryUpdate(ctx, time.Now())
		}
	}
}

func (w *scribeWorker) tryUpdate(ctx context.Context, t time.Time) {
	// Contract data model.
	wat, err := w.contract.Wat(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to get current asset name from the Scribe contract")
		return
	}
	if wat != w.dataModel {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("This is a bug in the configuration, probably a wrong contract address is used").
			Error("Contract asset name does not match the configured asset name")
		return
	}

	// Current price and time of the last update.
	pokeData, err := w.contract.Read(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to get current price from the Scribe contract")
		return
	}

	// Quorum.
	bar, err := w.contract.Bar(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to get quorum from the Scribe contract")
		return
	}

	// Feed list required to generate signersBlob.
	feeds, indices, err := w.contract.Feeds(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to get feed list from the Scribe contract")
		return
	}

	// Iterate over all signatures to check if any of them can be used to update
	// the price on the Scribe contract.
	for _, s := range w.muSigStore.SignaturesByDataModel(w.dataModel) {
		if s.Commitment.IsZero() || s.SchnorrSignature == nil {
			continue
		}

		meta := s.MsgMeta.TickV1()
		if meta == nil || meta.Val == nil {
			continue
		}

		// If the signature is older than the current price, skip it.
		if meta.Age.Before(pokeData.Age) {
			continue
		}

		// Check if price on the Scribe contract needs to be updated.
		// The price needs to be updated if:
		// - Price is older than the interval specified in the expiration
		//   field.
		// - Price differs from the current price by more than is specified in the
		//   OracleSpread field.
		spread := calculateSpread(pokeData.Val.DecFloatPoint(), meta.Val.DecFloatPoint())
		isExpired := time.Since(pokeData.Age) >= w.expiration
		isStale := math.IsInf(spread, 0) || spread >= w.spread

		// Generate signersBlob.
		// If signersBlob returns an error, it means that some signers are not
		// present in the feed list on the contract.
		signersBlob, err := contract.SignersBlob(s.Signers, feeds, indices)
		if err != nil {
			w.log.
				WithError(err).
				WithFields(w.logFields()).
				Error("Failed to generate signersBlob")
		}

		// Print logs.
		w.log.
			WithFields(w.logFields()).
			WithFields(log.Fields{
				"bar":              bar,
				"age":              pokeData.Age,
				"val":              pokeData.Val,
				"expired":          isExpired,
				"stale":            isStale,
				"expiration":       w.expiration,
				"spread":           w.spread,
				"timeToExpiration": time.Since(pokeData.Age).String(),
				"currentSpread":    spread,
			}).
			Debug("Scribe worker")

		// If price is stale or expired, send update.
		if isExpired || isStale {
			if w.delay > 0 {
				if w.shouldUpdateAt.IsZero() {
					w.shouldUpdateAt = t.Add(w.delay)
					return
				}
				if t.Before(w.shouldUpdateAt) {
					return
				}
			}

			// Send *actual* transaction.
			txHash, tx, err := w.contract.Poke(
				ctx,
				contract.PokeData{
					Val: meta.Val,
					Age: meta.Age,
				},
				contract.SchnorrData{
					Signature:   s.SchnorrSignature,
					Commitment:  s.Commitment,
					SignersBlob: signersBlob,
				},
			)
			if err != nil {
				if strings.Contains(err.Error(), "replacement transaction underpriced") {
					w.log.
						WithError(err).
						WithFields(w.logFields()).
						WithAdvice("This is expected during large price movements; the relay tries to update multiple contracts at once").
						Warn("Failed to poke the Scribe contract; previous transaction is still pending")
					return
				}
				if contract.IsRevert(err) {
					w.log.
						WithError(err).
						WithFields(w.logFields()).
						WithAdvice("Probably caused by a race condition between multiple relays; if this is a case, no action is required").
						Error("Failed to poke the Scribe contract")
					return
				}
				w.log.
					WithError(err).
					WithFields(w.logFields()).
					WithAdvice("Ignore if it is related to temporary network issues").
					Error("Failed to poke the Scribe contract")
				return
			}

			w.log.
				WithFields(w.logFields()).
				WithFields(log.Fields{
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
				Info("Sent update to the Scribe contract")
		} else {
			w.shouldUpdateAt = time.Time{}
		}
	}
}

func (w *scribeWorker) logFields() log.Fields {
	return log.Fields{
		"contractAddress": w.contract.Address(),
		"dataModel":       w.dataModel,
	}
}
