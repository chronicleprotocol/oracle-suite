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

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Deployments(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRPC(t)
	watRegistry := NewWatRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))
	feedRegistry := NewFeedRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))
	registry, err := NewRegistry(feedRegistry, watRegistry)
	require.NoError(t, err)

	mockClient.blockNumberFn = func(ctx context.Context) (*big.Int, error) {
		return big.NewInt(1), nil
	}

	callIdx := 0
	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.Equal(t, int64(1), blockNumber.Big().Int64())

		var data []byte
		switch callIdx {
		case 0: // feeds
			assert.Equal(t, hexutil.MustHexToBytes("0xd63605b8"), call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"0000000000000000000000000000000000000000000000000000000000000020" +
					"0000000000000000000000000000000000000000000000000000000000000002" +
					"0000000000000000000000001234567890123456789012345678901234567890" +
					"0000000000000000000000003456789012345678901234567890123456789012",
			)
		case 1: // wats
			assert.Equal(t, hexutil.MustHexToBytes("0xef293ea1"), call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"0000000000000000000000000000000000000000000000000000000000000020" +
					"0000000000000000000000000000000000000000000000000000000000000002" +
					"4254432f55534400000000000000000000000000000000000000000000000000" +
					"4441492f55534400000000000000000000000000000000000000000000000000",
			)
		case 2: // config BTC/USD
			assert.Equal(t, hexutil.MustHexToBytes(
				"0xcc718f76"+
					"4254432f55534400000000000000000000000000000000000000000000000000",
			), call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"000000000000000000000000000000000000000000000000000000000000000d" +
					"0000000000000000000000000000000000000000000000000000000000040000",
			)
		case 3: // chains BTC/USD
			assert.Equal(t, hexutil.MustHexToBytes(
				"0xc18de0ef"+
					"4254432f55534400000000000000000000000000000000000000000000000000",
			), call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"0000000000000000000000000000000000000000000000000000000000000020" +
					"0000000000000000000000000000000000000000000000000000000000000002" +
					"0000000000000000000000000000000000000000000000000000000000000001" +
					"0000000000000000000000000000000000000000000000000000000000000002",
			)
		case 4: // address BTC/USD on chain 1
			assert.Equal(t, hexutil.MustHexToBytes(
				"0x888039bc"+
					"4254432f55534400000000000000000000000000000000000000000000000000"+
					"0000000000000000000000000000000000000000000000000000000000000001",
			), call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"0000000000000000000000001234567890123456789012345678901234567890",
			)
		case 5: // address BTC/USD on chain 2
			assert.Equal(t, hexutil.MustHexToBytes(
				"0x888039bc"+
					"4254432f55534400000000000000000000000000000000000000000000000000"+
					"0000000000000000000000000000000000000000000000000000000000000002",
			), call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"0000000000000000000000003456789012345678901234567890123456789012",
			)
		case 6: // config DAI/USD
			assert.Equal(t, hexutil.MustHexToBytes(
				"0xcc718f76"+
					"4441492f55534400000000000000000000000000000000000000000000000000",
			), call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"000000000000000000000000000000000000000000000000000000000000000d" +
					"0000000000000000000000000000000000000000000000000010000000000000",
			)
		case 7: // chains DAI/USD
			assert.Equal(t, hexutil.MustHexToBytes(
				"0xc18de0ef"+
					"4441492f55534400000000000000000000000000000000000000000000000000",
			),
				call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"0000000000000000000000000000000000000000000000000000000000000020" +
					"0000000000000000000000000000000000000000000000000000000000000001" +
					"0000000000000000000000000000000000000000000000000000000000000003",
			)
		case 8: // address DAI/USD on chain 3
			assert.Equal(t, hexutil.MustHexToBytes(
				"0x888039bc"+
					"4441492f55534400000000000000000000000000000000000000000000000000"+
					"0000000000000000000000000000000000000000000000000000000000000003",
			), call.Input)
			data = hexutil.MustHexToBytes(
				"0x" +
					"0000000000000000000000004567890123456789012345678901234567890123",
			)
		default:
			t.Fatalf("unexpected call: %d", callIdx)
		}
		callIdx++
		return data, &types.Call{}, nil
	}

	deployments, err := registry.Deployments(ctx)
	require.NoError(t, err)

	assert.Len(t, deployments, 3)
	assert.Equal(t, types.MustAddressFromHex("0x1234567890123456789012345678901234567890"), deployments[0].Address)
	assert.Equal(t, uint64(1), deployments[0].ChainID)
	assert.Equal(t, []types.Address{types.MustAddressFromHex("0x1234567890123456789012345678901234567890")}, deployments[0].Feeds)
	assert.Equal(t, types.MustAddressFromHex("0x3456789012345678901234567890123456789012"), deployments[1].Address)
	assert.Equal(t, uint64(2), deployments[1].ChainID)
	assert.Equal(t, []types.Address{types.MustAddressFromHex("0x1234567890123456789012345678901234567890")}, deployments[1].Feeds)
	assert.Equal(t, types.MustAddressFromHex("0x4567890123456789012345678901234567890123"), deployments[2].Address)
	assert.Equal(t, uint64(3), deployments[2].ChainID)
	assert.Equal(t, []types.Address{types.MustAddressFromHex("0x3456789012345678901234567890123456789012")}, deployments[2].Feeds)
}
