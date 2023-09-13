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

func TestDecFloatPoint(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected *DecFloatPointNumber
	}{
		{
			name:     "IntNumber",
			input:    IntNumber{big.NewInt(42)},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(42), prec: 0}},
		},
		{
			name:     "*IntNumber",
			input:    &IntNumber{big.NewInt(42)},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(42), prec: 0}},
		},
		{
			name:     "FloatNumber",
			input:    FloatNumber{big.NewFloat(42.5)},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), prec: 1}},
		},
		{
			name:     "*FloatNumber",
			input:    &FloatNumber{big.NewFloat(42.5)},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), prec: 1}},
		},
		{
			name:     "DecFixedPointNumber",
			input:    DecFixedPointNumber{x: big.NewInt(4250), prec: 2},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 2}},
		},
		{
			name:     "*DecFixedPointNumber",
			input:    &DecFixedPointNumber{x: big.NewInt(4250), prec: 2},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 2}},
		},
		{
			name:     "DecFloatPointNumber",
			input:    DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 2}},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 2}},
		},
		{
			name:     "*DecFloatPointNumber",
			input:    &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 2}},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 2}},
		},
		{
			name:     "big.Int",
			input:    big.NewInt(42),
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(42), prec: 0}},
		},
		{
			name:     "big.Float",
			input:    big.NewFloat(42.5),
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), prec: 1}},
		},
		{
			name:     "int",
			input:    int(42),
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(42), prec: 0}},
		},
		{
			name:     "float64",
			input:    float64(42.5),
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), prec: 1}},
		},
		{
			name:     "string",
			input:    "42.5",
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), prec: 1}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DecFloatPoint(test.input)
			if test.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, test.expected.String(), result.String())
				assert.Equal(t, test.expected.Precision(), result.Precision())
			}
		})
	}
}

func TestDecFloatPointNumber_String(t *testing.T) {
	tests := []struct {
		name     string
		n        *DecFloatPointNumber
		expected string
	}{
		{
			name:     "zero precision",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), prec: 0}}, // 10625
			expected: "10625",
		},
		{
			name:     "two digits precision",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), prec: 2}}, // 106.25
			expected: "106.25",
		},
		{
			name:     "ten digits precision",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), prec: 10}}, // 0.0000010625
			expected: "0.0000010625",
		},
		{
			name:     "zero precision negative",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(-10625), prec: 0}}, // -10625
			expected: "-10625",
		},
		{
			name:     "two digits precision negative",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(-10625), prec: 2}}, // -106.25
			expected: "-106.25",
		},
		{
			name:     "ten digits precision negative",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(-10625), prec: 10}}, // -0.0000010625
			expected: "-0.0000010625",
		},
		{
			name:     "remove trailing zeros",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1062500), prec: 4}}, // 106.2500
			expected: "106.25",
		},
		{
			name:     "remove trailing zeros (no fractional part)",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1060000), prec: 4}}, // 1062500
			expected: "106",
		},
		{
			name:     "remove trailing zeros (no integer part)",
			n:        &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1062500), prec: 10}}, // 0.1062500
			expected: "0.00010625",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.String())
		})
	}
}

func TestDecFloatPointNumber_Add(t *testing.T) {
	tests := []struct {
		name         string
		n1           *DecFloatPointNumber
		n2           *DecFloatPointNumber
		expectedPrec uint8
		expectedNum  string
	}{
		{
			name:         "same precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10500), prec: 3}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), prec: 3}},  // 2.25
			expectedPrec: 2,
			expectedNum:  "12.75",
		},
		{
			name:         "first higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10500), prec: 3}}, // 10.500
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), prec: 2}},   // 2.25
			expectedPrec: 2,
			expectedNum:  "12.75",
		},
		{
			name:         "second higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), prec: 2}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), prec: 3}}, // 2.250
			expectedPrec: 2,
			expectedNum:  "12.75",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Add(tt.n2)
			assert.Equal(t, tt.expectedNum, result.String())
			assert.Equal(t, tt.expectedPrec, result.Precision())
		})
	}
}

func TestDecFloatPointNumber_Sub(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFloatPointNumber
		n2       *DecFloatPointNumber
		expected string
	}{
		{
			name:     "same precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), prec: 2}}, // 10.50
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), prec: 2}},  // 2.25
			expected: "8.25",
		},
		{
			name:     "first higher precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10500), prec: 3}}, // 10.500
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), prec: 2}},   // 2.25
			expected: "8.25",
		},
		{
			name:     "second higher precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), prec: 2}}, // 10.50
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), prec: 3}}, // 2.250
			expected: "8.25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Sub(tt.n2)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestDecFloatPointNumber_Mul(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFloatPointNumber
		n2       *DecFloatPointNumber
		expected string
	}{
		{
			name:     "same precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), prec: 2}}, // 10.50
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), prec: 2}},  // 2.25
			expected: "23.625",
		},
		{
			name:     "first higher precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10500), prec: 3}}, // 10.500
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), prec: 2}},   // 2.25
			expected: "23.625",
		},
		{
			name:     "second higher precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), prec: 2}}, // 10.50
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), prec: 3}}, // 2.250
			expected: "23.625",
		},
		{
			name:     "second higher precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), prec: 2}}, // 10.50
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), prec: 3}}, // 2.250
			expected: "23.625",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Mul(tt.n2)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestDecFloatPointNumber_Div(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFloatPointNumber
		n2       *DecFloatPointNumber
		expected string
	}{
		{
			name:     "same precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), prec: 2}}, // 106.25
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), prec: 2}},   // 4.25
			expected: "25",
		},
		{
			name:     "first higher precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(106250), prec: 3}}, // 106.250
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), prec: 2}},    // 4.25
			expected: "25",
		},
		{
			name:     "second higher precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), prec: 2}}, // 106.25
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), prec: 3}},  // 4.250
			expected: "25",
		},
		{
			name:     "infinite precision",
			n1:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1), prec: 0}}, // 1
			n2:       &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(3), prec: 0}}, // 3
			expected: "0.333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333333",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Div(tt.n2)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}
