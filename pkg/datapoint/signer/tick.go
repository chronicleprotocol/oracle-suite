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

package signer

import (
	"context"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
)

// TickSigner signs tick data points and recovers the signer address from a
// signature.
type TickSigner struct {
	signer wallet.Key
}

// NewTickSigner creates a new TickSigner instance.
func NewTickSigner(signer wallet.Key) *TickSigner {
	return &TickSigner{signer: signer}
}

// Supports implements the Signer interface.
func (t *TickSigner) Supports(_ context.Context, data datapoint.Point) bool {
	_, ok := data.Value.(value.Tick)
	return ok
}

// Sign implements the Signer interface.
func (t *TickSigner) Sign(_ context.Context, model string, data datapoint.Point) (*types.Signature, error) {
	return t.signer.SignMessage(
		chronicle.ConstructMedianPokeMessage(
			model,
			data.Value.(value.Tick).Price,
			data.Time,
		),
	)
}

// TickRecoverer recovers the signer address from a tick data point and a
// signature.
type TickRecoverer struct {
	recoverer crypto.Recoverer
}

// NewTickRecoverer creates a new TickRecoverer instance.
func NewTickRecoverer(recoverer crypto.Recoverer) *TickRecoverer {
	return &TickRecoverer{recoverer: recoverer}
}

// Supports implements the Recoverer interface.
func (t *TickRecoverer) Supports(_ context.Context, data datapoint.Point) bool {
	_, ok := data.Value.(value.Tick)
	return ok
}

// Recover implements the Recoverer interface.
func (t *TickRecoverer) Recover(
	_ context.Context,
	model string,
	data datapoint.Point,
	signature types.Signature,
) (*types.Address, error) {
	return t.recoverer.RecoverMessage(
		chronicle.ConstructMedianPokeMessage(
			model,
			data.Value.(value.Tick).Price,
			data.Time,
		),
		signature,
	)
}
