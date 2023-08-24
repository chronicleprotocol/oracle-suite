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
	"math/big"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/relay/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

// TODO: Because the code for OpScribe is so similar to the code for Scribe, we
//       should consider refactoring it to avoid code duplication.

type opScribeWorker struct {
	log        log.Logger
	muSigStore *MuSigStore
	contract   OpScribeContract
	dataModel  string
	spread     float64
	expiration time.Duration
	ticker     *timeutil.Ticker
}

func (w *opScribeWorker) workerRoutine(ctx context.Context) {
	w.ticker.Start(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.ticker.TickCh():
			if err := w.tryUpdate(ctx); err != nil {
				w.log.WithError(err).Error("Failed to update Scribe contract")
			}
		}
	}
}

func (w *opScribeWorker) tryUpdate(ctx context.Context) error {
	// Contract data model.
	wat, err := w.contract.Wat(ctx)
	if err != nil {
		return err
	}
	if wat != w.dataModel {
		return fmt.Errorf("invalid wat returned from contract: %s, expected %s", wat, w.dataModel)
	}

	// Current price and time of the last update.
	val, age, err := w.contract.Read(ctx)
	if err != nil {
		return err
	}

	// Quorum.
	bar, err := w.contract.Bar(ctx)
	if err != nil {
		return err
	}

	// Feed list required to generate signersBlob.
	feeds, indices, err := w.contract.Feeds(ctx)
	if err != nil {
		return err
	}

	// Iterate over all signatures to check if any of them can be used to update
	// the price on the Scribe contract.
	for _, s := range w.muSigStore.OptimisticSignaturesByDataModel(w.dataModel) {
		// Get price and its timestamp from the signature.
		sigValBin, ok := s.MsgMeta["val"]
		if !ok {
			continue
		}
		sigAgeBin, ok := s.MsgMeta["age"]
		if !ok {
			continue
		}
		sigVal := bn.DecFixedPointFromRawBigInt(new(big.Int).SetBytes(sigValBin), contract.ScribePricePrecision)
		sigAge := time.Unix(new(big.Int).SetBytes(sigAgeBin).Int64(), 0)

		// If the signature is older than the current price, skip it.
		if sigAge.Before(age) {
			continue
		}

		// Check if price on the Scribe contract needs to be updated.
		// The price needs to be updated if:
		// - Price is older than the interval specified in the expiration
		//   field.
		// - Price differs from the current price by more than is specified in the
		//   OracleSpread field.
		spread := calculateSpread(val, sigVal)
		isExpired := time.Since(age) >= w.expiration
		isStale := math.IsInf(spread, 0) || spread >= w.spread

		// Generate signersBlob.
		signersBlob := make([]byte, len(s.Signers))
		for i, signer := range s.Signers {
			f := false
			for j, feed := range feeds {
				if feed == signer {
					f = true
					signersBlob[i] = indices[j]
				}
				if !f {
					break
				}
			}
		}

		// If signersBlob is not the same length as the number of feeds, it
		// means that some signers are not present in the feed list on the
		// contract.
		if len(signersBlob) != len(s.Signers) {
			continue
		}

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
			Info("Trying to update Scribe contract")

		// If price is stale or expired, send update.
		if isExpired || isStale {
			// Send *actual* transaction.
			return w.contract.OpPoke(
				ctx,
				contract.PokeData{
					Val: sigVal,
					Age: sigAge,
				},
				contract.SchnorrData{
					Signature:   s.SchnorrSignature,
					Commitment:  s.Commitment,
					SignersBlob: signersBlob,
				},
				s.ECDSASignature,
			)
		}
	}

	return nil
}
