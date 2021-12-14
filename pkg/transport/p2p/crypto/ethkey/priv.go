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
	"bytes"
	"errors"

	"github.com/libp2p/go-libp2p-core/crypto"
	cryptoPB "github.com/libp2p/go-libp2p-core/crypto/pb"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type PrivKey struct {
	signer ethereum.Signer
}

func NewPrivKey(signer ethereum.Signer) crypto.PrivKey {
	return &PrivKey{
		signer: signer,
	}
}

// Bytes implements the crypto.Key interface.
func (p *PrivKey) Bytes() ([]byte, error) {
	return crypto.MarshalPrivateKey(p)
}

// Equals implements the crypto.Key interface.
func (p *PrivKey) Equals(key crypto.Key) bool {
	if p.Type() != key.Type() {
		return false
	}

	a, err := p.Raw()
	if err != nil {
		return false
	}
	b, err := key.Raw()
	if err != nil {
		return false
	}

	return bytes.Equal(a, b)
}

// Raw implements the crypto.Key interface.
func (p *PrivKey) Raw() ([]byte, error) {
	return p.signer.Address().Bytes(), nil
}

// Type implements the crypto.Key interface.
func (p *PrivKey) Type() cryptoPB.KeyType {
	return KeyTypeID
}

// Sign implements the crypto.PrivateKey interface.
func (p *PrivKey) Sign(bytes []byte) ([]byte, error) {
	s, err := p.signer.Signature(bytes)
	if err != nil {
		return nil, err
	}
	return s.Bytes(), nil
}

// GetPublic implements the crypto.PrivateKey interface.
func (p *PrivKey) GetPublic() crypto.PubKey {
	return NewPubKey(p.signer.Address())
}

// UnmarshalEthPrivateKey should return private key from input bytes, but this
// not supported for ethereum keys.
func UnmarshalEthPrivateKey(data []byte) (crypto.PrivKey, error) {
	return nil, errors.New("eth key type does not support unmarshalling")
}
