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

package bn

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected *IntNumber
	}{
		{
			name:     "IntNumber",
			input:    IntNumber{big.NewInt(42)},
			expected: &IntNumber{x: big.NewInt(42)},
		},
		{
			name:     "*IntNumber",
			input:    &IntNumber{big.NewInt(42)},
			expected: &IntNumber{x: big.NewInt(42)},
		},
		{
			name:     "FloatNumber",
			input:    &FloatNumber{x: big.NewFloat(42.5)},
			expected: &IntNumber{x: big.NewInt(43)},
		},
		{
			name:     "*FloatNumber",
			input:    &FloatNumber{x: big.NewFloat(42.5)},
			expected: &IntNumber{x: big.NewInt(43)},
		},
		{
			name:     "DecFixedPointNumber",
			input:    DecFixedPointNumber{x: big.NewInt(4250), p: 2},
			expected: &IntNumber{x: big.NewInt(43)},
		},
		{
			name:     "*DecFixedPointNumber",
			input:    &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
			expected: &IntNumber{x: big.NewInt(43)},
		},
		{
			name:     "DecFixedPointNumber",
			input:    DecFixedPointNumber{x: big.NewInt(4250), p: 2},
			expected: &IntNumber{x: big.NewInt(43)},
		},
		{
			name:     "*DecFixedPointNumber",
			input:    &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
			expected: &IntNumber{x: big.NewInt(43)},
		},
		{
			name:     "big.Int",
			input:    big.NewInt(42),
			expected: &IntNumber{x: big.NewInt(42)},
		},
		{
			name:     "big.Float",
			input:    big.NewFloat(42.5),
			expected: &IntNumber{x: big.NewInt(43)},
		},
		{
			name:     "int",
			input:    42,
			expected: &IntNumber{x: big.NewInt(42)},
		},
		{
			name:     "float",
			input:    42.5,
			expected: &IntNumber{x: big.NewInt(43)},
		},
		{
			name:     "string",
			input:    "42",
			expected: &IntNumber{x: big.NewInt(42)},
		},
		{
			name:     "[]byte",
			input:    []byte{0, 0, 0, 42},
			expected: &IntNumber{x: big.NewInt(42)},
		},
		{
			name:     "invalid string",
			input:    "invalid",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Int(test.input)
			if test.expected == nil {
				assert.Nil(t, result)
				return
			}
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIntNumber_String(t *testing.T) {
	i := Int(123)
	assert.Equal(t, "123", i.String())
}

func TestIntNumber_Text(t *testing.T) {
	i := Int(123)
	assert.Equal(t, "1111011", i.Text(2))
}

func TestIntNumber_Float(t *testing.T) {
	i := Int(123)
	f := i.Float()
	assert.IsType(t, (*FloatNumber)(nil), f)
	assert.Equal(t, big.NewFloat(123).String(), f.BigFloat().String())
}

func TestIntNumber_BigInt(t *testing.T) {
	i := Int(123)
	bi := i.BigInt()
	assert.IsType(t, (*big.Int)(nil), bi)
	assert.Equal(t, big.NewInt(123), bi)
}

func TestIntNumber_BigFloat(t *testing.T) {
	i := Int(123)
	bf := i.BigFloat()
	assert.IsType(t, (*big.Float)(nil), bf)
	assert.Equal(t, big.NewFloat(123).String(), bf.String())
}

func TestIntNumber_Sign(t *testing.T) {
	i := Int(-123)
	assert.Equal(t, -1, i.Sign())

	i = Int(0)
	assert.Equal(t, 0, i.Sign())

	i = Int(123)
	assert.Equal(t, 1, i.Sign())
}

func TestIntNumber_Add(t *testing.T) {
	i := Int(123)
	res := i.Add(Int(456))
	assert.Equal(t, Int(579), res)
}

func TestIntNumber_Sub(t *testing.T) {
	i := Int(123)
	res := i.Sub(Int(23))
	assert.Equal(t, Int(100), res)
}

func TestIntNumber_Mul(t *testing.T) {
	i := Int(123)
	res := i.Mul(Int(3))
	assert.Equal(t, Int(369), res)
}

func TestIntNumber_Div(t *testing.T) {
	i := Int(123)
	res := i.Div(Int(3))
	assert.Equal(t, Int(41), res)
}

func TestIntNumber_DivRoundUp(t *testing.T) {
	i := Int(123)
	res := i.DivRoundUp(Int(50))
	assert.Equal(t, Int(3), res)
}

func TestIntNumber_Rem(t *testing.T) {
	i := Int(123)
	res := i.Rem(Int(50))
	assert.Equal(t, Int(23), res)
}

func TestIntNumber_Pow(t *testing.T) {
	i := Int(2)
	res := i.Pow(Int(3))
	assert.Equal(t, Int(8), res)
}

func TestIntNumber_Sqrt(t *testing.T) {
	i := Int(144)
	res := i.Sqrt()
	assert.Equal(t, Int(12), res)
}

func TestIntNumber_Cmp(t *testing.T) {
	i := Int(123)
	j := Int(456)
	assert.Equal(t, -1, i.Cmp(j))

	i = Int(123)
	j = Int(123)
	assert.Equal(t, 0, i.Cmp(j))

	i = Int(456)
	j = Int(123)
	assert.Equal(t, 1, i.Cmp(j))
}

func TestIntNumber_Lsh(t *testing.T) {
	i := Int(7)
	res := i.Lsh(2)
	assert.Equal(t, Int(28), res)
}

func TestIntNumber_Rsh(t *testing.T) {
	i := Int(28)
	res := i.Rsh(2)
	assert.Equal(t, Int(7), res)
}

func TestIntNumber_Abs(t *testing.T) {
	i := Int(-123)
	res := i.Abs()
	assert.Equal(t, Int(123), res)

	i = Int(123)
	res = i.Abs()
	assert.Equal(t, Int(123), res)
}

func TestIntNumber_Neg(t *testing.T) {
	i := Int(-123)
	res := i.Neg()
	assert.Equal(t, Int(123), res)

	i = Int(123)
	res = i.Neg()
	assert.Equal(t, Int(-123), res)
}
