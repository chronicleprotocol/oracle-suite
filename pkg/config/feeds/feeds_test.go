//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package feeds

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

func TestFeeds_Addresses_Valid(t *testing.T) {
	feeds := Feeds{"0x07a35a1d4b751a818d93aa38e615c0df23064881", "2d800d93b065ce011af83f316cef9f0d005b0aa4"}
	addrs, err := feeds.Addresses()
	require.NoError(t, err)

	assert.Equal(t, ethereum.HexToAddress("0x07a35a1d4b751a818d93aa38e615c0df23064881"), addrs[0])
	assert.Equal(t, ethereum.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4"), addrs[1])
}

func TestFeeds_Addresses_Invalid(t *testing.T) {
	feeds := Feeds{"0x07a35a1d4b751a818d93aa38e615c0df23064881", "abc"}
	_, err := feeds.Addresses()

	require.ErrorIs(t, err, ErrInvalidEthereumAddress)
}
