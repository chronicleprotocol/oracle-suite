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
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
)

const ScribePricePrecision = 18

type FeedsResult struct {
	Feeds       []types.Address `abi:"feeds"`
	FeedIndices []uint8         `abi:"feedIndexes"`
}

type Scribe struct {
	client  rpc.RPC
	address types.Address
}

func NewScribe(client rpc.RPC, address types.Address) *Scribe {
	return &Scribe{
		client:  client,
		address: address,
	}
}

func (s *Scribe) Client() rpc.RPC {
	return s.client
}

func (s *Scribe) Address() types.Address {
	return s.address
}

func (s *Scribe) Read(ctx context.Context) (PokeData, error) {
	return s.readPokeData(ctx, pokeStorageSlot, types.LatestBlockNumber)
}

func (s *Scribe) Wat() contract.TypedSelfCaller[string] {
	return contract.NewTypedCall[string](
		contract.CallOpts{
			Client:  s.client,
			Address: s.address,
			Encoder: contract.NewCallEncoder(abiScribe.Methods["wat"]),
			Decoder: func(data []byte, res any) error {
				*res.(*string) = bytes32ToString(data)
				return nil
			},
			ErrorDecoder: contract.NewContractErrorDecoder(abiScribe),
		},
	)
}

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

func (s *Scribe) Feeds() contract.TypedSelfCaller[FeedsResult] {
	method := abiScribe.Methods["feeds"]
	return contract.NewTypedCall[FeedsResult](
		contract.CallOpts{
			Client:       s.client,
			Address:      s.address,
			Encoder:      contract.NewCallEncoder(method),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiScribe),
		},
	)
}

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

// SignersBlob helps to generate signersBlob for PokeData struct.
func SignersBlob(signers []types.Address, feeds []types.Address, indices []uint8) ([]byte, error) {
	if len(feeds) != len(indices) {
		return nil, errors.New("unable to create signers blob: signers and indices slices have different lengths")
	}

	// Make a copy of signers to avoid mutating the original slice.
	signers = sliceutil.Copy(signers)

	// Sort addresses in ascending order.
	sort.Slice(signers, func(i, j int) bool {
		return bytes.Compare(signers[i][:], signers[j][:]) < 0
	})

	// Create a blob where each byte represents the index of a signer.
	blob := make([]byte, 0, len(signers))
	for _, signer := range signers {
		for j, feed := range feeds {
			if feed == signer {
				blob = append(blob, indices[j])
				break
			}
		}
	}

	// Check if all signers were found. If not, probably the feeds is not
	// lifted in the contract.
	if len(blob) != len(signers) {
		return nil, errors.New("unable to create signers blob: unable to find indices for all signers")
	}

	return blob, nil
}
