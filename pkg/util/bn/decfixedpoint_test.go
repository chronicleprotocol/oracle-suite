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
			expected: &DecFixedPointNumber{x: big.NewInt(42), n: 0},
		},
		{
			name:     "Pointer to IntNumber",
			input:    &IntNumber{big.NewInt(42)},
			prec:     0,
			expected: &DecFixedPointNumber{x: big.NewInt(42), n: 0},
		},
		{
			name:     "FloatNumber",
			input:    FloatNumber{big.NewFloat(42.5)},
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), n: 2},
		},
		{
			name:     "Pointer to FloatNumber",
			input:    &FloatNumber{big.NewFloat(42.5)},
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), n: 2},
		},
		{
			name:     "big.Int",
			input:    big.NewInt(42),
			prec:     0,
			expected: &DecFixedPointNumber{x: big.NewInt(42), n: 0},
		},
		{
			name:     "big.Float",
			input:    big.NewFloat(42.5),
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), n: 2},
		},
		{
			name:     "int",
			input:    int(42),
			prec:     0,
			expected: &DecFixedPointNumber{x: big.NewInt(42), n: 0},
		},
		{
			name:     "float64",
			input:    float64(42.5),
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), n: 2},
		},
		{
			name:     "string",
			input:    "42.5",
			prec:     2,
			expected: &DecFixedPointNumber{x: big.NewInt(4250), n: 2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DecFixedPoint(test.input, test.prec)
			if test.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, test.expected.String(), result.String())
				assert.Equal(t, test.expected.Precision(), result.Precision())
			}
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
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), n: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(225), n: 2},  // 2.25
			expected: "12.75",
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10500), n: 3}, // 10.500
			n2:       &DecFixedPointNumber{x: big.NewInt(225), n: 2},   // 2.25
			expected: "12.75",
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), n: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(2250), n: 3}, // 2.250
			expected: "12.75",
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
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), n: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(225), n: 2},  // 2.25
			expected: "8.25",
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10500), n: 3}, // 10.500
			n2:       &DecFixedPointNumber{x: big.NewInt(225), n: 2},   // 2.25
			expected: "8.25",
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), n: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(2250), n: 3}, // 2.250
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

func TestDecFixedPointNumber_Mul(t *testing.T) {
	tests := []struct {
		name     string
		n1       *DecFixedPointNumber
		n2       *DecFixedPointNumber
		expected string
	}{
		{
			name:     "same precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), n: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(225), n: 2},  // 2.25
			expected: "23.62",
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10500), n: 3}, // 10.500
			n2:       &DecFixedPointNumber{x: big.NewInt(225), n: 2},   // 2.25
			expected: "23.625",
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(1050), n: 2}, // 10.50
			n2:       &DecFixedPointNumber{x: big.NewInt(2250), n: 3}, // 2.250
			expected: "23.62",
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
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), n: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(425), n: 2},   // 4.25
			expected: "25",
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(106250), n: 3}, // 106.250
			n2:       &DecFixedPointNumber{x: big.NewInt(425), n: 2},    // 4.25
			expected: "25",
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), n: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(4250), n: 3},  // 4.250
			expected: "25",
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
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), n: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(10625), n: 2}, // 106.25
			expected: 0,
		},
		{
			name:     "same precision less than",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), n: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(20625), n: 2}, // 206.25
			expected: -1,
		},
		{
			name:     "same precision greater than",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), n: 2}, // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(625), n: 2},   // 6.25
			expected: 1,
		},
		{
			name:     "first higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(106250), n: 3}, // 106.250
			n2:       &DecFixedPointNumber{x: big.NewInt(10625), n: 2},  // 106.25
			expected: 0,
		},
		{
			name:     "second higher precision",
			n1:       &DecFixedPointNumber{x: big.NewInt(10625), n: 2},  // 106.25
			n2:       &DecFixedPointNumber{x: big.NewInt(106250), n: 3}, // 106.250
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
	number := &DecFixedPointNumber{x: big.NewInt(-10625), n: 2} // -106.25

	expectedAbs := &DecFixedPointNumber{x: big.NewInt(10625), n: 2} // 106.25
	resultAbs := number.Abs()
	assert.Equal(t, expectedAbs.x, resultAbs.x)
	assert.Equal(t, expectedAbs.n, resultAbs.n)
}

func TestDecFixedPointNumber_Neg(t *testing.T) {
	number := &DecFixedPointNumber{x: big.NewInt(10625), n: 2} // 106.25

	expectedNeg := &DecFixedPointNumber{x: big.NewInt(-10625), n: 2} // -106.25
	resultNeg := number.Neg()
	assert.Equal(t, expectedNeg.x, resultNeg.x)
	assert.Equal(t, expectedNeg.n, resultNeg.n)
}

func TestDecFixedPointNumber_MarshalBinary(t *testing.T) {
	number := &DecFixedPointNumber{x: big.NewInt(10625), n: 2} // 106.25

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

	expectedNumber := &DecFixedPointNumber{x: big.NewInt(10625), n: 2}
	assert.Equal(t, expectedNumber, number)
}
