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

package ethkey

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
)

func TestPubKey_Equals(t *testing.T) {
	pub1 := NewPubKey(testAddress1)
	pub2 := NewPubKey(testAddress2)

	assert.True(t, pub1.Equals(pub1)) // nolint:gocritic
	assert.False(t, pub1.Equals(pub2))
}

func TestPubKey_Raw(t *testing.T) {
	pub := NewPubKey(testAddress1)

	b, err := pub.Raw()
	assert.NoError(t, err)
	assert.Equal(t, testAddress1.Bytes(), b)
}

func TestPubKey_Type(t *testing.T) {
	assert.Equal(t, KeyTypeID, NewPubKey(testAddress1).Type())
}

func TestPubKey_Verify(t *testing.T) {
	sig := &mocks.Signer{}
	ethSig := ethereum.SignatureFromBytes([]byte("bar"))
	orgSigner := NewSigner
	NewSigner = func() ethereum.Signer { return sig }
	pub := NewPubKey(testAddress1)
	bts := []byte("foo")

	// Valid:
	sig.On("Recover", ethSig, []byte("foo")).Return(&testAddress1, nil).Once()
	ok, err := pub.Verify(bts, ethSig.Bytes())
	assert.True(t, ok)
	assert.NoError(t, err)

	// Invalid:
	sig.On("Recover", ethSig, []byte("foo")).Return(&testAddress2, nil).Once()
	ok, err = pub.Verify(bts, ethSig.Bytes())
	assert.False(t, ok)
	assert.NoError(t, err)

	// Error:
	sig.On("Recover", ethSig, []byte("foo")).Return((*common.Address)(nil), errors.New("err")).Once()
	ok, err = pub.Verify(bts, ethSig.Bytes())
	assert.False(t, ok)
	assert.Error(t, err)

	NewSigner = orgSigner
}
