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

func TestOpScribe_OpChallengePeriod(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	scribe := NewOpScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &scribe.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0x646edb68"), call.Input)
		return hexutil.MustHexToBytes("0x000000000000000000000000000000000000000000000000000000000000012c"), &types.Call{}, nil
	}

	challengePeriod, err := scribe.OpChallengePeriod().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, time.Second*300, challengePeriod)
}

func TestOpScribe_ReadAt(t *testing.T) {
	tests := []struct {
		name        string
		pokeSlot    string
		opPokeSlot  string
		readTime    int64
		expectedVal string
		expectedAge int64
	}{
		{
			name:        "opPoke not finalized",
			pokeSlot:    "0x00000000000000000000000064fa286c0000000000000058a76ad2daafcd2e00",
			opPokeSlot:  "0x00000000000000000000000064fa36c40000000000000058b02c286109d9c580",
			readTime:    1694119920,
			expectedVal: "1635.377164875",
			expectedAge: 1694115948,
		},
		{
			name:        "opPoke finalized",
			pokeSlot:    "0x00000000000000000000000064fa286c0000000000000058a76ad2daafcd2e00",
			opPokeSlot:  "0x00000000000000000000000064fa36c40000000000000058b02c286109d9c580",
			readTime:    1694119921,
			expectedVal: "1636.008044333333333376",
			expectedAge: 1694119620,
		},
		{
			name:        "opPoke overridden",
			pokeSlot:    "0x00000000000000000000000064fa37a10000000000000058a76ad2daafcd2e00",
			opPokeSlot:  "0x00000000000000000000000064fa36c40000000000000058b02c286109d9c580",
			readTime:    1694119921,
			expectedVal: "1635.377164875",
			expectedAge: 1694119841,
		},
		{
			name:        "empty opPoke slot",
			pokeSlot:    "0x00000000000000000000000064fa286c0000000000000058a76ad2daafcd2e00",
			opPokeSlot:  "0x0000000000000000000000000000000000000000000000000000000000000000",
			readTime:    1694119921,
			expectedVal: "1635.377164875",
			expectedAge: 1694115948,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockClient := new(mockRPC)
			scribe := NewOpScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

			mockClient.blockNumberFn = func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(42), nil
			}

			mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
				assert.Equal(t, types.BlockNumberFromUint64(42), blockNumber)
				assert.Equal(t, &scribe.address, call.To)
				assert.Equal(t, hexutil.MustHexToBytes("0x646edb68"), call.Input)
				return hexutil.MustHexToBytes("0x000000000000000000000000000000000000000000000000000000000000012c"), &types.Call{}, nil
			}

			getStorageAtCall := 0
			mockClient.getStorageAtFn = func(ctx context.Context, account types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error) {
				getStorageAtCall++
				switch getStorageAtCall {
				case 1:
					assert.Equal(t, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"), account)
					assert.Equal(t, types.MustHashFromBigInt(big.NewInt(4)), key)
					assert.Equal(t, types.BlockNumberFromUint64(42), block)
					return types.MustHashFromHexPtr(tt.pokeSlot, types.PadNone), nil
				case 2:
					assert.Equal(t, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"), account)
					assert.Equal(t, types.MustHashFromBigInt(big.NewInt(8)), key)
					assert.Equal(t, types.BlockNumberFromUint64(42), block)
					return types.MustHashFromHexPtr(tt.opPokeSlot, types.PadNone), nil
				}
				return nil, nil
			}

			pokeData, _, err := scribe.ReadAt(ctx, time.Unix(tt.readTime, 0))
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVal, pokeData.Val.String())
			assert.Equal(t, tt.expectedAge, pokeData.Age.Unix())
		})
	}
}

func TestOpScribe_OpPoke(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	scribe := NewOpScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	// Mocked data for the test
	pokeData := PokeData{
		Val: bn.DecFixedPoint(26064.535, 18),
		Age: time.Unix(1692913991, 0),
	}
	schnorrData := SchnorrData{
		Signature:  new(big.Int).SetBytes(hexutil.MustHexToBytes("0x1234567890123456789012345678901234567890123456789012345678901234")),
		Commitment: types.MustAddressFromHex("0x1234567890123456789012345678901234567890"),
		FeedIDs:    FeedIDsFromIDs([]byte{0x01, 0x02, 0x03, 0x04}),
	}
	ecdsaData := types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00")

	calldata := hexutil.MustHexToBytes(
		"0x" +
			//"6712af9e" +
			"00000000" + // for optimized opPoke
			"000000000000000000000000000000000000000000000584f61606acd0134800" +
			"0000000000000000000000000000000000000000000000000000000064e7d147" +
			"00000000000000000000000000000000000000000000000000000000000000c0" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff" +
			"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff" +
			"1234567890123456789012345678901234567890123456789012345678901234" +
			"0000000000000000000000001234567890123456789012345678901234567890" +
			"0000000000000000000000000000000000000000000000000000000000000060" +
			"0000000000000000000000000000000000000000000000000000000000000004" +
			"0102030400000000000000000000000000000000000000000000000000000000",
	)

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &scribe.address, call.To)
		assert.Equal(t, calldata, call.Input)
		return []byte{}, &types.Call{}, nil
	}

	mockClient.sendTransactionFn = func(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error) {
		assert.Equal(t, types.Call{
			To:    &scribe.address,
			Input: calldata,
		}, tx.Call)
		return &types.Hash{}, &types.Transaction{}, nil
	}

	_, _, err := scribe.OpPoke(pokeData, schnorrData, ecdsaData).SendTransaction(ctx)
	require.NoError(t, err)
}

func Test_ConstructOpPokeMessage(t *testing.T) {
	wat := "ETH/USD"

	// Poke data.
	pokeData := PokeData{
		Val: bn.DecFixedPointFromRawBigInt(bn.Int("1645737751800000004480").BigInt(), ScribePricePrecision),
		Age: time.Unix(1693259253, 0),
	}

	// Schnorr data.
	schnorrData := SchnorrData{
		Signature:  new(big.Int).SetBytes(hexutil.MustHexToBytes("0xc33523e7517d76ec1260f1a3a9a93808eb2af13986dc89910703916a527a6eba")),
		Commitment: types.MustAddressFromHex("0x139593f8afdd87d1695afa5f839788206f0a09e6"),
	}

	message := ConstructScribeOpPokeMessage(wat, pokeData, schnorrData, FeedIDsFromIDs([]byte{1, 2, 3, 4}))
	assert.Equal(t, "0xc0b793eee861973d6486620f324e9c53a08992d5931218d10302f1f174f411ed", toEIP191(message).String())
}
