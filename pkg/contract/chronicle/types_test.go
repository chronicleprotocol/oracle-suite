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

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
)

func TestFeedBloom(t *testing.T) {
	tests := []struct {
		name       string
		addresses  []string
		checkBytes [32]byte
	}{
		{
			name:       "Single Address",
			addresses:  []string{"0x1234567890123456789012345678901234567890"},
			checkBytes: [32]byte{0x0, 0x0, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:       "Multiple Addresses",
			addresses:  []string{"0x1234567890123456789012345678901234567890", "0x3456789012345678901234567890123456789012"},
			checkBytes: [32]byte{0x0, 0x0, 0x4, 0x0, 0x0, 0x0, 0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bloom FeedBloom
			for _, addressHex := range tt.addresses {
				address := types.MustAddressFromHex(addressHex)
				bloom.Set(address)
				assert.True(t, bloom.Has(address), "Address should be present in bloom")
			}
			assert.Equal(t, tt.checkBytes, bloom.Bytes32())
		})
	}
}

func TestFeedBloom_Bytes32(t *testing.T) {
	input := [32]byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
	var bloom FeedBloom
	bloom.SetBytes32(input)
	assert.Equal(t, input, bloom.Bytes32())
}
