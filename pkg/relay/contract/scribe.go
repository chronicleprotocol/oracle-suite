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
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

const ScribePricePrecision = 18

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

func (s *Scribe) Read(ctx context.Context) (*bn.DecFixedPointNumber, time.Time, error) {
	const (
		storageSlot = 4
		ageOffset   = 0
		valOffset   = 16
		ageLength   = 16
		valLength   = 16
	)
	b, err := s.client.GetStorageAt(
		ctx,
		s.address,
		types.MustHashFromBigInt(big.NewInt(storageSlot)),
		types.LatestBlockNumber,
	)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("scribe: read query failed: %v", err)
	}
	val := bn.DecFixedPointFromRawBigInt(
		new(big.Int).SetBytes(b[valOffset:valOffset+valLength]),
		ScribePricePrecision,
	)
	age := time.Unix(
		new(big.Int).SetBytes(b[ageOffset:ageOffset+ageLength]).Int64(),
		0,
	)
	return val, age, nil
}

func (s *Scribe) Wat(ctx context.Context) (string, error) {
	res, err := s.client.Call(
		ctx,
		types.Call{
			To:    &s.address,
			Input: errutil.Must(abiScribe["wat"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return "", fmt.Errorf("scribe: wat query failed: %v", err)
	}
	return bytesToString(res), nil
}

func (s *Scribe) Bar(ctx context.Context) (int, error) {
	res, err := s.client.Call(
		ctx,
		types.Call{
			To:    &s.address,
			Input: errutil.Must(abiScribe["bar"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return 0, fmt.Errorf("scribe: bar query failed: %v", err)
	}
	var bar uint8
	if err := abiScribe["bar"].DecodeValues(res, &bar); err != nil {
		return 0, fmt.Errorf("scribe: bar query failed: %v", err)
	}
	return int(bar), nil
}

func (s *Scribe) Feeds(ctx context.Context) ([]types.Address, []uint8, error) {
	res, err := s.client.Call(
		ctx,
		types.Call{
			To:    &s.address,
			Input: errutil.Must(abiScribe["feeds"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("scribe: feeds query failed: %v", err)
	}
	var feeds []types.Address
	var feedIndices []uint8
	if err := abiScribe["feeds"].DecodeValues(res, &feeds, &feedIndices); err != nil {
		return nil, nil, fmt.Errorf("scribe: feeds query failed: %v", err)
	}
	return feeds, feedIndices, nil
}

func (s *Scribe) Poke(ctx context.Context, pokeData PokeData, schnorrData SchnorrData) error {
	calldata, err := abiScribe["poke"].EncodeArgs(toPokeDataStruct(pokeData), toSchnorrDataStruct(schnorrData))
	if err != nil {
		return fmt.Errorf("scribe: poke failed: %v", err)
	}
	tx := (&types.Transaction{}).
		SetTo(s.address).
		SetInput(calldata)
	if err := simulateTransaction(ctx, s.client, *tx); err != nil {
		return fmt.Errorf("median: poke failed: %v", err)
	}
	_, err = s.client.SendTransaction(ctx, *tx)
	if err != nil {
		return fmt.Errorf("scribe: poke failed: %v", err)
	}
	return nil
}

// bytesToString converts a string terminated by a null byte to a Go string.
func bytesToString(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		return string(b)
	}
	return string(b[:n])
}
