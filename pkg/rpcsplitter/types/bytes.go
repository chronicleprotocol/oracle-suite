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

// Bytes represents a byte slice.
type Bytes []byte

// HexToBytes parses a hex string into a Bytes.
func HexToBytes(hex string) Bytes {
	b := Bytes{}
	_ = bytesUnmarshalText([]byte(hex), (*[]byte)(&b))
	return b
}

// Bytes represents a byte slice.
func (t *Bytes) Bytes() []byte {
	return *t
}

func (t *Bytes) String() string {
	if t == nil {
		return ""
	}
	return string(bytesToHex(*t))
}

// MarshalJSON implements json.Marshaler.
func (t Bytes) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(t), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Bytes) UnmarshalJSON(input []byte) error {
	return bytesUnmarshalJSON(input, (*[]byte)(t))
}

// MarshalText implements encoding.TextMarshaler.
func (t Bytes) MarshalText() ([]byte, error) {
	return bytesMarshalText(t), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Bytes) UnmarshalText(input []byte) error {
	return bytesUnmarshalText(input, (*[]byte)(t))
}
