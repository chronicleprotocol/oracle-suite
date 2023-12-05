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
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_uint256FeedBloomValue(t *testing.T) {
	tests := []struct {
		name       string
		addresses  []string
		checkBytes goethABI.Word
	}{
		{
			name:       "single address",
			addresses:  []string{"0x1234567890123456789012345678901234567890"},
			checkBytes: goethABI.Word{0x0, 0x0, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:       "multiple addresses",
			addresses:  []string{"0x1234567890123456789012345678901234567890", "0x3456789012345678901234567890123456789012"},
			checkBytes: goethABI.Word{0x0, 0x0, 0x4, 0x0, 0x0, 0x0, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var feedIDs FeedIDs
			for _, addressHex := range tt.addresses {
				feedIDs.Add(types.MustAddressFromHex(addressHex))
			}
			bloom := &uint256FeedBloomValue{}

			// Encode
			require.NoError(t, bloom.MapFrom(nil, feedIDs))
			words, err := bloom.EncodeABI()
			require.NoError(t, err)
			assert.Equal(t, tt.checkBytes, words[0])

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
