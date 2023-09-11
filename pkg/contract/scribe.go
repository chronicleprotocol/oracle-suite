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
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
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

func (s *Scribe) Address() types.Address {
	return s.address
}

func (s *Scribe) Read(ctx context.Context) (PokeData, error) {
	return s.readPokeData(ctx, pokeStorageSlot, types.LatestBlockNumber)
}

func (s *Scribe) Wat(ctx context.Context) (string, error) {
	res, _, err := s.client.Call(
		ctx,
		types.Call{
			To:    &s.address,
			Input: errutil.Must(abiScribe.Methods["wat"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return "", fmt.Errorf("scribe: wat query failed: %v", err)
	}
	return bytes32ToString(res), nil
}

func (s *Scribe) Bar(ctx context.Context) (int, error) {
	res, _, err := s.client.Call(
		ctx,
		types.Call{
			To:    &s.address,
			Input: errutil.Must(abiScribe.Methods["bar"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return 0, fmt.Errorf("scribe: bar query failed: %v", err)
	}
	var bar uint8
	if err := abiScribe.Methods["bar"].DecodeValues(res, &bar); err != nil {
		return 0, fmt.Errorf("scribe: bar query failed: %v", err)
	}
	return int(bar), nil
}

func (s *Scribe) Feeds(ctx context.Context) ([]types.Address, []uint8, error) {
	res, _, err := s.client.Call(
		ctx,
		types.Call{
			To:    &s.address,
			Input: errutil.Must(abiScribe.Methods["feeds"].EncodeArgs()),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("scribe: feeds query failed: %v", err)
	}
	var feeds []types.Address
	var feedIndices []uint8
	if err := abiScribe.Methods["feeds"].DecodeValues(res, &feeds, &feedIndices); err != nil {
		return nil, nil, fmt.Errorf("scribe: feeds query failed: %v", err)
	}
	return feeds, feedIndices, nil
}

func (s *Scribe) Poke(ctx context.Context, pokeData PokeData, schnorrData SchnorrData) (*types.Hash, *types.Transaction, error) {
	calldata, err := abiScribe.Methods["poke"].EncodeArgs(toPokeDataStruct(pokeData), toSchnorrDataStruct(schnorrData))
	if err != nil {
		return nil, nil, fmt.Errorf("scribe: poke failed: %v", err)
	}
	tx := (&types.Transaction{}).
		SetTo(s.address).
		SetInput(calldata)
	if err := simulateTransaction(ctx, s.client, *tx); err != nil {
		return nil, nil, fmt.Errorf("scribe: poke failed: %v", err)
	}
	txHash, txCpy, err := s.client.SendTransaction(ctx, *tx)
	if err != nil {
		return nil, nil, fmt.Errorf("scribe: poke failed: %v", err)
	}
	return txHash, txCpy, nil
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
