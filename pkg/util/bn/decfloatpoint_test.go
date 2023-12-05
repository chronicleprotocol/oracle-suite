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
	"strings"
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
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(42), p: 0}},
		},
		{
			name:     "*IntNumber",
			input:    &IntNumber{big.NewInt(42)},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(42), p: 0}},
		},
		{
			name:     "FloatNumber",
			input:    FloatNumber{big.NewFloat(42.5)},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), p: 1}},
		},
		{
			name:     "*FloatNumber",
			input:    &FloatNumber{big.NewFloat(42.5)},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), p: 1}},
		},
		{
			name:     "DecFixedPointNumber",
			input:    DecFixedPointNumber{x: big.NewInt(4250), p: 2},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), p: 2}},
		},
		{
			name:     "*DecFixedPointNumber",
			input:    &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), p: 2}},
		},
		{
			name:     "DecFloatPointNumber",
			input:    DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), p: 2}},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), p: 2}},
		},
		{
			name:     "*DecFloatPointNumber",
			input:    &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), p: 2}},
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), p: 2}},
		},
		{
			name:     "big.Int",
			input:    big.NewInt(42),
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(42), p: 0}},
		},
		{
			name:     "big.Float",
			input:    big.NewFloat(42.5),
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), p: 1}},
		},
		{
			name:     "int",
			input:    int(42),
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(42), p: 0}},
		},
		{
			name:     "float64",
			input:    float64(42.5),
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), p: 1}},
		},
		{
			name:     "string",
			input:    "42.5",
			expected: &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), p: 1}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DecFloatPoint(test.input)
			if test.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, test.expected.String(), result.String())
				assert.Equal(t, test.expected.Prec(), result.Prec())
			}
		})
	}
}

func TestDecFloatPointNumber_String(t *testing.T) {
	tests := []struct {
		name         string
		n            *DecFloatPointNumber
		expectedNum  string
		expectedPrec uint8
	}{
		{
			name:         "zero precision",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), p: 0}}, // 10625
			expectedNum:  "10625",
			expectedPrec: 0,
		},
		{
			name:         "two digits precision",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), p: 2}}, // 106.25
			expectedNum:  "106.25",
			expectedPrec: 2,
		},
		{
			name:         "ten digits precision",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), p: 10}}, // 0.0000010625
			expectedNum:  "0.0000010625",
			expectedPrec: 10,
		},
		{
			name:         "zero precision negative",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(-10625), p: 0}}, // -10625
			expectedNum:  "-10625",
			expectedPrec: 0,
		},
		{
			name:         "two digits precision negative",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(-10625), p: 2}}, // -106.25
			expectedNum:  "-106.25",
			expectedPrec: 2,
		},
		{
			name:         "ten digits precision negative",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(-10625), p: 10}}, // -0.0000010625
			expectedNum:  "-0.0000010625",
			expectedPrec: 10,
		},
		{
			name:         "remove trailing zeros",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1062500), p: 4}}, // 106.2500
			expectedNum:  "106.25",
			expectedPrec: 4,
		},
		{
			name:         "remove trailing zeros (no fractional part)",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1060000), p: 4}}, // 1062500
			expectedNum:  "106",
			expectedPrec: 4,
		},
		{
			name:         "remove trailing zeros (no integer part)",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1062500), p: 10}}, // 0.1062500
			expectedNum:  "0.00010625",
			expectedPrec: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedNum, tt.n.String())
			assert.Equal(t, tt.expectedPrec, tt.n.Prec())
		})
	}
}

