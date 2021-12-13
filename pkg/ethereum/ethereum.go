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

package ethereum

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Aliases for the go-ethereum types and functions used in multiple packages.
// These aliases was created to not rely directly on the go-ethereum packages.

// AddressLength is the expected length of the address
const AddressLength = common.AddressLength

type (
	Address = common.Address
	Hash    = common.Hash
)

// HexToAddress returns Address from hex representation.
var HexToAddress = common.HexToAddress

// IsHexAddress verifies if given string is a valid Ethereum address.
var IsHexAddress = common.IsHexAddress

// EmptyAddress contains empty Ethereum address: 0x0000000000000000000000000000000000000000
var EmptyAddress Address

// HexToBytes returns bytes from hex string.
var HexToBytes = common.FromHex

// HexToHash returns Hash from hex string.
var HexToHash = common.HexToHash

// SHA3Hash calculates SHA3 hash.
func SHA3Hash(b []byte) []byte {
	return crypto.Keccak256Hash(b).Bytes()
}
