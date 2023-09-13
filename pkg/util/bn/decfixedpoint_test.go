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

func TestDecFixedPoint(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		prec     uint8
		expected *DecFixedPointNumber
	}{
		{
			name:     "IntNumber",
			input:    IntNumber{big.NewInt(42)},
			prec:     0,
			expected: &DecFixedPointNumber{x: big.NewInt(42), p: 0},
		},
		{
			name:     "*IntNumber",
			input:    &IntNumber{big.NewInt(42)},
			prec:     0,
			expected: &DecFixedPointNumber{x: big.NewInt(42), p: 0},
		},
		{
			name:     "FloatNumber",
			input:    FloatNumber{big.NewFloat(42.5)},
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},
		{
			name:     "*FloatNumber",
			input:    &FloatNumber{big.NewFloat(42.5)},
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},
		{
			name:     "DecFixedPointNumber",
			input:    DecFixedPointNumber{x: big.NewInt(4250), p: 2},
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},
		{
			name:     "*DecFixedPointNumber",
			input:    &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},
		{
			name:     "DecFloatPointNumber",
			input:    DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), p: 2}},
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},
		{
			name:     "*DecFloatPointNumber",
			input:    &DecFloatPointNumber{x: &DecFixedPointNumber{x: big.NewInt(4250), p: 2}},
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},
		{
			name:     "big.Int",
			input:    big.NewInt(42),
			prec:     0,
			expected: &DecFixedPointNumber{x: big.NewInt(42), p: 0},
		},
		{
			name:     "big.Float",
			input:    big.NewFloat(42.5),
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},
		{
			name:     "int",
			input:    int(42),
			prec:     0,
			expected: &DecFixedPointNumber{x: big.NewInt(42), p: 0},
		},
		{
			name:     "float64",
			input:    float64(42.5),
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},
		{
			name:     "string",
			input:    "42.5",
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), p: 2},
		},

		{
			name:     "big.Float",
			input:    big.NewFloat(1.03),
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(103), p: 2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DecFixedPoint(test.input, test.prec)
			if test.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, test.expected.String(), result.String())
				assert.Equal(t, test.expected.Prec(), result.Prec())
			}
		})
	}
}

func TestDecFixedPointNumber_String(t *testing.T) {
	tests := []struct {
		name     string
		n        *DecFixedPointNumber
		expected string
	}{
		{
			name:     "zero precision",
			n:        &DecFixedPointNumber{x: big.NewInt(10625), p: 0}, // 10625
			expected: "10625",
		},
		{
			name:     "two digits precision",
			n:        &DecFixedPointNumber{x: big.NewInt(10625), p: 2}, // 106.25
			expected: "106.25",
		},
		{
			name:     "ten digits precision",
			n:        &DecFixedPointNumber{x: big.NewInt(10625), p: 10}, // 0.0000010625
			expected: "0.0000010625",
		},
		{
			name:     "zero precision negative",
			n:        &DecFixedPointNumber{x: big.NewInt(-10625), p: 0}, // -10625
			expected: "-10625",
		},
		{
			name:     "two digits precision negative",
			n:        &DecFixedPointNumber{x: big.NewInt(-10625), p: 2}, // -106.25
			expected: "-106.25",
		},
		{
			name:     "ten digits precision negative",
			n:        &DecFixedPointNumber{x: big.NewInt(-10625), p: 10}, // -0.0000010625
			expected: "-0.0000010625",
		},
		{
			name:     "remove trailing zeros",
			n:        &DecFixedPointNumber{x: big.NewInt(1062500), p: 4}, // 106.2500
			expected: "106.25",
		},
		{
			name:     "remove trailing zeros (no fractional part)",
			n:        &DecFixedPointNumber{x: big.NewInt(1060000), p: 4}, // 106
			expected: "106",
		},
		{
			name:     "remove trailing zeros (no integer part)",
			n:        &DecFixedPointNumber{x: big.NewInt(1062500), p: 10}, // 0.00010625
			expected: "0.00010625",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.String())
		})
	}
}