func TestDecFloatPointNumber_Add(t *testing.T) {
	tests := []struct {
		name         string
		n1           *DecFloatPointNumber
		n2           *DecFloatPointNumber
		expectedNum  string
		expectedPrec uint8
	}{
		{
			name:         "same precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10500), p: 3}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), p: 3}},  // 2.25
			expectedNum:  "12.75",
			expectedPrec: 2,
		},
		{
			name:         "first higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10500), p: 3}}, // 10.500
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), p: 2}},   // 2.25
			expectedNum:  "12.75",
			expectedPrec: 2,
		},
		{
			name:         "second higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), p: 2}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), p: 3}}, // 2.250
			expectedNum:  "12.75",
			expectedPrec: 2,
		},
		{
			name:         "large precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}},
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}},
			expectedNum:  "2",
			expectedPrec: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Add(tt.n2)
			assert.Equal(t, tt.expectedNum, result.String())
			assert.Equal(t, tt.expectedPrec, result.Prec())
		})
	}
}

func TestDecFloatPointNumber_Sub(t *testing.T) {
	tests := []struct {
		name         string
		n1           *DecFloatPointNumber
		n2           *DecFloatPointNumber
		expectedNum  string
		expectedPrec uint8
	}{
		{
			name:         "same precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), p: 2}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), p: 2}},  // 2.25
			expectedNum:  "8.25",
			expectedPrec: 2,
		},
		{
			name:         "first higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10500), p: 3}}, // 10.500
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), p: 2}},   // 2.25
			expectedNum:  "8.25",
			expectedPrec: 2,
		},
		{
			name:         "second higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), p: 2}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), p: 3}}, // 2.250
			expectedNum:  "8.25",
			expectedPrec: 2,
		},
		{
			name:         "large precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}},
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}},
			expectedNum:  "0",
			expectedPrec: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Sub(tt.n2)
			assert.Equal(t, tt.expectedNum, result.String())
			assert.Equal(t, tt.expectedPrec, result.Prec())
		})
	}
}

func TestDecFloatPointNumber_Mul(t *testing.T) {
	tests := []struct {
		name         string
		n1           *DecFloatPointNumber
		n2           *DecFloatPointNumber
		expectedNum  string
		expectedPrec uint8
	}{
		{
			name:         "same precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), p: 2}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), p: 2}},  // 2.25
			expectedNum:  "23.625",
			expectedPrec: 3,
		},
		{
			name:         "first higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10500), p: 3}}, // 10.500
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(225), p: 2}},   // 2.25
			expectedNum:  "23.625",
			expectedPrec: 3,
		},
		{
			name:         "second higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), p: 2}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), p: 3}}, // 2.250
			expectedNum:  "23.625",
			expectedPrec: 3,
		},
		{
			name:         "second higher precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), p: 2}}, // 10.50
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(2250), p: 3}}, // 2.250
			expectedNum:  "23.625",
			expectedPrec: 3,
		},
		{
			name:         "large precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}},
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}},
			expectedNum:  "1",
			expectedPrec: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Mul(tt.n2)
			assert.Equal(t, tt.expectedNum, result.String())
			assert.Equal(t, tt.expectedPrec, result.Prec())
		})
	}
}

func TestDecFloatPointNumber_Div(t *testing.T) {
	tests := []struct {
		name         string
		n1           *DecFloatPointNumber
		n2           *DecFloatPointNumber
		expectedNum  string
		expectedPrec uint8
	}{
		{
			name:         "same precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), p: 2}}, // 106.25
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(425), p: 2}},   // 4.25
			expectedNum:  "25",                                                                      // 106.25 / 4.25 = 25
			expectedPrec: 0,
		},
		{
			name:         "precision increase",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1), p: 0}}, // 1
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(3), p: 0}}, // 3
			expectedNum:  "0.3333333333333333",                                                  // 1 / 3 = 0.3333333333333333
			expectedPrec: 16,
		},
		{
			name:         "adjust precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(5185185138), p: 7}}, // 518.5185138
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1234567890), p: 8}}, // 12.34567890
			expectedNum:  "42",                                                                           // 518.5185138 / 12.34567890 = 42
			expectedPrec: 0,
		},
		{
			name:         "guard digits",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10), p: 1}}, // 1.0
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(15), p: 1}}, // 1.5
			expectedNum:  "0.66666666666666667",                                                  // 1.0 / 1.5 = 0.66666666666666667
			expectedPrec: 17,
		},
		{
			name:         "large precision",
			n1:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}},
			n2:           &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}},
			expectedNum:  "1",
			expectedPrec: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Div(tt.n2)
			assert.Equal(t, tt.expectedNum, result.String())
			assert.Equal(t, tt.expectedPrec, result.Prec())
		})
	}
}

