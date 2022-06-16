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
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type SeedSigner struct {
	key *ecdsa.PrivateKey
}

func NewSeedSigner(seed []byte) *SeedSigner {
	key, err := crypto.ToECDSA(seed)
	if err != nil {
		panic(err.Error())
	}
	return &SeedSigner{
		key: key,
	}
}

// Address implements the ethereum.Signer interface.
func (s *SeedSigner) Address() ethereum.Address {
	if s.key == nil {
		return ethereum.Address{}
	}
	return crypto.PubkeyToAddress(s.key.PublicKey)
}

// SignTransaction implements the ethereum.Signer interface.
func (s *SeedSigner) SignTransaction(transaction *ethereum.Transaction) error {
	return nil
}

// Signature implements the ethereum.Signer interface.
func (s *SeedSigner) Signature(data []byte) (ethereum.Signature, error) {
	msg := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))
	sig, err := crypto.Sign(crypto.Keccak256Hash(msg).Bytes(), s.key)
	sig[64] += 27
	if err != nil {
		panic(err.Error())
	}
	return ethereum.SignatureFromBytes(sig), nil
}

// Recover implements the ethereum.Signer interface.
func (s *SeedSigner) Recover(signature ethereum.Signature, data []byte) (*ethereum.Address, error) {
	return Recover(signature, data)
}