func TestDecFixedPointNumber_Add(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFixedPointNumber
		n2       *DecFixedPointNumber
		expected string
	}{
		{
			name:     "same precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), p: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(225), p: 2},  // 2.25
			expected: "12.75",                                         // 10.50 + 2.25 = 12.75
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10500), p: 3}, // 10.500
			n2:       &DecFixedPointNumber{x: big.NewInt(225), p: 2},   // 2.25
			expected: "12.75",                                          // 10.500 + 2.25 = 12.75
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), p: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(2250), p: 3}, // 2.250
			expected: "12.75",                                         // 10.50 + 2.250 = 12.75
		},
		{
			name:     "maximum precision",
			n1:       DecFixedPoint(10.50, MaxDecPointPrecision), // 10.50
			n2:       DecFixedPoint(2.25, MaxDecPointPrecision),  // 2.25
			expected: "12.75",                                    // 10.50 + 2.25 = 12.75
		},
		{
			name:     "zero and maximum precision",
			n1:       DecFixedPoint(10.50, 0),                   // 11.00
			n2:       DecFixedPoint(2.25, MaxDecPointPrecision), // 2.25
			expected: "13",                                      // 11.00 + 2.25 = 13.25 -> 13
		},
		{
			name:     "maximum and zero precision",
			n1:       DecFixedPoint(10.50, MaxDecPointPrecision), // 10.50
			n2:       DecFixedPoint(2.25, 0),                     // 2.00
			expected: "12.5",                                     // 10.50 + 2.00 = 12.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Add(tt.n2)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestDecFixedPointNumber_Sub(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFixedPointNumber
		n2       *DecFixedPointNumber
		expected string
	}{
		{
			name:     "same precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), p: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(225), p: 2},  // 2.25
			expected: "8.25",                                          // 10.50 - 2.25 = 8.25
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10500), p: 3}, // 10.500
			n2:       &DecFixedPointNumber{x: big.NewInt(225), p: 2},   // 2.25
			expected: "8.25",                                           // 10.500 - 2.25 = 8.25
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), p: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(2250), p: 3}, // 2.250
			expected: "8.25",                                          // 10.50 - 2.250 = 8.25
		},
		{
			name:     "maximum precision",
			n1:       DecFixedPoint(10.50, MaxDecPointPrecision), // 10.50
			n2:       DecFixedPoint(2.25, MaxDecPointPrecision),  // 2.25
			expected: "8.25",                                     // 10.50 - 2.25 = 8.25
		},
		{
			name:     "zero and maximum precision",
			n1:       DecFixedPoint(10.50, 0),                   // 11.00
			n2:       DecFixedPoint(2.25, MaxDecPointPrecision), // 2.25
			expected: "9",                                       // 11.00 - 2.25 = 8.75 -> 9
		},
		{
			name:     "maximum and zero precision",
			n1:       DecFixedPoint(10.50, MaxDecPointPrecision), // 10.50
			n2:       DecFixedPoint(2.25, 0),                     // 2.00
			expected: "8.5",                                      // 10.50 - 2.00 = 8.50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Sub(tt.n2)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestDecFixedPointNumber_Mul(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFixedPointNumber
		n2       *DecFixedPointNumber
		expected string
	}{
		{
			name:     "same precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), p: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(225), p: 2},  // 2.25
			expected: "23.63",                                         // 10.50 * 2.25 = 23.625 -> 23.62
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10500), p: 3}, // 10.500
			n2:       &DecFixedPointNumber{x: big.NewInt(225), p: 2},   // 2.25
			expected: "23.625",                                         // 10.500 * 2.25 = 23.625
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), p: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(2250), p: 3}, // 2.250
			expected: "23.63",                                         // 10.50 * 2.250 = 23.625 -> 23.63
		},
		{
			name:     "maximum precision",
			n1:       DecFixedPoint("10.5", MaxDecPointPrecision), // 10.50
			n2:       DecFixedPoint("2.25", MaxDecPointPrecision), // 2.25
			expected: "23.625",                                    // 10.50 * 2.25 = 23.625
		},
		{
			name:     "zero and maximum precision",
			n1:       DecFixedPoint("10.5", 0),                    // 11.00
			n2:       DecFixedPoint("2.25", MaxDecPointPrecision), // 2.25
			expected: "25",                                        // 11.00 * 2.25 = 24.75 -> 25
		},
		{
			name:     "maximum and zero precision",
			n1:       DecFixedPoint("10.5", MaxDecPointPrecision), // 10.50
			n2:       DecFixedPoint("2.25", 0),                    // 2.00
			expected: "21",                                        // 10.50 * 2.00 = 21.00
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Mul(tt.n2)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestDecFixedPointNumber_Div(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFixedPointNumber
		n2       *DecFixedPointNumber
		expected string
	}{
		{
			name:     "same precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(20), p: 1}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(1), p: 1},  // 4.25
			expected: "20",
		},
		{
			name:     "same precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), p: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(425), p: 2},   // 4.25
			expected: "25",
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(106250), p: 3}, // 106.250
			n2:       &DecFixedPointNumber{x: big.NewInt(425), p: 2},    // 4.25
			expected: "25",
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), p: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(4250), p: 3},  // 4.250
			expected: "25",
		},
		{
			name:     "maximum precision",
			n1:       DecFixedPoint("106.25", MaxDecPointPrecision), // 106.25
			n2:       DecFixedPoint("4.25", MaxDecPointPrecision),   // 4.25
			expected: "25",
		},
		{
			name:     "zero and maximum precision",
			n1:       DecFixedPoint("106.25", 0),                  // 106.00
			n2:       DecFixedPoint("4.25", MaxDecPointPrecision), // 4.25
			expected: "25",                                        // 106.00 / 4.25 = 24.94 -> 25
		},
		{
			name:     "maximum and zero precision",
			n1:       DecFixedPoint("106.25", MaxDecPointPrecision), // 106.25
			n2:       DecFixedPoint("4.25", 0),                      // 4.00
			expected: "26.5625",                                     // 106.25 / 4.00 = 26.5625
		},
		{
			name:     "guard digits",
			n1:       DecFixedPoint("1", 1),   // 1.0
			n2:       DecFixedPoint("1.5", 1), // 1.5
			expected: "0.7",                   // 1.0 / 1.5 = 0.66 -> 0.7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Div(tt.n2)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestDecFixedPointNumber_Cmp(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFixedPointNumber
		n2       *DecFixedPointNumber
		expected int
	}{
		{
			name:     "same precision equal",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), p: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(10625), p: 2}, // 106.25
			expected: 0,
		},
		{
			name:     "same precision less than",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), p: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(20625), p: 2}, // 206.25
			expected: -1,
		},
		{
			name:     "same precision greater than",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), p: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(625), p: 2},   // 6.25
			expected: 1,
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(106250), p: 3}, // 106.250
			n2:       &DecFixedPointNumber{x: big.NewInt(10625), p: 2},  // 106.25
			expected: 0,
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), p: 2},  // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(106250), p: 3}, // 106.250
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n1.Cmp(tt.n2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecFixedPointNumber_Abs(t *testing.T) {
	number := &DecFixedPointNumber{x: big.NewInt(-10625), p: 2} // -106.25

	expectedAbs := &DecFixedPointNumber{x: big.NewInt(10625), p: 2} // 106.25
	resultAbs := number.Abs()
	assert.Equal(t, expectedAbs.x, resultAbs.x)
	assert.Equal(t, expectedAbs.p, resultAbs.p)
}

func TestDecFixedPointNumber_Neg(t *testing.T) {
	number := &DecFixedPointNumber{x: big.NewInt(10625), p: 2} // 106.25

	expectedNeg := &DecFixedPointNumber{x: big.NewInt(-10625), p: 2} // -106.25
	resultNeg := number.Neg()
	assert.Equal(t, expectedNeg.x, resultNeg.x)
	assert.Equal(t, expectedNeg.p, resultNeg.p)
}

func TestDecFixedPointNumber_Inv(t *testing.T) {
	tests := []struct {
		name     string
		number   *DecFixedPointNumber
		expected string
	}{
		{
			name:     "0-digits",
			number:   &DecFixedPointNumber{x: big.NewInt(106), p: 0},
			expected: "0",
		},
		{
			name:     "6-digits",
			number:   &DecFixedPointNumber{x: big.NewInt(106250000), p: 6},
			expected: "0.009412",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.number.Inv()
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestDecFixedPointNumber_MarshalBinary(t *testing.T) {
	number := &DecFixedPointNumber{x: big.NewInt(10625), p: 2} // 106.25

	data, err := number.MarshalBinary()
	assert.NoError(t, err)

	expectedData := append([]byte{0, 2}, number.x.Bytes()...)
	assert.Equal(t, expectedData, data)
}

func TestDecFixedPointNumber_UnmarshalBinary(t *testing.T) {
	data := append([]byte{0, 2}, big.NewInt(10625).Bytes()...)

	number := &DecFixedPointNumber{}
	err := number.UnmarshalBinary(data)
	assert.NoError(t, err)

	expectedNumber := &DecFixedPointNumber{x: big.NewInt(10625), p: 2}
	assert.Equal(t, expectedNumber, number)
}
