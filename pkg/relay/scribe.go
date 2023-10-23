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
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/musig/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type scribeWorker struct {
	contract       ScribeContract
	muSigStore     store.SignatureProvider
	dataModel      string
	spread         float64
	expiration     time.Duration
	shouldUpdateAt time.Time
	ticker         *timeutil.Ticker
	log            log.Logger
}

func (w *scribeWorker) client() rpc.RPC {
	return w.contract.Client()
}

func (w *scribeWorker) address() types.Address {
	return w.contract.Address()
}

func (w *scribeWorker) createRelayCall(ctx context.Context) (gasEstimate uint64, call contract.Callable) {
	wat, bar, feeds, pokeData, err := w.currentState(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to call Scribe contract")
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
		signersBlob, err := chronicle.SignersBlob(s.Signers, feeds.Feeds, feeds.FeedIndices)
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
			poke := w.contract.Poke(
				chronicle.PokeData{
					Val: meta.Val,
					Age: meta.Age,
				},
				chronicle.SchnorrData{
					Signature:   s.SchnorrSignature,
					Commitment:  s.Commitment,
					SignersBlob: signersBlob,
				},
			)

			gas, err := poke.Gas(ctx, types.LatestBlockNumber)
			if err != nil {
				w.handlePokeErr(err)
				return 0, nil
			}

			return gas, poke
		}
	}
	return 0, nil
}

func (w *scribeWorker) currentState(ctx context.Context) (wat string, bar int, feeds chronicle.FeedsResult, pokeData chronicle.PokeData, err error) {
	pokeData, err = w.contract.Read(ctx)
	if err != nil {
		return "", 0, chronicle.FeedsResult{}, chronicle.PokeData{}, err
	}
	if err := multicall.AggregateCallables(
		w.contract.Client(),
		w.contract.Wat(),
		w.contract.Bar(),
		w.contract.Feeds(),
	).Call(ctx, types.LatestBlockNumber, []any{
		&wat,
		&bar,
		&feeds,
	}); err != nil {
		return "", 0, chronicle.FeedsResult{}, chronicle.PokeData{}, err
	}
	return wat, bar, feeds, pokeData, nil
}

func (w *scribeWorker) handlePokeErr(err error) {
	if strings.Contains(err.Error(), "replacement transaction underpriced") {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("This is expected during large price movements; the relay tries to update multiple contracts at once").
			Warn("Failed to poke the Scribe contract; previous transaction is still pending")
		return
	}
	/*
		if contract.IsRevert(err) {
			w.log.
				WithError(err).
				WithFields(w.logFields()).
				WithAdvice("Probably caused by a race condition between multiple relays; if this is a case, no action is required").
				Error("Failed to poke the Scribe contract")
			return
		}
	*/
	w.log.
		WithError(err).
		WithFields(w.logFields()).
		WithAdvice("Ignore if it is related to temporary network issues").
		Error("Failed to poke the Scribe contract")
}

func (w *scribeWorker) logFields() log.Fields {
	return log.Fields{
		"address":   w.contract.Address(),
		"dataModel": w.dataModel,
	}
}
