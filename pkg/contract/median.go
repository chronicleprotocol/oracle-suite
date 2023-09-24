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
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

const MedianPricePrecision = 18

type MedianVal struct {
	Val *bn.DecFixedPointNumber
	Age time.Time
	V   uint8
	R   *big.Int
	S   *big.Int
}

type Median struct {
	client  rpc.RPC
	address types.Address
}

func NewMedian(client rpc.RPC, address types.Address) *Median {
	return &Median{
		client:  client,
		address: address,
	}
}

func (m *Median) Address() types.Address {
	return m.address
}

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
		return nil, fmt.Errorf("median: val query failed: %v", err)
	}
	if len(b) < (offset + length) {
		return nil, errors.New("median: val query failed: result is too short")
	}
	return bn.DecFixedPointFromRawBigInt(
		new(big.Int).SetBytes(b[length:offset+length]),
		MedianPricePrecision,
	), nil
}

func (m *Median) Age(ctx context.Context) (time.Time, error) {
	res, _, err := m.client.Call(
		ctx,
		types.Call{
			To:    &m.address,
			Input: errutil.Must(abiMedian.Methods["age"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("median: age query failed: %v", err)
	}
	return time.Unix(new(big.Int).SetBytes(res).Int64(), 0), nil
}

func (m *Median) Wat(ctx context.Context) (string, error) {
	res, _, err := m.client.Call(
		ctx,
		types.Call{
			To:    &m.address,
			Input: errutil.Must(abiMedian.Methods["wat"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return "", fmt.Errorf("median: wat query failed: %v", err)
	}
	return bytes32ToString(res), nil
}

func (m *Median) Bar(ctx context.Context) (int, error) {
	res, _, err := m.client.Call(
		ctx,
		types.Call{
			To:    &m.address,
			Input: errutil.Must(abiMedian.Methods["bar"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return 0, fmt.Errorf("median: bar query failed: %v", err)
	}
	return int(new(big.Int).SetBytes(res).Int64()), nil
}

func (m *Median) Poke(ctx context.Context, vals []MedianVal) (*types.Hash, *types.Transaction, error) {
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].Val.Cmp(vals[j].Val) < 0
	})
	valSlice := make([]*big.Int, len(vals))
	ageSlice := make([]uint64, len(vals))
	vSlice := make([]uint8, len(vals))
	rSlice := make([]*big.Int, len(vals))
	sSlice := make([]*big.Int, len(vals))
	for i, v := range vals {
		if v.Val.Prec() != MedianPricePrecision {
			return nil, nil, fmt.Errorf("median: poke failed: invalid precision: %d", v.Val.Prec())
		}
		valSlice[i] = v.Val.RawBigInt()
		ageSlice[i] = uint64(v.Age.Unix())
		vSlice[i] = v.V
		rSlice[i] = v.R
		sSlice[i] = v.S
	}
	calldata, err := abiMedian.Methods["poke"].EncodeArgs(valSlice, ageSlice, vSlice, rSlice, sSlice)
	if err != nil {
		return nil, nil, fmt.Errorf("median: poke failed: %v", err)
	}
	tx := (&types.Transaction{}).
		SetTo(m.address).
		SetInput(calldata)
	if err := simulateTransaction(ctx, m.client, abiMedian, *tx); err != nil {
		return nil, nil, fmt.Errorf("median: poke failed: %v", err)
	}
	txHash, txCpy, err := m.client.SendTransaction(ctx, *tx)
	if err != nil {
		return nil, nil, fmt.Errorf("median: poke failed: %v", err)
	}
	return txHash, txCpy, nil
}

// ConstructMedianPokeMessage returns the message expected to be signed via ECDSA for calling
// Median.poke method.
//
// The message structure is defined as:
// H(tag ‖ H(val ‖ age ‖ wat)
//
// Where:
// - tag:
// - val: a price value
// - age: a time when the price was observed
// - wat: an asset name
func ConstructMedianPokeMessage(wat string, val *bn.DecFloatPointNumber, age time.Time) types.Hash {
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

	return crypto.Keccak256(crypto.AddMessagePrefix(crypto.Keccak256(data).Bytes()))
}
