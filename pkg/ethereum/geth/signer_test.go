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

package geth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

// Below values were compared with the recover function in Oracle contracts:
var signerData = []byte("foo")
var signerAddress = common.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
var signerSignature = ethereum.SignatureFromBytes(common.Hex2Bytes("470b7f40fe94916326125b927b4044a496b6fa961beca492b30fce8073f17ff938c2a53ac9c6fb41f7352a38f0ff03bad7d667e91cbf0b3932f7c10fd8475e6b1c"))

func TestSigner_Signature(t *testing.T) {
	account, err := NewAccount("./testdata/keystore", "test123", signerAddress)
	assert.NoError(t, err)

	signer := NewSigner(account)
	retSignature, err := signer.Signature(signerData)
	assert.NoError(t, err)
	assert.Len(t, retSignature, 65)
	assert.Equal(t, signerSignature, retSignature)
}

func TestSigner_Recover(t *testing.T) {
	signer := NewSigner(nil)
	retAddress, err := signer.Recover(signerSignature, signerData)
	assert.NoError(t, err)
	assert.Equal(t, signerAddress.String(), retAddress.String())
}
