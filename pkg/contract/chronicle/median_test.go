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

// Hash for the AAABBB asset pair, with the price set to 42 and the age to 1605371361:
var priceHash = "0x5e7aa8f6514c872b2020a7f63c72a382e813dc0624a2fb3c28367fee763be154"

func TestMedian_Val(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	median := NewMedian(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899001"))

	mockClient.getStorageAtFn = func(ctx context.Context, account types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error) {
		assert.Equal(t, types.MustAddressFromHex("0x1122344556677889900112233445566778899001"), account)
		assert.Equal(t, types.MustHashFromBigInt(big.NewInt(1)), key)
		assert.Equal(t, types.LatestBlockNumber, block)
		return types.MustHashFromHexPtr("0x00000000000000000000000064e7d1470000000000000584f61606acd0158000", types.PadNone), nil
	}

	val, err := median.Val(ctx)
	require.NoError(t, err)
	assert.Equal(t, "26064.535", val.String())
}

func TestMedian_Age(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	median := NewMedian(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899001"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &median.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0x262a9dff"), call.Input)
		return hexutil.MustHexToBytes("0x0000000000000000000000000000000000000000000000000000000064e7d147"), &types.Call{}, nil
	}

	age, err := median.Age().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, int64(1692913991), age.Unix())
}

func TestMedian_Wat(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	median := NewMedian(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899001"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &median.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0x4ca29923"), call.Input)
		return hexutil.MustHexToBytes("0x4254435553440000000000000000000000000000000000000000000000000000"), &types.Call{}, nil
	}

	wat, err := median.Wat().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, "BTCUSD", wat)
}

func TestMedian_Bar(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	median := NewMedian(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899001"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &median.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0xfebb0f7e"), call.Input)
		return hexutil.MustHexToBytes("0x000000000000000000000000000000000000000000000000000000000000000d"), &types.Call{}, nil
	}

	bar, err := median.Bar().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, 13, bar)
}

func TestMedian_Poke(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	median := NewMedian(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899001"))

	vals := []MedianVal{
		{
			Val: bn.DecFixedPoint(200, 18),
			Age: time.Unix(1888888888, 0),
			V:   1,
			R:   big.NewInt(3),
			S:   big.NewInt(5),
		},
		{
			Val: bn.DecFixedPoint(100, 18),
			Age: time.Unix(1888888889, 0),
			V:   2,
			R:   big.NewInt(4),
			S:   big.NewInt(6),
		},
	}

	calldata := hexutil.MustHexToBytes(
		"0x" +
			"89bbb8b2" +
			"00000000000000000000000000000000000000000000000000000000000000a0" +
			"0000000000000000000000000000000000000000000000000000000000000100" +
			"0000000000000000000000000000000000000000000000000000000000000160" +
			"00000000000000000000000000000000000000000000000000000000000001c0" +
			"0000000000000000000000000000000000000000000000000000000000000220" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000056bc75e2d63100000" +
			"00000000000000000000000000000000000000000000000ad78ebc5ac6200000" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000070962839" +
			"0000000000000000000000000000000000000000000000000000000070962838" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000001" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000004" +
			"0000000000000000000000000000000000000000000000000000000000000003" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000000000000000000000000000000000000000000006" +
			"0000000000000000000000000000000000000000000000000000000000000005",
	)

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &median.address, call.To)
		assert.Equal(t, calldata, call.Input)
		return []byte{}, &types.Call{}, nil
	}

	mockClient.sendTransactionFn = func(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error) {
		assert.Equal(t, types.Transaction{
			Call: types.Call{
				To:    &median.address,
				Input: calldata,
			},
		}, tx)
		return &types.Hash{}, &types.Transaction{}, nil
	}

	_, _, err := median.Poke(vals).SendTransaction(ctx)
	require.NoError(t, err)
}

func Test_ConstructMedianPokeMessage(t *testing.T) {
	assert.Equal(
		t,
		priceHash,
		toEIP191(ConstructMedianPokeMessage("AAABBB", bn.DecFloatPoint(42), time.Unix(1605371361, 0))).String(),
	)
}
