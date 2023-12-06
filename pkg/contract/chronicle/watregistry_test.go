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
	"testing"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatRegistry_Wats(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	watRegistry := NewWatRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		data := hexutil.MustHexToBytes(
			"0x" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000002" +
				"4254432f55534400000000000000000000000000000000000000000000000000" +
				"4441492f55534400000000000000000000000000000000000000000000000000",
		)

		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &watRegistry.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0xef293ea1"), call.Input)
		return data, &types.Call{}, nil
	}

	wats, err := watRegistry.Wats().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, []string{"BTC/USD", "DAI/USD"}, wats)
}

func TestWatRegistry_Exists(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	watRegistry := NewWatRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		data := hexutil.MustHexToBytes("0x0000000000000000000000000000000000000000000000000000000000000001")

		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &watRegistry.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0x38a699a44441492f55534400000000000000000000000000000000000000000000000000"), call.Input)
		return data, &types.Call{}, nil
	}

	exists, err := watRegistry.Exists("DAI/USD").Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, true, exists)
}

func TestWatRegistry_Config(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	watRegistry := NewWatRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		data := hexutil.MustHexToBytes("0x000000000000000000000000000000000000000000000000000000000000000d0000000000000020000000000000000000200000100000000000000000001000")

		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &watRegistry.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0xcc718f764441492f55534400000000000000000000000000000000000000000000000000"), call.Input)
		return data, &types.Call{}, nil
	}

	expectedConfig := ConfigResult{
		Bar:   13,
		Bloom: FeedIDs{0x0c: true, 0x5c: true, 0x75: true, 0xc5: true},
	}

	config, err := watRegistry.Config("DAI/USD").Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, expectedConfig, config)
}

func TestWatRegistry_Chains(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	watRegistry := NewWatRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		data := hexutil.MustHexToBytes(
			"0x" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000002" +
				"0000000000000000000000000000000000000000000000000000000000000001" +
				"0000000000000000000000000000000000000000000000000000000000000002",
		)
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &watRegistry.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0xc18de0ef4441492f55534400000000000000000000000000000000000000000000000000"), call.Input)
		return data, &types.Call{}, nil
	}

	chains, err := watRegistry.Chains("DAI/USD").Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, []uint64{1, 2}, chains)
}

func TestWatRegistry_Deployment(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	watRegistry := NewWatRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &watRegistry.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0x888039bc4441492f555344000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"), call.Input)
		return hexutil.MustHexToBytes("0x0000000000000000000000001234567890123456789012345678901234567890"), &types.Call{}, nil
	}

	config, err := watRegistry.Deployment("DAI/USD", 1).Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, types.MustAddressFromHex("0x1234567890123456789012345678901234567890"), config)
}
