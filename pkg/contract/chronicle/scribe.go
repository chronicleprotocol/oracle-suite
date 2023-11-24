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

package chronicle

import (
	"context"
	"encoding/binary"
	"math/big"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// ScribePricePrecision is the precision of the price value in the Scribe contract
// as a number of decimal places after the decimal point.
const ScribePricePrecision = 18

// Scribe allows interacting with the Scribe contract.
type Scribe struct {
	client  rpc.RPC
	address types.Address
}

// NewScribe creates a new Scribe instance.
func NewScribe(client rpc.RPC, address types.Address) *Scribe {
	return &Scribe{
		client:  client,
		address: address,
	}
}

// Client returns the RPC client used to interact with the Scribe.
func (s *Scribe) Client() rpc.RPC {
	return s.client
}

// Address returns the address of the Scribe contract.
func (s *Scribe) Address() types.Address {
	return s.address
}

// Read reads the poke data from the contract.
func (s *Scribe) Read(ctx context.Context) (PokeData, error) {
	return s.readPokeData(ctx, pokeStorageSlot, types.LatestBlockNumber)
}

// Wat returns the wat value from the contract.
func (s *Scribe) Wat() contract.TypedSelfCaller[string] {
	method := abiScribe.Methods["wat"]
	return contract.NewTypedCall[string](
		contract.CallOpts{
			Client:       s.client,
			Address:      s.address,
			Encoder:      contract.NewCallEncoder(method),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiScribe),
		},
	)
}

// Bar returns the bar value from the contract.
func (s *Scribe) Bar() contract.TypedSelfCaller[int] {
	method := abiScribe.Methods["bar"]
	return contract.NewTypedCall[int](
		contract.CallOpts{
			Client:       s.client,
			Address:      s.address,
			Encoder:      contract.NewCallEncoder(method),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiScribe),
		},
	)
}

// Feeds returns Chronicle Protocol's feeds that are lifted in the contract.
func (s *Scribe) Feeds() contract.TypedSelfCaller[[]types.Address] {
	method := abiScribe.Methods["feeds"]
	return contract.NewTypedCall[[]types.Address](
		contract.CallOpts{
			Client:       s.client,
			Address:      s.address,
			Encoder:      contract.NewCallEncoder(method),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiScribe),
		},
	)
}

// Poke updates the poke data in the contract.
func (s *Scribe) Poke(pokeData PokeData, schnorrData SchnorrData) contract.SelfTransactableCaller {
	return contract.NewTransactableCall(
		contract.CallOpts{
			Client:  s.client,
			Address: s.address,
			Encoder: contract.NewCallEncoder(
				abiScribe.Methods["poke"],
				toPokeDataStruct(pokeData),
				toSchnorrDataStruct(schnorrData),
			),
			ErrorDecoder: contract.NewContractErrorDecoder(abiScribe),
		},
	)
}

func (s *Scribe) readPokeData(ctx context.Context, storageSlot int, block types.BlockNumber) (PokeData, error) {
	const (
		ageOffset = 0
		valOffset = 16
		ageLength = 16
		valLength = 16
	)
	b, err := s.client.GetStorageAt(
		ctx,
		s.address,
		types.MustHashFromBigInt(big.NewInt(int64(storageSlot))),
		block,
	)
	if err != nil {
		return PokeData{}, err
	}
	val := bn.DecFixedPointFromRawBigInt(
		new(big.Int).SetBytes(b[valOffset:valOffset+valLength]),
		ScribePricePrecision,
	)
	age := time.Unix(
		new(big.Int).SetBytes(b[ageOffset:ageOffset+ageLength]).Int64(),
		0,
	)
	return PokeData{
		Val: val,
		Age: age,
	}, nil
}

// ConstructScribePokeMessage returns the message expected to be signed via ECDSA for calling
// Scribe.poke method.
//
// The message is defined as:
// H(wat ‖ val ‖ age)
//
// Where:
// - wat: an asset name
// - val: a price value
// - age: a time when the price was observed
func ConstructScribePokeMessage(wat string, pokeData PokeData) []byte {
	// Asset name (wat):
	bytes32Wat := make([]byte, 32)
	copy(bytes32Wat, wat)

	// Price (val):
	uint128Val := make([]byte, 16)
	pokeData.Val.SetPrec(ScribePricePrecision).RawBigInt().FillBytes(uint128Val)

	// Time (age):
	uint32Age := make([]byte, 4)
	binary.BigEndian.PutUint32(uint32Age, uint32(pokeData.Age.Unix()))

	data := make([]byte, 52) //nolint:gomnd
	copy(data[0:32], bytes32Wat)
	copy(data[32:48], uint128Val)
	copy(data[48:52], uint32Age)

	return crypto.Keccak256(data).Bytes()
}
