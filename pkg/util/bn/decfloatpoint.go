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
)

// DecFloatPoint returns the DecFloatPointNumber representation of x.
//
// The argument x can be one of the following types:
// - IntNumber
// - FloatNumber
// - DecFixedPointNumber
// - DecFloatPointNumber
// - big.Int
// - big.Float
// - int, int8, int16, int32, int64
// - uint, uint8, uint16, uint32, uint64
// - float32, float64
// - string - a string accepted by big.Float.SetString, otherwise it returns nil
//
// If the input value is not one of the supported types, nil is returned.
func DecFloatPoint(x any) *DecFloatPointNumber {
	switch x := x.(type) {
	case IntNumber:
		return convertIntToDecFloatPoint(&x)
	case *IntNumber:
		return convertIntToDecFloatPoint(x)
	case FloatNumber:
		return convertFloatToDecFloatPoint(&x)
	case *FloatNumber:
		return convertFloatToDecFloatPoint(x)
	case DecFixedPointNumber:
		return convertDecFixedPointToDecFloatPoint(&x)
	case *DecFixedPointNumber:
		return convertDecFixedPointToDecFloatPoint(x)
	case DecFloatPointNumber:
		return &x
	case *DecFloatPointNumber:
		return x
	case *big.Int:
		return convertBigIntToDecFloatPoint(x)
	case *big.Float:
		return convertBigFloatToDecFloatPoint(x)
	case int, int8, int16, int32, int64:
		return convertInt64ToDecFloatPoint(anyToInt64(x))
	case uint, uint8, uint16, uint32, uint64:
		return convertUint64ToDecFloatPoint(anyToUint64(x))
	case float32, float64:
		return convertFloat64ToDecFloatPoint(anyToFloat64(x))
	case string:
		return convertStringToDecFloatPoint(x)
	}
	return nil
}

// DecFloatPointNumber represents a decimal floating-point number.
//
// Unlike the DecFixedPointNumber, the precision of the DecFloatPointNumber is
// not fixed. The precision is dynamically adjusted to fit the number.
type DecFloatPointNumber struct {
	x *DecFixedPointNumber
}

// Int returns the Int representation of the DecFloatPointNumber.
func (x *DecFloatPointNumber) Int() *IntNumber {
	return convertDecFloatPointToInt(x)
}

// DecFixedPoint returns the DecFixedPointNumber representation of the Float.
func (x *DecFloatPointNumber) DecFixedPoint(n uint8) *DecFixedPointNumber {
	return convertDecFloatPointToDecFixedPoint(x, n)
}

// Float returns the Float representation of the DecFloatPointNumber.
func (x *DecFloatPointNumber) Float() *FloatNumber {
	return convertDecFloatPointToFloat(x)
}

// BigInt returns the *big.Int representation of the DecFloatPointNumber.
func (x *DecFloatPointNumber) BigInt() *big.Int {
	return x.x.BigInt()
}

// BigFloat returns the *big.Float representation of the DecFloatPointNumber.
func (x *DecFloatPointNumber) BigFloat() *big.Float {
	return x.x.BigFloat()
}

// String returns the 10-base string representation of the DecFloatPointNumber.
func (x *DecFloatPointNumber) String() string {
	return x.x.String()
}

// Text returns the string representation of the DecFloatPointNumber.
// The format and prec arguments are the same as in big.Float.Text.
//
// For any format other than 'f' and prec of -1, the result may be rounded.
func (x *DecFloatPointNumber) Text(format byte, prec int) string {
	if format == 'f' && prec < 0 {
		return x.x.String()
	}
	return x.x.Text(format, prec)
}

// Precision returns the precision of the DecFloatPointNumber.
//
// Precision is the number of decimal digits in the fractional part.
func (x *DecFloatPointNumber) Precision() uint8 {
	return x.x.Precision()
}

