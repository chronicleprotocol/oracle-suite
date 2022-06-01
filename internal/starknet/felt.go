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

package starknet

import (
	"fmt"
	"math/big"
	"strings"
)

type Felt struct {
	*big.Int
}

func HexToFelt(s string) *Felt {
	f := new(Felt)
	f.Int, _ = new(big.Int).SetString(strings.TrimPrefix(s, "0x"), 16)
	return f
}

func (f Felt) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"0x%s"`, f.Text(16))), nil
}

func (f *Felt) UnmarshalJSON(p []byte) error {
	if string(p) == "null" {
		return nil
	}

	// Value must be surrounded with quotes.
	if len(p) < 2 || p[0] != '"' || p[len(p)-1] != '"' {
		return fmt.Errorf("unable to parse felt: %s", string(p))
	}
	f.Int = new(big.Int)
	p = p[1 : len(p)-1]

	// Empty string is treated as zero.
	if len(p) == 0 {
		return nil
	}

	// If value starts with 0x is parsed as hexadecimal value.
	if has0xPrefix(p) {
		f.Int.SetString(strings.TrimPrefix(string(p), "0x"), 16)
		return nil
	}

	// Otherwise, value is parsed as a decimal value.
	if _, ok := f.Int.SetString(string(p), 10); !ok {
		return fmt.Errorf("unable to parse felt: %s", string(p))
	}

	return nil
}

func has0xPrefix(i []byte) bool {
	return len(i) >= 2 && i[0] == '0' && (i[1] == 'x' || i[1] == 'X')
}
