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
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected *FloatNumber
	}{
		{
			name:     "IntNumber",
			input:    IntNumber{big.NewInt(42)},
			expected: &FloatNumber{x: big.NewFloat(42)},
		},
		{
			name:     "*IntNumber",
			input:    &IntNumber{big.NewInt(42)},
			expected: &FloatNumber{x: big.NewFloat(42)},
		},
		{
			name:     "FloatNumber",
			input:    FloatNumber{x: big.NewFloat(42.5)},
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "*FloatNumber",
			input:    &FloatNumber{x: big.NewFloat(42.5)},
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "DecFixedPointNumber",
			input:    DecFixedPointNumber{x: big.NewInt(4250), prec: 2},
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "*DecFixedPointNumber",
			input:    &DecFixedPointNumber{x: big.NewInt(4250), prec: 2},
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "DecFloatPointNumber",
			input:    DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 2}},
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "*DecFloatPointNumber",
			input:    &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 2}},
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "big.Int",
			input:    big.NewInt(42),
			expected: &FloatNumber{x: big.NewFloat(42)},
		},
		{
			name:     "big.Float",
			input:    big.NewFloat(42.5),
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "int",
			input:    42,
			expected: &FloatNumber{x: big.NewFloat(42)},
		},
		{
			name:     "float",
			input:    42.5,
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "string",
			input:    "42.5",
			expected: &FloatNumber{x: big.NewFloat(42.5)},
		},
		{
			name:     "invalid string",
			input:    "invalid",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Float(test.input)
			if test.expected == nil {
				assert.Nil(t, result)
				return
			}
			assert.Equal(t, test.expected.String(), result.String())
		})
	}
}

func TestFloatNumber_String(t *testing.T) {
	f := Float(3.14)
	assert.Equal(t, "3.14", f.String())
}

func TestFloatNumber_Text(t *testing.T) {
	f := Float(3.14)
	assert.Equal(t, "3.1", f.Text('f', 1))
}

func TestFloatNumber_Int(t *testing.T) {
	f := Float(3.14)
	i := f.Int()
	assert.IsType(t, (*IntNumber)(nil), i)
	assert.Equal(t, Int(3), i)
}

func TestFloatNumber_BigInt(t *testing.T) {
	f := Float(3.14)
	bi := f.BigInt()
	assert.IsType(t, (*big.Int)(nil), bi)
	assert.Equal(t, Int(3).BigInt(), bi)
}

func TestFloatNumber_BigFloat(t *testing.T) {
	f := Float(3.14)
	bf := f.BigFloat()
	assert.IsType(t, (*big.Float)(nil), bf)
	assert.Equal(t, f.x, bf)
}

func TestFloatNumber_Sign(t *testing.T) {
	f := Float(-3.14)
	assert.Equal(t, -1, f.Sign())

	f = Float(0.0)
	assert.Equal(t, 0, f.Sign())

	f = Float(3.14)
	assert.Equal(t, 1, f.Sign())
}

func TestFloatNumber_Add(t *testing.T) {
	f := Float(3.14)
	res := f.Add(1.86)
	assert.Equal(t, Float(5.0).String(), res.String())
}

func TestFloatNumber_Sub(t *testing.T) {
	f := Float(3.14)
	res := f.Sub(1.14)
	assert.Equal(t, Float(2.0).String(), res.String())
}

func TestFloatNumber_Mul(t *testing.T) {
	f := Float(3.14)
	res := f.Mul(2)
	assert.Equal(t, Float(6.28).String(), res.String())
}

func TestFloatNumber_Div(t *testing.T) {
	f := Float(3.14)
	res := f.Div(2)
	assert.Equal(t, Float(1.57).String(), res.String())
}

func TestFloatNumber_Sqrt(t *testing.T) {
	f := Float(9)
	res := f.Sqrt()
	assert.Equal(t, Float(3).String(), res.String())
}

func TestFloatNumber_Cmp(t *testing.T) {
	f1 := Float(3.14)
	f2 := Float(3.14)
	assert.Equal(t, 0, f1.Cmp(f2))

	f2 = Float(4)
	assert.Equal(t, -1, f1.Cmp(f2))

	f2 = Float(2)
	assert.Equal(t, 1, f1.Cmp(f2))
}

func TestFloatNumber_Abs(t *testing.T) {
	f := Float(-3.14)
	res := f.Abs()
	assert.Equal(t, Float(3.14), res)

	f = Float(3.14)
	res = f.Abs()
	assert.Equal(t, Float(3.14), res)
}

func TestFloatNumber_Neg(t *testing.T) {
	f := Float(-3.14)
	res := f.Neg()
	assert.Equal(t, Float(3.14), res)

	f = Float(3.14)
	res = f.Neg()
	assert.Equal(t, Float(-3.14), res)
}

func TestFloatNumber_Inv(t *testing.T) {
	f := Float(2.0)
	res := f.Inv()
	assert.Equal(t, Float(0.5), res)

	f = Float(0.5)
	res = f.Inv()
	assert.Equal(t, Float(2.0), res)
}

func TestFloatNumber_IsInf(t *testing.T) {
	f := Float(0.0)
	assert.False(t, f.IsInf())

	f = Float(1.0)
	assert.False(t, f.IsInf())

	f = Float(math.Inf(1))
	assert.True(t, f.IsInf())
}
