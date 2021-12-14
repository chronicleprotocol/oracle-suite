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

package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type Signer struct {
	mock.Mock
}

func (s *Signer) Address() ethereum.Address {
	args := s.Called()
	return args.Get(0).(ethereum.Address)
}

func (s *Signer) SignTransaction(transaction *ethereum.Transaction) error {
	args := s.Called(transaction)
	return args.Error(0)
}

func (s *Signer) Signature(data []byte) (ethereum.Signature, error) {
	args := s.Called(data)
	return args.Get(0).(ethereum.Signature), args.Error(1)
}

func (s *Signer) Recover(signature ethereum.Signature, data []byte) (*ethereum.Address, error) {
	args := s.Called(signature, data)
	return args.Get(0).(*ethereum.Address), args.Error(1)
}
