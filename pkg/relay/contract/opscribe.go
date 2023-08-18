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

package contract

import (
	"context"
	"fmt"
	"math/big"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
)

type OpScribe struct {
	Scribe
}

func NewOpScribe(client rpc.RPC, address types.Address) *OpScribe {
	return &OpScribe{
		Scribe: Scribe{
			client:  client,
			address: address,
		},
	}
}

func (s *OpScribe) OpPoke(ctx context.Context, pokeData PokeData, schnorrData SchnorrData, ecdsaData types.Signature) error {
	calldata, err := abiOpScribe["opPoke"].EncodeArgs(
		toPokeDataStruct(pokeData),
		toSchnorrDataStruct(schnorrData),
		toECDSADataStruct(ecdsaData),
	)
	if err != nil {
		return fmt.Errorf("opScribe: opPoke failed: %v", err)
	}
	nonce, err := s.client.GetTransactionCount(
		ctx,
		s.address,
		types.LatestBlockNumber,
	)
	if err != nil {
		return fmt.Errorf("scribe: poke failed: %v", err)
	}
	tx := (&types.Transaction{}).
		SetType(types.DynamicFeeTxType).
		SetTo(s.address).
		SetInput(calldata).
		SetNonce(nonce).
		SetGasLimit(200000).
		SetMaxPriorityFeePerGas(big.NewInt(1)).
		SetMaxFeePerGas(big.NewInt(2000 * 1e9)) // 2000 Gwei TODO: use gas estimator
	if err := simulateTransaction(ctx, s.client, *tx); err != nil {
		return fmt.Errorf("median: poke failed: %v", err)
	}
	_, err = s.client.SendTransaction(ctx, *tx)
	if err != nil {
		return fmt.Errorf("opScribe: opPoke failed: %v", err)
	}
	return nil
}
