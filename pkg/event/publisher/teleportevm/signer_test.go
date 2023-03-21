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

package teleportevm

import (
	"testing"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	"github.com/defiweb/go-eth/wallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestSigner_IgnoreUnsupportedType(t *testing.T) {
	msg := &messages.Event{Type: "foo"}
	key := &mocks.Key{}
	signer := NewSigner(key, []string{"bar"})

	// If message is of different type, signer should do nothing:
	ok, err := signer.Sign(msg)
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestSigner_MissingHashField(t *testing.T) {
	msg := &messages.Event{Type: "foo"}
	key := &mocks.Key{}
	signer := NewSigner(key, []string{"foo"})

	// If hash field is missing, an error must be returned:
	ok, err := signer.Sign(msg)
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestSigner_Sign(t *testing.T) {
	address := types.MustAddressFromHex("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	key, err := wallet.NewKeyFromJSON("./keystore/1.json", "test123")
	require.NoError(t, err)
	msg := &messages.Event{Type: "foo", Data: map[string][]byte{"hash": common.HexToHash("f76b84eff86432f629ab567880256b50c8eb31cafaec58c5edb24d9b4c246470").Bytes()}}
	signer := NewSigner(key, []string{"foo"})

	ok, err := signer.Sign(msg)
	assert.True(t, ok)
	assert.NoError(t, err)

	// Verify if address in signer field is correct:
	assert.Equal(t, msg.Signatures[SignatureKey].Signer, address.Bytes())

	// Verify signature:
	recovered, err := crypto.ECRecoverer.RecoverMessage(msg.Data["hash"], types.MustSignatureFromBytes(msg.Signatures[SignatureKey].Signature))
	require.NoError(t, err)
	assert.Equal(t, address, *recovered)
}