func TestDecFloatPointNumber_Inv(t *testing.T) {
	tests := []struct {
		name         string
		n            *DecFloatPointNumber
		expectedNum  string
		expectedPrec uint8
	}{
		{
			name:         "106",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(106), p: 0}},
			expectedNum:  "0.0094339622641509",
			expectedPrec: 16,
		},
		{
			name:         "106.25",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(10625), p: 2}},
			expectedNum:  "0.009411764705882353",
			expectedPrec: 18,
		},
		{
			name:         "large precision",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: new(big.Int).Add(pow10(MaxDecPointPrecision), big.NewInt(1)), p: MaxDecPointPrecision}},
			expectedNum:  "0." + strings.Repeat("9", 255),
			expectedPrec: 255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n.Inv()
			assert.Equal(t, tt.expectedNum, result.String())
			assert.Equal(t, tt.expectedPrec, result.Prec())
		})
	}
}

func TestDecFloatPointNumber_adjustPrec(t *testing.T) {
	tests := []struct {
		name         string
		n            *DecFloatPointNumber
		expectedNum  string
		expectedPrec uint8
	}{
		{
			name:         "no change",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(105), p: 1}}, // 10.5
			expectedNum:  "10.5",
			expectedPrec: 1,
		},
		{
			name:         "decrease by 1",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(1050), p: 2}}, // 10.50
			expectedNum:  "10.5",
			expectedPrec: 1,
		},
		{
			name:         "decrease by 6",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(105000000), p: 7}}, // 10.50000
			expectedNum:  "10.5",
			expectedPrec: 1,
		},
		{
			name:         "10^10/10^9",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(10), p: 9}}, // 10
			expectedNum:  "10",
			expectedPrec: 0,
		},
		{
			name:         "10^10/10^10",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(10), p: 10}}, // 1
			expectedNum:  "1",
			expectedPrec: 0,
		},
		{
			name:         "10^10/10^11",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(10), p: 11}}, // 0.1
			expectedNum:  "0.1",
			expectedPrec: 1,
		},
		{
			name:         "10^max/10^(max-1)",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision - 1}}, // 10
			expectedNum:  "10",
			expectedPrec: 0,
		},
		{
			name:         "10^(max+1)/10^max",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision + 1), p: MaxDecPointPrecision}}, // 10
			expectedNum:  "10",
			expectedPrec: 0,
		},
		{
			name:         "10^max/10^max",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: pow10(MaxDecPointPrecision), p: MaxDecPointPrecision}}, // 1
			expectedNum:  "1",
			expectedPrec: 0,
		},
		{
			name:         "0/0",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(0), p: 0}}, // 0
			expectedNum:  "0",
			expectedPrec: 0,
		},
		{
			name:         "0/max",
			n:            &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(0), p: MaxDecPointPrecision}}, // 0
			expectedNum:  "0",
			expectedPrec: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.adjustPrec()
			assert.Equal(t, tt.expectedNum, tt.n.String())
			assert.Equal(t, tt.expectedPrec, tt.n.Prec())
		})
	}
}

func TestDecFloatPointNumber_MulDownFixed(t *testing.T) {
}

func TestDecFloatPointNumber_MulUpFixed(t *testing.T) {
}

func TestDecFloatPointNumber_DivUp(t *testing.T) {}

func TestDecFloatPointNumber_DivUpFixed(t *testing.T) {}

func TestDecFloatPointNumber_DivDownFixed(t *testing.T) {}

func TestDecFloatPointNumber_Mod(t *testing.T) {}

func TestDecFloatPointNumber_Exp(t *testing.T) {}

func TestDecFloatPointNumber_Inflate(t *testing.T) {}

func TestDecFloatPointNumber_Deflate(t *testing.T) {}
