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
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
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

func (s *OpScribe) OpChallengePeriod(ctx context.Context) (time.Duration, error) {
	return s.opChallengePeriod(ctx, types.LatestBlockNumber)
}

func (s *OpScribe) Read(ctx context.Context) (PokeData, error) {
	return s.ReadAt(ctx, time.Now())
}

func (s *OpScribe) ReadAt(ctx context.Context, readTime time.Time) (PokeData, error) {
	blockNumber, err := s.client.BlockNumber(ctx)
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: read query failed: %v", err)
	}
	challengePeriod, err := s.opChallengePeriod(ctx, types.BlockNumberFromBigInt(blockNumber))
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: read query failed: %v", err)
	}
	pokeData, err := s.readPokeData(ctx, pokeStorageSlot, types.BlockNumberFromBigInt(blockNumber))
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: read query failed: %v", err)
	}
	opPokeData, err := s.readPokeData(ctx, opPokeStorageSlot, types.BlockNumberFromBigInt(blockNumber))
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: read query failed: %v", err)
	}
	opPokeDataFinalized := opPokeData.Age.Add(challengePeriod).Before(readTime)
	if opPokeDataFinalized && opPokeData.Age.After(pokeData.Age) {
		return opPokeData, nil
	}
	return pokeData, nil
}

func (s *OpScribe) ReadPokeData(ctx context.Context) (PokeData, error) {
	pokeData, err := s.readPokeData(ctx, pokeStorageSlot, types.LatestBlockNumber)
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: readPokeData query failed: %v", err)
	}
	return pokeData, nil
}

func (s *OpScribe) ReadOpPokeData(ctx context.Context) (PokeData, error) {
	pokeData, err := s.readPokeData(ctx, opPokeStorageSlot, types.LatestBlockNumber)
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: readOpPokeData query failed: %v", err)
	}
	return pokeData, nil
}

func (s *OpScribe) OpPoke(
	ctx context.Context,
	pokeData PokeData,
	schnorrData SchnorrData,
	ecdsaData types.Signature,
) (
	*types.Hash,
	*types.Transaction,
	error,
) {

	calldata, err := abiOpScribe.Methods["opPoke"].EncodeArgs(
		toPokeDataStruct(pokeData),
		toSchnorrDataStruct(schnorrData),
		toECDSADataStruct(ecdsaData),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("opScribe: opPoke failed: %v", err)
	}
	tx := (&types.Transaction{}).
		SetTo(s.address).
		SetInput(calldata)
	if err := simulateTransaction(ctx, s.client, *tx); err != nil {
		return nil, nil, fmt.Errorf("opScribe: opPoke failed: %v", err)
	}
	txHash, txCpy, err := s.client.SendTransaction(ctx, *tx)
	if err != nil {
		return nil, nil, fmt.Errorf("opScribe: opPoke failed: %v", err)
	}
	return txHash, txCpy, nil
}

func (s *OpScribe) opChallengePeriod(ctx context.Context, block types.BlockNumber) (time.Duration, error) {
	res, _, err := s.client.Call(
		ctx,
		types.Call{
			To:    &s.address,
			Input: errutil.Must(abiOpScribe.Methods["opChallengePeriod"].EncodeArgs()),
		},
		block,
	)
	if err != nil {
		return 0, fmt.Errorf("opScribe: opChallengePeriod query failed: %v", err)
	}
	var period uint16
	if err := abiOpScribe.Methods["opChallengePeriod"].DecodeValues(res, &period); err != nil {
		return 0, fmt.Errorf("opScribe: opChallengePeriod query failed: %v", err)
	}
	return time.Second * time.Duration(period), nil
}
