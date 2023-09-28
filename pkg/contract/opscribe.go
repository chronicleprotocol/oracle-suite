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
	"encoding/binary"
	"fmt"
	"time"

	"github.com/defiweb/go-eth/crypto"
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
		return PokeData{}, fmt.Errorf("opScribe: read query failed: %w", err)
	}
	challengePeriod, err := s.opChallengePeriod(ctx, types.BlockNumberFromBigInt(blockNumber))
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: read query failed: %w", err)
	}
	pokeData, err := s.readPokeData(ctx, pokeStorageSlot, types.BlockNumberFromBigInt(blockNumber))
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: read query failed: %w", err)
	}
	opPokeData, err := s.readPokeData(ctx, opPokeStorageSlot, types.BlockNumberFromBigInt(blockNumber))
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: read query failed: %w", err)
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
		return PokeData{}, fmt.Errorf("opScribe: readPokeData query failed: %w", err)
	}
	return pokeData, nil
}

func (s *OpScribe) ReadOpPokeData(ctx context.Context) (PokeData, error) {
	pokeData, err := s.readPokeData(ctx, opPokeStorageSlot, types.LatestBlockNumber)
	if err != nil {
		return PokeData{}, fmt.Errorf("opScribe: readOpPokeData query failed: %w", err)
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
		return nil, nil, fmt.Errorf("opScribe: opPoke failed: %w", err)
	}
	tx := (&types.Transaction{}).
		SetTo(s.address).
		SetInput(calldata)
	if err := simulateTransaction(ctx, s.client, abiOpScribe, *tx); err != nil {
		return nil, nil, fmt.Errorf("opScribe: opPoke failed: %w", err)
	}
	txHash, txCpy, err := s.client.SendTransaction(ctx, *tx)
	if err != nil {
		return nil, nil, fmt.Errorf("opScribe: opPoke failed: %w", err)
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
		return 0, fmt.Errorf("opScribe: opChallengePeriod query failed: %w", err)
	}
	var period uint16
	if err := abiOpScribe.Methods["opChallengePeriod"].DecodeValues(res, &period); err != nil {
		return 0, fmt.Errorf("opScribe: opChallengePeriod query failed: %w", err)
	}
	return time.Second * time.Duration(period), nil
}

// ConstructScribeOpPokeMessage returns the message expected to be signed via ECDSA for calling
// OpScribe.opPoke method.
//
// The message structure is defined as:
// H(tag ‖ H(wat ‖ val ‖ age ‖ signature ‖ commitment ‖ signersBlob))
//
// Where:
// - tag is the message prefix (EIP-191)
// - wat: an asset name
// - val: a price value
// - age: a time when the price was observed
// - signature: a Schnorr signature
// - commitment: a Schnorr commitment
// - signersBlob: a byte slice with signers indexes obtained from a contract
func ConstructScribeOpPokeMessage(wat string, pokeData PokeData, schnorrData SchnorrData, signersBlob []byte) types.Hash {
	// Asset name (wat):
	bytes32Wat := make([]byte, 32)
	copy(bytes32Wat, wat)

	// Price (val):
	uint128Val := make([]byte, 16)
	pokeData.Val.SetPrec(ScribePricePrecision).RawBigInt().FillBytes(uint128Val)

	// Time (age):
	uint32Age := make([]byte, 4)
	binary.BigEndian.PutUint32(uint32Age, uint32(pokeData.Age.Unix()))

	// Signature:
	bytes32Signature := make([]byte, 32)
	schnorrData.Signature.FillBytes(bytes32Signature)

	// Address:
	bytes20Commitment := make([]byte, 20) //nolint:gomnd
	copy(bytes20Commitment, schnorrData.Commitment.Bytes())

	data := make([]byte, len(signersBlob)+104) //nolint:gomnd
	copy(data[0:32], bytes32Wat)
	copy(data[32:48], uint128Val)
	copy(data[48:52], uint32Age)
	copy(data[52:84], bytes32Signature)
	copy(data[84:104], bytes20Commitment)
	copy(data[104:], signersBlob)

	return crypto.Keccak256(crypto.AddMessagePrefix(crypto.Keccak256(data).Bytes()))
}
