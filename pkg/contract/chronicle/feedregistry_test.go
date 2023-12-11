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

func TestFeedRegistry_Feeds(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	feedRegistry := NewFeedRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		data := hexutil.MustHexToBytes(
			"0x" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000002" +
				"0000000000000000000000001234567890123456789012345678901234567890" +
				"0000000000000000000000003456789012345678901234567890123456789012",
		)

		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &feedRegistry.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0xd63605b8"), call.Input)
		return data, &types.Call{}, nil
	}

	expectedFeeds := []types.Address{
		types.MustAddressFromHex("0x1234567890123456789012345678901234567890"),
		types.MustAddressFromHex("0x3456789012345678901234567890123456789012"),
	}

	wats, err := feedRegistry.Feeds().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, expectedFeeds, wats)
}

func TestFeedRegistry_FeedExists(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	feedRegistry := NewFeedRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		data := hexutil.MustHexToBytes("0x0000000000000000000000000000000000000000000000000000000000000001")

		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &feedRegistry.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0x2fba4aa90000000000000000000000001234567890123456789012345678901234567890"), call.Input)
		return data, &types.Call{}, nil
	}

	exists, err := feedRegistry.FeedExists(types.MustAddressFromHex("0x1234567890123456789012345678901234567890")).Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, true, exists)
}
