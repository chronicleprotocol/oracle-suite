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
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/musig/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
)

// stateCacheExpiration is the time after which the cached state is considered
// stale and a new state is fetched from the contract.
const stateCacheExpiration = 10 * time.Second

type scribe struct {
	contract   ScribeContract
	muSigStore store.SignatureProvider
	dataModel  string
	spread     float64
	expiration time.Duration
	log        log.Logger

	cachedState scribeState
}

type scribeState struct {
	wat       string
	bar       int
	feeds     []types.Address
	pokeData  chronicle.PokeData
	finalized bool      // If price is finalized (only for Scribe Optimistic contracts).
	time      time.Time // Date and time when the state was fetched.
}

func (w *scribe) createRelayCall(ctx context.Context) []relayCall {
	state, err := w.currentState(ctx)
	if err != nil {
		w.log.
			WithError(err).
			WithFields(w.logFields()).
			WithAdvice("Ignore if it is related to temporary network issues").
			Error("Failed to call Scribe contract")
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

	// Iterate over all signatures to check if any of them can be used to update
	// the price on the Scribe contract.
	hasValidSigns := false
	for _, s := range w.muSigStore.SignaturesByDataModel(w.dataModel) {
		if s.Commitment.IsZero() || s.SchnorrSignature == nil || len(s.Signers) != state.bar {
			continue
		}

		meta := s.MsgMeta.TickV1()
		if meta == nil || meta.Val == nil {
			continue
		}

		// hasValidSigns is used to check if there are at least one valid signature
		// for the current data model.
		hasValidSigns = true

		// If the signature is older than the current price, skip it.
		if meta.Age.Before(state.pokeData.Age) {
			continue
		}

		// Check if feed addresses included in the signature match the feed
		// addresses from the contract.
		if !sliceutil.ContainsAll(s.Signers, state.feeds) {
			w.log.
				WithError(err).
				WithFields(w.logFields()).
				WithFields(log.Fields{
					"signatureFeeds": chronicle.FeedIDsFromAddresses(s.Signers),
					"contractFeeds":  state.feeds,
				}).
				WithAdvice("This is a bug in the configuration or a list of lifted feeds in not synchronized with the FeedRegistry contract").
				Warn("Signature includes feeds that are not lifted in the contract")
			continue
		}

		// Check if price on the Scribe contract needs to be updated.
		// The price needs to be updated if:
		// - Price is older than the interval specified in the expiration
		//   field.
		// - Price differs from the current price by more than is specified in
		//   the spread field.
		spread := calculateSpread(state.pokeData.Val.DecFloatPoint(), meta.Val.DecFloatPoint())
		isExpired := time.Since(state.pokeData.Age) >= w.expiration
		isStale := math.IsInf(spread, 0) || spread >= w.spread

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

		// If price is stale or expired, return a poke transaction.
		if isExpired || isStale {
			poke := w.contract.Poke(
				chronicle.PokeData{
					Val: meta.Val,
					Age: meta.Age,
				},
				chronicle.SchnorrData{
					Signature:  s.SchnorrSignature,
					Commitment: s.Commitment,
					FeedIDs:    chronicle.FeedIDsFromAddresses(s.Signers),
				},
			)

			gas, err := poke.Gas(ctx, types.LatestBlockNumber)
			if err != nil {
				w.log.
					WithError(err).
					WithFields(w.logFields()).
					WithAdvice("Ignore if it is related to temporary network issues").
					Error("Failed to poke the Scribe contract")
				return nil
			}

			return []relayCall{{
				client:      w.contract.Client(),
				address:     w.contract.Address(),
				callable:    poke,
				gasEstimate: gas,
			}}
		}
	}

	// If there are no valid signatures, this could mean a problem with the
	// configuration.
	if !hasValidSigns {
		w.log.
			WithFields(w.logFields()).
			WithAdvice("Ignore if this occurs within the first few minutes after the relay starts; otherwise, it indicates a configuration error, either in the relay or in the contract"). //nolint:lll
			Warn("No valid signatures found for the current data model")
	}

	return nil
}

func (w *scribe) currentState(ctx context.Context) (state scribeState, err error) {
	if time.Since(w.cachedState.time) <= stateCacheExpiration {
		return w.cachedState, nil
	}
	// Always fetch the latest, non-finalized price from the contract. This is
	// done for two reasons:
	// 1. If a price movement is significant enough to trigger both poke and
	//    opPoke, opPoke is sent first and immediately overwritten by poke.
	//    Using a non-finalized price prevents this because spread for regular
	//	  poke will be calculated using the optimistic price.
	// 2. If the optimistic price is incorrect, poke will overwrite it before
	//    it is finalized.
	switch c := w.contract.(type) {
	case OpScribeContract:
		state.pokeData, state.finalized, err = c.ReadNext(ctx)
	case ScribeContract:
		state.pokeData, err = c.Read(ctx)
		state.finalized = true
	}
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
	state.time = time.Now()
	w.cachedState = state
	return state, nil
}

func (w *scribe) logFields() log.Fields {
	return log.Fields{
		"address":   w.contract.Address(),
		"dataModel": w.dataModel,
	}
}
