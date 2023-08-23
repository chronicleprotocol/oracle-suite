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
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

const MedianPricePrecision = 18

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
	res, err := m.client.Call(
		ctx,
		types.Call{
			To:    &m.address,
			Input: errutil.Must(abiMedian["age"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("median: age query failed: %v", err)
	}
	return time.Unix(new(big.Int).SetBytes(res).Int64(), 0), nil
}

func (m *Median) Wat(ctx context.Context) (string, error) {
	res, err := m.client.Call(
		ctx,
		types.Call{
			To:    &m.address,
			Input: errutil.Must(abiMedian["wat"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return "", fmt.Errorf("median: wat query failed: %v", err)
	}
	return string(res), nil
}

func (m *Median) Bar(ctx context.Context) (int, error) {
	res, err := m.client.Call(
		ctx,
		types.Call{
			To:    &m.address,
			Input: errutil.Must(abiMedian["bar"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return 0, fmt.Errorf("median: bar query failed: %v", err)
	}
	return int(new(big.Int).SetBytes(res).Int64()), nil
}

func (m *Median) Poke(ctx context.Context, val []*bn.DecFixedPointNumber, age []time.Time, v []uint8, r []*big.Int, s []*big.Int) error {
	intVal := make([]*big.Int, len(val))
	intAge := make([]int64, len(age))
	for i, v := range val {
		if v.Precision() != MedianPricePrecision {
			return fmt.Errorf("median: poke failed: invalid precision: %d", v.Precision())
		}
		intVal[i] = v.RawBigInt()
	}
	for i, v := range age {
		intAge[i] = v.Unix()
	}
	calldata, err := abiMedian["poke"].EncodeArgs(intVal, intAge, v, r, s)
	if err != nil {
		return fmt.Errorf("median: poke failed: %v", err)
	}
	tx := (&types.Transaction{}).
		SetTo(m.address).
		SetInput(calldata)
	if err := simulateTransaction(ctx, m.client, *tx); err != nil {
		return fmt.Errorf("median: poke failed: %v", err)
	}
	_, err = m.client.SendTransaction(ctx, *tx)
	if err != nil {
		return fmt.Errorf("median: poke failed: %v", err)
	}
	return nil
}
