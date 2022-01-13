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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
)

func TestPrivKey_Equals(t *testing.T) {
	sig1 := &mocks.Signer{}
	sig1.On("Address").Return(testAddress1)
	prv1 := NewPrivKey(sig1)

	sig2 := &mocks.Signer{}
	sig2.On("Address").Return(testAddress2)
	prv2 := NewPrivKey(sig2)

	assert.True(t, prv1.Equals(prv1)) //nolint:gocritic
	assert.False(t, prv1.Equals(prv2))
}

func TestPrivKey_GetPublic(t *testing.T) {
	sig := &mocks.Signer{}
	sig.On("Address").Return(testAddress1)
	prv := NewPrivKey(sig)

	pub := prv.GetPublic()
	assert.Equal(t, &PubKey{address: testAddress1}, pub)
}

func TestPrivKey_Raw(t *testing.T) {
	sig := &mocks.Signer{}
	sig.On("Address").Return(testAddress1)
	prv := NewPrivKey(sig)

	bts, err := prv.Raw()
	assert.NoError(t, err)
	assert.Equal(t, testAddress1.Bytes(), bts)
}

func TestPrivKey_Sign(t *testing.T) {
	wthSig := ethereum.SignatureFromBytes([]byte("bar"))
	sig := &mocks.Signer{}
	sig.On("Signature", []byte("foo")).Return(wthSig, nil)
	prv := NewPrivKey(sig)

	ethBts, err := prv.Sign([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, wthSig.Bytes(), ethBts)
}

func TestPrivKey_Type(t *testing.T) {
	assert.Equal(t, KeyTypeID, NewPrivKey(&mocks.Signer{}).Type())
}
