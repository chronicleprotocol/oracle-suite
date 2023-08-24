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
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestScribe_Read(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	scribe := NewScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.On(
		"GetStorageAt",
		ctx,
		scribe.address,
		types.MustHashFromBigInt(big.NewInt(4)),
		types.LatestBlockNumber,
	).
		Return(
			types.MustHashFromHexPtr("0x00000000000000000000000064e7d1470000000000000584f61606acd0158000", types.PadNone),
			nil,
		)

	val, age, err := scribe.Read(ctx)
	require.NoError(t, err)
	assert.Equal(t, "26064.535", val.String())
	assert.Equal(t, int64(1692913991), age.Unix())
}

func TestScribe_Wat(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	scribe := NewScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.On(
		"Call",
		ctx,
		types.Call{
			To:    &scribe.address,
			Input: hexutil.MustHexToBytes("0x4ca29923"),
		},
		types.LatestBlockNumber,
	).
		Return(
			hexutil.MustHexToBytes("0x4254435553440000000000000000000000000000000000000000000000000000"),
			nil,
		)

	wat, err := scribe.Wat(ctx)
	require.NoError(t, err)
	assert.Equal(t, "BTCUSD", wat)
}

func TestScribe_Bar(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	scribe := NewScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.On(
		"Call",
		ctx,
		types.Call{
			To:    &scribe.address,
			Input: hexutil.MustHexToBytes("0xfebb0f7e"),
		},
		types.LatestBlockNumber,
	).
		Return(
			hexutil.MustHexToBytes("0x000000000000000000000000000000000000000000000000000000000000000d"),
			nil,
		)

	bar, err := scribe.Bar(ctx)
	require.NoError(t, err)
	assert.Equal(t, 13, bar)
}

func TestScribe_Feeds(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	scribe := NewScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	// Mocked data for the test
	expectedFeeds := []types.Address{
		types.MustAddressFromHex("0x1234567890123456789012345678901234567890"),
		types.MustAddressFromHex("0x3456789012345678901234567890123456789012"),
	}
	expectedFeedIndices := []uint8{1, 2}

	feedData := hexutil.MustHexToBytes(
		"0x" +
			"0000000000000000000000000000000000000000000000000000000000000040" +
			"00000000000000000000000000000000000000000000000000000000000000a0" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000001234567890123456789012345678901234567890" +
			"0000000000000000000000003456789012345678901234567890123456789012" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"0000000000000000000000000000000000000000000000000000000000000002",
	)

	mockClient.On(
		"Call",
		ctx,
		types.Call{
			To:    &scribe.address,
			Input: hexutil.MustHexToBytes("0xd63605b8"),
		},
		types.LatestBlockNumber,
	).
		Return(
			feedData,
			nil,
		)

	feeds, feedIndices, err := scribe.Feeds(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedFeeds, feeds)
	assert.Equal(t, expectedFeedIndices, feedIndices)
}

func TestScribe_Poke(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	scribe := NewScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	// Mocked data for the test
	pokeData := PokeData{
		Val: bn.DecFixedPoint(26064.535, 18),
		Age: time.Unix(1692913991, 0),
	}
	schnorrData := SchnorrData{
		Signature:   new(big.Int).SetBytes(hexutil.MustHexToBytes("0x1234567890123456789012345678901234567890123456789012345678901234")),
		Commitment:  types.MustAddressFromHex("0x1234567890123456789012345678901234567890"),
		SignersBlob: []byte{0x01, 0x02, 0x03, 0x04},
	}

	calldata := hexutil.MustHexToBytes(
		"0x" +
			"2f529d73" +
			"000000000000000000000000000000000000000000000584f61606acd0134800" +
			"0000000000000000000000000000000000000000000000000000000064e7d147" +
			"0000000000000000000000000000000000000000000000000000000000000060" +
			"1234567890123456789012345678901234567890123456789012345678901234" +
			"0000000000000000000000001234567890123456789012345678901234567890" +
			"0000000000000000000000000000000000000000000000000000000000000060" +
			"0000000000000000000000000000000000000000000000000000000000000004" +
			"0102030400000000000000000000000000000000000000000000000000000000",
	)

	mockClient.On(
		"Call",
		ctx,
		types.Call{
			To:    &scribe.address,
			Input: calldata,
		},
		types.LatestBlockNumber,
	).
		Return(
			[]byte{},
			nil,
		)

	mockClient.On(
		"SendTransaction",
		ctx,
		types.Transaction{
			Call: types.Call{
				To:    &scribe.address,
				Input: calldata,
			},
		},
	).
		Return(
			&types.Hash{},
			nil,
		)

	err := scribe.Poke(ctx, pokeData, schnorrData)
	require.NoError(t, err)
}
