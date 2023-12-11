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
	"errors"
	"math"
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
)

type opScribe struct {
	scribe

	opContract   OpScribeContract
	opSpread     float64
	opExpiration time.Duration
}

func (w *opScribe) createRelayCall(ctx context.Context) []relayCall {
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

	// If the latest poke is not finalized, we cannot send optimistic poke,
	// try to send a regular poke.
	if !state.finalized {
		return w.scribe.createRelayCall(ctx)
	}

	// Iterate over all signatures to check if any of them can be used to update
	// the price on the Scribe Optimistic contract.
	hasValidSigns := false
	for _, s := range w.muSigStore.SignaturesByDataModel(w.dataModel) {
		if s.Commitment.IsZero() || s.SchnorrSignature == nil || len(s.Signers) != state.bar {
			continue
		}
		meta := s.MsgMeta.TickV1()

		// To send an optimistic poke, the signature must contain an additional
		// ECDSA signature.
		if meta.ECDSAData == nil {
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

		// Check if price on ScribeOptimistic contract needs to be updated.
		// The price needs to be updated if:
		// - Price is older than the interval specified in the opExpiration
		//   field.
		// - Price differs from the current price by more than is specified in
		//   the opSpread field.
		spread := calculateSpread(state.pokeData.Val.DecFloatPoint(), meta.Val.DecFloatPoint())
		isExpired := time.Since(state.pokeData.Age) >= w.opExpiration
		isStale := math.IsInf(spread, 0) || spread >= w.opSpread

		// Print logs.
		w.log.
			WithFields(w.logFields()).
			WithFields(log.Fields{
				"bar":           state.bar,
				"age":           state.pokeData.Age,
				"val":           state.pokeData.Val,
				"expired":       isExpired,
				"stale":         isStale,
				"expiration":    w.opExpiration,
				"spread":        w.opSpread,
				"currentSpread": spread,
			}).
			Debug("ScribeOptimistic")

		// If price is stale or expired, return an optimistic poke transaction.
		if isExpired || isStale {
			poke := w.opContract.OpPoke(
				chronicle.PokeData{
					Val: meta.Val,
					Age: meta.Age,
				},
				chronicle.SchnorrData{
					Signature:  s.SchnorrSignature,
					Commitment: s.Commitment,
					FeedIDs:    chronicle.FeedIDsFromAddresses(s.Signers),
				},
				*meta.ECDSAData,
			)
			gas, err := poke.Gas(ctx, types.LatestBlockNumber)
			if err != nil {
				w.handlePokeErr(err)
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

	// Typically, no poke will be sent at this point because an optimistic poke
	// should have a lower spread and expiration than a regular poke. However,
	// if there are no signatures with an additional optimistic signature for
	// some reason, a regular poke may be sent.
	return w.scribe.createRelayCall(ctx)
}

func (w *opScribe) handlePokeErr(err error) {
	var customError abi.CustomError
	if errors.As(err, &customError) && customError.Type.Name() == "InChallengePeriod" {
		return
	}
	w.log.
		WithError(err).
		WithFields(w.logFields()).
		WithAdvice("Ignore if it is related to temporary network issues").
		Error("Failed to poke the ScribeOptimistic contract")
}
