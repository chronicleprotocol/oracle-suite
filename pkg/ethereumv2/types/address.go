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

package types

import (
	"encoding/json"
)

const AddressLength = 20

// Address represents the 20 byte address of an Ethereum account.
type Address [AddressLength]byte

// HexToAddress parses a hex string into an Address.
func HexToAddress(hex string) Address {
	a := Address{}
	_ = fixedBytesUnmarshalText([]byte(hex), a[:])
	return a
}

// BytesToAddress returns an Address from a byte slice.
func BytesToAddress(bts []byte) Address {
	var a Address
	if len(bts) > len(a) {
		return a
	}
	copy(a[AddressLength-len(bts):], bts)
	return a
}

// Bytes returns the byte representation of the address.
func (t *Address) Bytes() []byte {
	return t[:]
}

// String returns the hex string representation of the address.
func (t *Address) String() string {
	if t == nil {
		return ""
	}
	return string(bytesToHex(t[:]))
}

// MarshalJSON implements json.Marshaler.
func (t Address) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(t[:]), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Address) UnmarshalJSON(input []byte) error {
	return fixedBytesUnmarshalJSON(input, t[:])
}

// MarshalText implements encoding.TextMarshaler.
func (t Address) MarshalText() ([]byte, error) {
	return bytesMarshalText(t[:]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Address) UnmarshalText(input []byte) error {
	return fixedBytesUnmarshalText(input, t[:])
}

// Addresses is a slice of Address.
type Addresses []Address

// HexToAddresses parses a list of hex strings into Addresses.
func HexToAddresses(addresses ...string) Addresses {
	a := make(Addresses, len(addresses))
	for i, address := range addresses {
		a[i] = HexToAddress(address)
	}
	return a
}

// MarshalJSON implements json.Marshaler.
func (t Addresses) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Address(t))
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Addresses) UnmarshalJSON(input []byte) error {
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		*t = Addresses{{}}
		return json.Unmarshal(input, &((*t)[0]))
	}
	return json.Unmarshal(input, (*[]Address)(t))
}
