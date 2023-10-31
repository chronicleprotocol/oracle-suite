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

func TestWatRegistry_Bar(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	watRegistry := NewWatRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	mockClient.On(
		"Call",
		ctx,
		types.Call{
			To:    &watRegistry.address,
			Input: hexutil.MustHexToBytes("0xbeaf3edc4554482f55534400000000000000000000000000000000000000000000000000"),
		},
		types.LatestBlockNumber,
	).
		Return(
			hexutil.MustHexToBytes("0x000000000000000000000000000000000000000000000000000000000000000d"),
			&types.Call{},
			nil,
		)

	bar, err := watRegistry.Bar("ETH/USD").Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, 13, bar)
}

func TestWatRegistry_Feeds(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	watRegistry := NewWatRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	// Mocked data for the test
	expectedFeeds := []types.Address{
		types.MustAddressFromHex("0x1234567890123456789012345678901234567890"),
		types.MustAddressFromHex("0x3456789012345678901234567890123456789012"),
	}

	feedData := hexutil.MustHexToBytes(
		"0x" +
			"0000000000000000000000000000000000000000000000000000000000000020" +
			"0000000000000000000000000000000000000000000000000000000000000002" +
			"0000000000000000000000001234567890123456789012345678901234567890" +
			"0000000000000000000000003456789012345678901234567890123456789012",
	)

	mockClient.On(
		"Call",
		ctx,
		types.Call{
			To:    &watRegistry.address,
			Input: hexutil.MustHexToBytes("0xe90f1a434554482f55534400000000000000000000000000000000000000000000000000"),
		},
		types.LatestBlockNumber,
	).
		Return(
			feedData,
			&types.Call{},
			nil,
		)

	feeds, err := watRegistry.Feeds("ETH/USD").Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, expectedFeeds, feeds)
}
