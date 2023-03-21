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

const BloomLength = 256

// Bloom represents a 2048 bit bloom filter.
type Bloom [BloomLength]byte

// BytesToBloom converts a byte slice to a Bloom.
func BytesToBloom(bts []byte) Bloom {
	var b Bloom
	if len(bts) > len(b) {
		return b
	}
	copy(b[BloomLength-len(bts):], bts)
	return b
}

// String returns the hex string representation of the bloom.
func (t *Bloom) String() string {
	if t == nil {
		return ""
	}
	return string(bytesToHex(t[:]))
}

// MarshalJSON implements the json.Marshaler interface.
func (t Bloom) MarshalJSON() ([]byte, error) {
	return bytesMarshalJSON(t[:]), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Bloom) UnmarshalJSON(input []byte) error {
	return fixedBytesUnmarshalJSON(input, t[:])
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t Bloom) MarshalText() ([]byte, error) {
	return bytesMarshalText(t[:]), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *Bloom) UnmarshalText(input []byte) error {
	return fixedBytesUnmarshalText(input, t[:])
}