// SetPrecision returns a new DecFloatPointNumber with the given precision.
//
// Precision is the number of decimal digits in the fractional part.
func (x *DecFloatPointNumber) SetPrecision(n uint8) *DecFloatPointNumber {
	if n == x.x.prec {
		return x
	}
	return &DecFloatPointNumber{x: x.x.SetPrecision(n)}
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (x *DecFloatPointNumber) Sign() int {
	return x.x.Sign()
}

// Add adds y to the number and returns the result.
//
// The y argument can be any of the types accepted by DecFloatPointNumber.
//
// The precision is adjusted to fit the result.
func (x *DecFloatPointNumber) Add(y any) *DecFloatPointNumber {
	f := DecFloatPoint(y)
	p := x.x.prec
	if f.x.prec > p {
		p = f.x.prec
	}
	a := x.x.SetPrecision(p)
	b := f.x.SetPrecision(p)
	return (&DecFloatPointNumber{x: a.Add(b)}).setMinimumPrecision()
}

// Sub subtracts y from the number and returns the result.
//
// The y argument can be any of the types accepted by DecFloatPointNumber.
//
// The precision is adjusted to fit the result.
func (x *DecFloatPointNumber) Sub(y any) *DecFloatPointNumber {
	f := DecFloatPoint(y)
	p := x.x.prec
	if f.x.prec > p {
		p = f.x.prec
	}
	a := x.x.SetPrecision(p)
	b := f.x.SetPrecision(p)
	return (&DecFloatPointNumber{x: a.Sub(b)}).setMinimumPrecision()
}

// Mul multiplies the number by y and returns the result.
//
// The y argument can be any of the types accepted by DecFloatPointNumber.
//
// The precision is adjusted to fit the result.
func (x *DecFloatPointNumber) Mul(y any) *DecFloatPointNumber {
	f := DecFloatPoint(y)
	px := int(f.x.prec)
	py := int(x.x.prec)
	p := px + py
	if p > math.MaxUint8 {
		p = math.MaxUint8
	}
	a := x.x.SetPrecision(uint8(p))
	b := f.x.SetPrecision(uint8(p))
	return (&DecFloatPointNumber{x: a.Mul(b)}).setMinimumPrecision()
}

// Div divides the number by y and returns the result.
//
// Division by zero panics.
//
// The y argument can be any of the types accepted by DecFloatPointNumber.
//
// The precision is adjusted to fit the result.
func (x *DecFloatPointNumber) Div(y any) *DecFloatPointNumber {
	f := DecFloatPoint(y)
	a := x.x.SetPrecision(math.MaxUint8)
	b := f.x.SetPrecision(math.MaxUint8)
	return (&DecFloatPointNumber{x: a.Div(b)}).setMinimumPrecision()
}

// Cmp compares the number to y and returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
//
// The y argument can be any of the types accepted by DecFloatPointNumber.
func (x *DecFloatPointNumber) Cmp(y any) int {
	return x.x.Cmp(DecFloatPoint(y).x)
}

// Abs returns the absolute number of x.
func (x *DecFloatPointNumber) Abs() *DecFloatPointNumber {
	return &DecFloatPointNumber{x: x.x.Abs()}
}

// Neg returns the negative number of x.
func (x *DecFloatPointNumber) Neg() *DecFloatPointNumber {
	return &DecFloatPointNumber{x: x.x.Neg()}
}

// Inv returns the inverse value of the number of x.
//
// If x is zero, Inv panics.
func (x *DecFloatPointNumber) Inv() *DecFloatPointNumber {
	if x.x.Sign() == 0 {
		panic("division by zero")
	}
	i := x.Float().Inv()
	p := stringNumberDecimalPrecision(i.Text('f', -1))
	if p > math.MaxUint8 {
		p = math.MaxUint8
	}
	i = i.Mul(pow10(p))
	return &DecFloatPointNumber{x: &DecFixedPointNumber{x: i.BigInt(), prec: uint8(p)}}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (x *DecFloatPointNumber) MarshalBinary() (data []byte, err error) {
	return x.x.MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (x *DecFloatPointNumber) UnmarshalBinary(data []byte) error {
	x.x = new(DecFixedPointNumber)
	return x.x.UnmarshalBinary(data)
}

// setMinimumPrecision sets the precision to the lowest possible precision
// required to represent the number.
func (x *DecFloatPointNumber) setMinimumPrecision() *DecFloatPointNumber {
	str := x.x.x.String()
	tz := 0
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == '0' {
			tz++
		} else {
			break
		}
	}
	prec := int(x.x.prec) - tz
	if prec < 0 {
		prec = 0
	}
	if prec > math.MaxUint8 {
		prec = math.MaxUint8
	}
	x.x = x.x.SetPrecision(uint8(prec))
	return x
}
