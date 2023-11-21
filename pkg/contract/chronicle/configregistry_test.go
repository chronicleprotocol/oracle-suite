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

func TestConfigRegistry_Latest(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	configRegistry := NewConfigRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	ipfsExpected := "ipfs://sample"

	mockClient.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		data := hexutil.MustHexToBytes(
			"0x" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"000000000000000000000000000000000000000000000000000000000000000d" +
				"697066733a2f2f73616d706c6500000000000000000000000000000000000000",
		)
		assert.Equal(t, types.LatestBlockNumber, blockNumber)
		assert.Equal(t, &configRegistry.address, call.To)
		assert.Equal(t, hexutil.MustHexToBytes("0x52bfe789"), call.Input)

		return data, &types.Call{}, nil
	}

	ipfs, err := configRegistry.Latest().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, ipfsExpected, ipfs)
}
