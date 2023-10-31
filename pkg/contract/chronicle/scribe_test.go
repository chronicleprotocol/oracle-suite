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

	pokeData, err := scribe.Read(ctx)
	require.NoError(t, err)
	assert.Equal(t, "26064.535", pokeData.Val.String())
	assert.Equal(t, int64(1692913991), pokeData.Age.Unix())
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
			&types.Call{},
			nil,
		)

	wat, err := scribe.Wat().Call(ctx, types.LatestBlockNumber)
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
			&types.Call{},
			nil,
		)

	bar, err := scribe.Bar().Call(ctx, types.LatestBlockNumber)
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
			&types.Call{},
			nil,
		)

	feeds, err := scribe.Feeds().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, expectedFeeds, feeds.Feeds)
	assert.Equal(t, expectedFeedIndices, feeds.FeedIndices)
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
			// "2f529d73" +
			"00000082" + // for optimized poke
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
			&types.Call{},
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
			&types.Transaction{},
			nil,
		)

	_, _, err := scribe.Poke(pokeData, schnorrData).SendTransaction(ctx)
	require.NoError(t, err)
}

func Test_ConstructPokeMessage(t *testing.T) {
	pokeData := PokeData{
		Val: bn.DecFixedPointFromRawBigInt(bn.Int("1649381927392550000000").BigInt(), ScribePricePrecision),
		Age: time.Unix(1693248989, 0),
	}

	message := ConstructScribePokeMessage("ETH/USD", pokeData)
	assert.Equal(t, "0xd469eb1a48223875f0cc0275c64d90077f23cd70dcf2b3d474e5ac3335cb6274", toEIP191(message).String())
}

func TestSignersBlob(t *testing.T) {
	signers := []types.Address{
		types.MustAddressFromHex("0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4"),
		types.MustAddressFromHex("0x0c4FC7D66b7b6c684488c1F218caA18D4082da18"),
		types.MustAddressFromHex("0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9"),
	}
	feeds := []types.Address{
		types.MustAddressFromHex("0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9"),
		types.MustAddressFromHex("0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD"),
		types.MustAddressFromHex("0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4"),
		types.MustAddressFromHex("0x0c4FC7D66b7b6c684488c1F218caA18D4082da18"),
	}
	indices := []uint8{1, 2, 3, 4}

	blob, err := SignersBlob(signers, feeds, indices)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x04, 0x01, 0x03}, blob)
}
