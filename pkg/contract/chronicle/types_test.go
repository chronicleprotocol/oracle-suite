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
	"testing"

	goethABI "github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_uint256FeedBloomValue(t *testing.T) {
	tests := []struct {
		name       string
		addresses  []string
		checkBytes string
	}{
		{
			name:       "single address",
			addresses:  []string{"0x1234567890123456789012345678901234567890"},
			checkBytes: "0x0000000000000000000000000000000000000000000000000000000000040000",
		},
		{
			name:       "multiple addresses",
			addresses:  []string{"0x0c4FC7D66b7b6c684488c1F218caA18D4082da18", "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD", "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9", "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4"},
			checkBytes: "0x0000000000000020000000000000000000200000100000000000000000001000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bloom goethABI.Word
			copy(bloom[:], hexutil.MustHexToBytes(tt.checkBytes))

			var feedIDs FeedIDs
			for _, addressHex := range tt.addresses {
				feedIDs.Add(types.MustAddressFromHex(addressHex))
			}
			bloomType := &uint256FeedBloomValue{}

			// Encode
			require.NoError(t, bloomType.MapFrom(nil, feedIDs))
			words, err := bloomType.EncodeABI()
			require.NoError(t, err)
			assert.Equal(t, tt.checkBytes, hexutil.BytesToHex(words.Bytes()))

			// Decode
			decFeedIDs := FeedIDs{}
			decBloom := &uint256FeedBloomValue{}
			_, err = decBloom.DecodeABI(words)
			require.NoError(t, err)
			require.NoError(t, decBloom.MapTo(nil, &decFeedIDs))
			assert.Equal(t, feedIDs, decFeedIDs)
		})
	}
}
