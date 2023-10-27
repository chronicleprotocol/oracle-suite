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

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/multicall"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/musig/store"
)

type scribe struct {
	contract   ScribeContract
	muSigStore store.SignatureProvider
	dataModel  string
	spread     float64
	expiration time.Duration
	log        log.Logger
}

type scribeState struct {
	wat      string
	bar      int
	feeds    chronicle.FeedsResult
	pokeData chronicle.PokeData
}

func (w *scribe) client() rpc.RPC {
	return w.contract.Client()
}

func (w *scribe) address() types.Address {
	return w.contract.Address()
}

func (w *scribe) createRelayCall(ctx context.Context) (gasEstimate uint64, call contract.Callable) {
	state, err := w.currentState(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to call Scribe contract")
		return 0, nil
	}
	if state.wat != w.dataModel {
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
		if meta.Age.Before(state.pokeData.Age) {
			continue
		}

		// Check if price on the Scribe contract needs to be updated.
		// The price needs to be updated if:
		// - Price is older than the interval specified in the expiration
		//   field.
		// - Price differs from the current price by more than is specified in the
		//   OracleSpread field.
		spread := calculateSpread(state.pokeData.Val.DecFloatPoint(), meta.Val.DecFloatPoint())
		isExpired := time.Since(state.pokeData.Age) >= w.expiration
		isStale := math.IsInf(spread, 0) || spread >= w.spread

		// Generate signersBlob.
		// If signersBlob returns an error, it means that some signers are not
		// present in the feed list on the contract.
		signersBlob, err := chronicle.SignersBlob(s.Signers, state.feeds.Feeds, state.feeds.FeedIndices)
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
				"bar":           state.bar,
				"age":           state.pokeData.Age,
				"val":           state.pokeData.Val,
				"expired":       isExpired,
				"stale":         isStale,
				"expiration":    w.expiration,
				"spread":        w.spread,
				"currentSpread": spread,
			}).
			Debug("Scribe")

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
				w.log.
					WithError(err).
					WithFields(w.logFields()).
					WithAdvice("Ignore if it is related to temporary network issues").
					Error("Failed to poke the Scribe contract")
				return 0, nil
			}

			return gas, poke
		}
	}
	return 0, nil
}

func (w *scribe) currentState(ctx context.Context) (state scribeState, err error) {
	state.pokeData, err = w.contract.Read(ctx)
	if err != nil {
		return scribeState{}, err
	}
	if err := multicall.AggregateCallables(
		w.contract.Client(),
		w.contract.Wat(),
		w.contract.Bar(),
		w.contract.Feeds(),
	).Call(ctx, types.LatestBlockNumber, []any{
		&state.wat,
		&state.bar,
		&state.feeds,
	}); err != nil {
		return scribeState{}, err
	}
	return state, nil
}

func (w *scribe) logFields() log.Fields {
	return log.Fields{
		"address":   w.contract.Address(),
		"dataModel": w.dataModel,
	}
}
