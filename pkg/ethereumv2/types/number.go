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
	"math/big"
)

// Number represents an integer up to 256 bits.
type Number struct{ x big.Int }

func HexToNumber(hex string) Number {
	b := &Number{}
	_ = b.UnmarshalText([]byte(hex))
	return *b
}

// BigToNumber converts a big.Int to a Number.
func BigToNumber(x *big.Int) Number {
	return Number{x: *new(big.Int).Set(x)}
}

// Uint64ToNumber converts an uint64 to a Number.
func Uint64ToNumber(x uint64) Number {
	return Number{x: *new(big.Int).SetUint64(x)}
}

func (t *Number) Big() *big.Int {
	return new(big.Int).Set(&t.x)
}

func (t *Number) String() string {
	if t == nil {
		return ""
	}
	return "0x" + t.Big().Text(16)
}

// MarshalJSON implements json.Marshaler.
func (t Number) MarshalJSON() ([]byte, error) {
	return numberMarshalJSON(t.Big()), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Number) UnmarshalJSON(input []byte) error {
	return numberUnmarshalJSON(input, &t.x)
}

// MarshalText implements encoding.TextMarshaler.
func (t Number) MarshalText() ([]byte, error) {
	return numberMarshalText(t.Big()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Number) UnmarshalText(input []byte) error {
	return numberUnmarshalText(input, &t.x)
}
