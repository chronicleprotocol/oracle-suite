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

package ethkey

import (
	"bytes"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"

	"github.com/defiweb/go-eth/types"
)

var (
	testAddress1  = types.MustAddressFromHex("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	testAddress2  = types.MustAddressFromHex("0x8eb3daaf5cb4138f5f96711c09c0cfd0288a36e9")
	fakeSignature = bytes.Repeat([]byte{0x01}, 65)
)

func TestAddressToPeerID(t *testing.T) {
	t.Run("correct-address", func(t *testing.T) {
		assert.Equal(
			t,
			"1Afqz6rsuyYpr7Dpp12PbftE22nYH3k2Fw5",
			HexAddressToPeerID("0x69B352cbE6Fc5C130b6F62cc8f30b9d7B0DC27d0").Pretty(),
		)
	})
	t.Run("empty-address", func(t *testing.T) {
		assert.Equal(
			t,
			"",
			HexAddressToPeerID("").Pretty(),
		)
	})
}

func TestPeerIDToAddress(t *testing.T) {
	id, _ := peer.Decode("1Afqz6rsuyYpr7Dpp12PbftE22nYH3k2Fw5")

	assert.Equal(
		t,
		"0x69b352cbe6fc5c130b6f62cc8f30b9d7b0dc27d0",
		PeerIDToAddress(id).String(),
	)
}
