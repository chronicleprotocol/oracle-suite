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

	goethABI "github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigRegistry_Latest(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	configRegistry := NewConfigRegistry(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	ipfsExpected := "ipfs://sample"

	stringAbi := goethABI.MustParseType("(string memory)")
	stringMap := make(map[string]string)
	stringMap["arg0"] = ipfsExpected
	ipfsBytes := goethABI.MustEncodeValue(stringAbi, stringMap)

	mockClient.On(
		"Call",
		ctx,
		types.Call{
			To:    &configRegistry.address,
			Input: hexutil.MustDecode("0x52bfe789"),
		},
		types.LatestBlockNumber,
	).Return(
		ipfsBytes,
		&types.Call{},
		nil,
	)

	ipfs, err := configRegistry.Latest().Call(ctx, types.LatestBlockNumber)
	require.NoError(t, err)
	assert.Equal(t, ipfsExpected, ipfs)
}
