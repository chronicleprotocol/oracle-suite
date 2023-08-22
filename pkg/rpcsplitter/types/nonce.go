//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

const NonceLength = 8

// Nonce represents a 64 bit nonce.
type Nonce [NonceLength]byte

func HexToNonce(hex string) Nonce {
	var n Nonce
	_ = fixedBytesUnmarshalText([]byte(hex), n[:])
	return n
}

// BytesToNonce converts a byte slice to a Nonce.
func BytesToNonce(b []byte) Nonce {
	var n Nonce
	if len(b) > len(n) {
		return n
	}
	copy(n[NonceLength-len(b):], b)
	return n
}

func (t *Nonce) String() string {
	if t == nil {
		return ""
	}
	return string(bytesToHex(t[:]))
}

func (t Nonce) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(t[:]), nil
}

func (t *Nonce) UnmarshalJSON(input []byte) error {
	return fixedBytesUnmarshalJSON(input, t[:])
}

func (t Nonce) MarshalText() ([]byte, error) {
	return bytesMarshalText(t[:]), nil
}

func (t *Nonce) UnmarshalText(input []byte) error {
	return fixedBytesUnmarshalText(input, t[:])
}
