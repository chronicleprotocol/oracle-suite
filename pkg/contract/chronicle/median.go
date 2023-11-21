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
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// MedianPricePrecision is the precision of the price value in the Median contract
// as a number of decimal places after the decimal point.
const MedianPricePrecision = 18

type MedianVal struct {
	Val *bn.DecFixedPointNumber
	Age time.Time
	V   uint8
	R   *big.Int
	S   *big.Int
}

// Median allows interacting with the Median contract.
type Median struct {
	client  rpc.RPC
	address types.Address
}

// NewMedian creates a new Median instance.
func NewMedian(client rpc.RPC, address types.Address) *Median {
	return &Median{
		client:  client,
		address: address,
	}
}

// Client returns the RPC client used to interact with the Median.
func (m *Median) Client() rpc.RPC {
	return m.client
}

// Address returns the address of the Median contract.
func (m *Median) Address() types.Address {
	return m.address
}

// Val returns the current median price value.
func (m *Median) Val(ctx context.Context) (*bn.DecFixedPointNumber, error) {
	const (
		offset = 16
		length = 16
	)
	b, err := m.client.GetStorageAt(
		ctx,
		m.address,
		types.MustHashFromBigInt(big.NewInt(1)),
		types.LatestBlockNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("median: val query failed: %w", err)
	}
	if len(b) < (offset + length) {
		return nil, errors.New("median: val query failed: result is too short")
	}
	return bn.DecFixedPointFromRawBigInt(
		new(big.Int).SetBytes(b[length:offset+length]),
		MedianPricePrecision,
	), nil
}

// Age returns the current median price age.
func (m *Median) Age() contract.TypedSelfCaller[time.Time] {
	method := abiMedian.Methods["age"]
	return contract.NewTypedCall[time.Time](
		contract.CallOpts{
			Client:  m.client,
			Address: m.address,
			Encoder: contract.NewCallEncoder(method),
			Decoder: func(data []byte, res any) error {
				*res.(*time.Time) = time.Unix(new(big.Int).SetBytes(data).Int64(), 0)
				return nil
			},
			ErrorDecoder: contract.NewContractErrorDecoder(abiMedian),
		},
	)
}

// Wat returns the median price asset name.
func (m *Median) Wat() contract.TypedSelfCaller[string] {
	method := abiMedian.Methods["wat"]
	return contract.NewTypedCall[string](
		contract.CallOpts{
			Client:       m.client,
			Address:      m.address,
			Encoder:      contract.NewCallEncoder(method),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiMedian),
		},
	)
}

// Bar returns the median price bar.
func (m *Median) Bar() contract.TypedSelfCaller[int] {
	method := abiMedian.Methods["bar"]
	return contract.NewTypedCall[int](
		contract.CallOpts{
			Client:       m.client,
			Address:      m.address,
			Encoder:      contract.NewCallEncoder(method),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiMedian),
		},
	)
}

// Poke updates the median price value.
func (m *Median) Poke(vals []MedianVal) contract.SelfTransactableCaller {
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].Val.Cmp(vals[j].Val) < 0
	})
	valSlice := make([]*big.Int, len(vals))
	ageSlice := make([]uint64, len(vals))
	vSlice := make([]uint8, len(vals))
	rSlice := make([]*big.Int, len(vals))
	sSlice := make([]*big.Int, len(vals))
	for i, v := range vals {
		valSlice[i] = v.Val.SetPrec(MedianPricePrecision).RawBigInt()
		ageSlice[i] = uint64(v.Age.Unix())
		vSlice[i] = v.V
		rSlice[i] = v.R
		sSlice[i] = v.S
	}
	return contract.NewTransactableCall(
		contract.CallOpts{
			Client:       m.client,
			Address:      m.address,
			Encoder:      contract.NewCallEncoder(abiMedian.Methods["poke"], valSlice, ageSlice, vSlice, rSlice, sSlice),
			ErrorDecoder: contract.NewContractErrorDecoder(abiMedian),
		},
	)
}

// ConstructMedianPokeMessage returns the message expected to be signed via ECDSA for calling
// Median.poke method.
//
// The message structure is defined as:
// H(val ‖ age ‖ wat)
//
// Where:
// - val: a price value
// - age: a time when the price was observed
// - wat: an asset name
func ConstructMedianPokeMessage(wat string, val *bn.DecFloatPointNumber, age time.Time) []byte {
	// Price (val):
	uint256Val := make([]byte, 32)
	val.DecFixedPoint(MedianPricePrecision).RawBigInt().FillBytes(uint256Val)

	// Time (age):
	uint256Age := make([]byte, 32)
	binary.BigEndian.PutUint64(uint256Age[24:], uint64(age.Unix()))

	// Asset name (wat):
	bytes32Wat := make([]byte, 32)
	copy(bytes32Wat, wat)

	// Hash:
	data := make([]byte, 96)
	copy(data[0:32], uint256Val)
	copy(data[32:64], uint256Age)
	copy(data[64:96], bytes32Wat)

	return crypto.Keccak256(data).Bytes()
}
