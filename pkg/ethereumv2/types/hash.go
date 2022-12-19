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

const HashLength = 32

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte

// HexToHash returns the hash corresponding to the hexadecimal string.
func HexToHash(hex string) Hash {
	h := Hash{}
	_ = fixedBytesUnmarshalText([]byte(hex), h[:])
	return h
}

// BytesToHash returns the hash corresponding to the byte slice.
func BytesToHash(bts []byte) Hash {
	var h Hash
	if len(bts) > len(h) {
		return h
	}
	copy(h[HashLength-len(bts):], bts)
	return h
}

// Bytes returns the byte representation of the hash.
func (t *Hash) Bytes() []byte {
	return t[:]
}

// String returns the hex string representation of the hash.
func (t *Hash) String() string {
	if t == nil {
		return ""
	}
	return string(bytesToHex(t[:]))
}

// MarshalJSON implements json.Marshaler.
func (t Hash) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(t[:]), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Hash) UnmarshalJSON(input []byte) error {
	return fixedBytesUnmarshalJSON(input, t[:])
}

// MarshalText implements encoding.TextMarshaler.
func (t Hash) MarshalText() ([]byte, error) {
	return bytesMarshalText(t[:]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Hash) UnmarshalText(input []byte) error {
	return fixedBytesUnmarshalText(input, t[:])
}

// Hashes marshals/unmarshals as hash.
type Hashes []Hash

// HexToHashes returns the hashes corresponding to the hexadecimal strings.
func HexToHashes(hashes ...string) Hashes {
	h := make([]Hash, len(hashes))
	for i, v := range hashes {
		h[i] = HexToHash(v)
	}
	return h
}

// MarshalJSON implements json.Marshaler.
func (b Hashes) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Hash(b))
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Hashes) UnmarshalJSON(input []byte) error {
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		*b = Hashes{{}}
		return json.Unmarshal(input, &((*b)[0]))
	}
	return json.Unmarshal(input, (*[]Hash)(b))
}
