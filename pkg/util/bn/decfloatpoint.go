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
// adjusted dynamically to fit the value.
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

// Text returns the string representation of the DecFixedPointNumber.
// The format and prec arguments are the same as in big.Float.Text.
//
// For any format other than 'f' and prec of -1, the result may be rounded.
func (x *DecFloatPointNumber) Text(format byte, prec int) string {
	return x.x.Text(format, prec)
}

// Prec returns the precision of the DecFloatPointNumber.
//
// Prec is the number of decimal digits in the fractional part.
func (x *DecFloatPointNumber) Prec() uint8 {
	return x.x.Prec()
}

// SetPrec returns a new DecFloatPointNumber with the given precision.
//
// Prec is the number of decimal digits in the fractional part.
func (x *DecFloatPointNumber) SetPrec(n uint8) *DecFloatPointNumber {
	if n == x.x.p {
		return x
	}
	return &DecFloatPointNumber{x: x.x.SetPrec(n)}
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
func (x *DecFloatPointNumber) Add(y *DecFloatPointNumber) *DecFloatPointNumber {
	p := max(x.x.p, y.x.p)
	xi := bigIntSetPrec(x.x.x, uint32(x.x.p), uint32(p))
	yi := bigIntSetPrec(y.x.x, uint32(y.x.p), uint32(p))
	n := &DecFloatPointNumber{x: &DecFixedPointNumber{x: new(big.Int).Add(xi, yi), p: p}}
	n.adjustPrec()
	return n
}

// Sub subtracts y from the number and returns the result.
func (x *DecFloatPointNumber) Sub(y *DecFloatPointNumber) *DecFloatPointNumber {
	p := max(x.x.p, y.x.p)
	xi := bigIntSetPrec(x.x.x, uint32(x.x.p), uint32(p))
	yi := bigIntSetPrec(y.x.x, uint32(y.x.p), uint32(p))
	n := &DecFloatPointNumber{x: &DecFixedPointNumber{x: new(big.Int).Sub(xi, yi), p: p}}
	n.adjustPrec()
	return n
}

// Mul multiplies the number by y and returns the result.
func (x *DecFloatPointNumber) Mul(y *DecFloatPointNumber) *DecFloatPointNumber {
	wp := uint32(x.x.p) + uint32(y.x.p) // working precision
	rp := wp                            // result precision
	if wp > MaxDecPointPrecision {
		rp = MaxDecPointPrecision
	}
	xi := bigIntSetPrec(x.x.x, uint32(x.x.p), wp)
	yi := bigIntSetPrec(y.x.x, uint32(y.x.p), wp)
	z := bigIntSetPrec(new(big.Int).Mul(xi, yi), wp*2, rp)
	n := &DecFloatPointNumber{x: &DecFixedPointNumber{x: z, p: uint8(rp)}}
	n.adjustPrec()
	return n
}

// MulDownFixed multiplies the number y and deflates prec precision
func (x *DecFloatPointNumber) MulDownFixed(y *DecFloatPointNumber, prec uint8) *DecFloatPointNumber {
	return x.Mul(y).Deflate(prec)
}

// MulUpFixed multiplies the number y up and deflates prec precision
func (x *DecFloatPointNumber) MulUpFixed(y *DecFloatPointNumber, prec uint8) *DecFloatPointNumber {
	ret := x.Mul(y)
	if ret.Sign() == 0 {
		return ret
	}
	one := DecFloatPoint(intOne)
	return ret.Sub(one).Deflate(prec).Add(one)
}

// Div divides the number by y and returns the result.
//
// Division by zero panics.
//
// During division, the precision is increased by divPrecisionIncrease and then
// lowered to the smallest possible that still fits the result.
func (x *DecFloatPointNumber) Div(y *DecFloatPointNumber) *DecFloatPointNumber {
	if y.x.Sign() == 0 {
		panic("division by zero")
	}
	p := max(x.x.p, y.x.p)
	wp := uint32(p) + divPrecisionIncrease + decGuardDigits // working precision
	rp := uint32(p) + divPrecisionIncrease                  // result precision
	if rp > MaxDecPointPrecision {
		rp = MaxDecPointPrecision
	}
	n := x.DivPrec(y, wp)
	n.x.x = bigIntSetPrec(n.x.x, uint32(n.x.p), rp)
	n.x.p = uint8(rp)
	n.adjustPrec()
	return n
}

// DivUp divides the number y up and return the result.
func (x *DecFloatPointNumber) DivUp(y *DecFloatPointNumber) *DecFloatPointNumber {
	if x.Sign() == 0 {
		return x
	}
	one := DecFloatPoint(intOne)
	// The traditional divUp formula is:
	// divUp(x, y) := (x + y - 1) / y
	// To avoid intermediate overflow in the addition, we distribute the division and get:
	// divUp(x, y) := (x - 1) / y + 1
	// Note that this requires x != 0, if x == 0 then the result is zero
	return x.Sub(one).DivPrec(y, uint32(x.x.p)).Add(one)
}

// DivUpFixed divides the number y up and deflates prec precision
func (x *DecFloatPointNumber) DivUpFixed(y *DecFloatPointNumber, prec uint8) *DecFloatPointNumber {
	if x.Sign() == 0 {
		return x
	}
	one := DecFloatPoint(intOne)
	return x.Inflate(prec).Sub(one).DivPrec(y, uint32(x.x.p)).Add(one)
}

// DivDownFixed inflates prec precision and divides the number y down
func (x *DecFloatPointNumber) DivDownFixed(y *DecFloatPointNumber, prec uint8) *DecFloatPointNumber {
	if x.Sign() == 0 {
		return x
	}
	return x.Inflate(prec).DivPrec(y, uint32(x.x.p))
}

// Exp exponential function by y and return the x ^ y.
func (x *DecFloatPointNumber) Exp(y *DecFloatPointNumber) *DecFloatPointNumber {
	if x.x.Sign() == 0 {
		return x
	}
	if x.Cmp(DecFloatPoint(1)) == 0 {
		return DecFloatPoint(1)
	}
	z := new(big.Int).Exp(x.x.x, y.x.x, nil)
	n := &DecFloatPointNumber{x: &DecFixedPointNumber{x: z, p: x.x.p}}
	n.adjustPrec()
	return n
}

// Inflate inflates the number by prec precision and return the result.
// return = x * (10 ^ prec)
// Do not change precision.
func (x *DecFloatPointNumber) Inflate(prec uint8) *DecFloatPointNumber {
	if prec == 0 {
		return x
	}
	p := uint32(x.x.p) + uint32(prec)
	z := bigIntSetPrec(x.x.x, uint32(x.x.p), p)
	return &DecFloatPointNumber{x: &DecFixedPointNumber{x: z, p: x.x.p}}
}

// Deflate deflates the number by prec precision and return the results.
// return = x / (10 ^ prec)
// Do not change precision. If x < 10 ^ prec, return 0
func (x *DecFloatPointNumber) Deflate(prec uint8) *DecFloatPointNumber {
	z := bigIntSetPrec(x.x.x, uint32(x.x.p)+uint32(prec), uint32(x.x.p))
	n := &DecFloatPointNumber{x: &DecFixedPointNumber{x: z, p: x.x.p}}
	n.adjustPrec()
	return n
}

// DivPrec divides the number by y and returns the result.
//
// Division by zero panics.
//
// During division, the precision is set to prec and then lowered to the
// smallest possible that still fits the result.
func (x *DecFloatPointNumber) DivPrec(y *DecFloatPointNumber, prec uint32) *DecFloatPointNumber {
	if y.x.Sign() == 0 {
		panic("division by zero")
	}
	wp := prec // working precision
	rp := prec // result precision
	if rp > MaxDecPointPrecision {
		rp = MaxDecPointPrecision
	}
	xi := bigIntSetPrec(x.x.x, uint32(x.x.p), wp)
	yi := bigIntSetPrec(y.x.x, uint32(y.x.p), wp)
	z := bigIntSetPrec(new(big.Int).Quo(new(big.Int).Mul(xi, pow10(prec)), yi), wp, rp)
	n := &DecFloatPointNumber{x: &DecFixedPointNumber{x: z, p: uint8(rp)}}
	n.adjustPrec()
	return n
}

// Cmp compares the number to y and returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (x *DecFloatPointNumber) Cmp(y *DecFloatPointNumber) int {
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
//
// During inversion, the precision is increased by divPrecisionIncrease and
// then lowered to the smallest possible that still fits the result.
func (x *DecFloatPointNumber) Inv() *DecFloatPointNumber {
	if x.x.Sign() == 0 {
		panic("division by zero")
	}
	wp := uint32(x.x.p) + divPrecisionIncrease + decGuardDigits // working precision
	rp := uint32(x.x.p) + divPrecisionIncrease                  // result precision
	if rp > MaxDecPointPrecision {
		rp = MaxDecPointPrecision
	}
	z := bigIntSetPrec(x.x.x, uint32(x.x.p), wp)
	z = bigIntSetPrec(new(big.Int).Quo(pow10(wp*2), z), wp, rp)
	n := &DecFloatPointNumber{x: &DecFixedPointNumber{x: z, p: uint8(rp)}}
	n.adjustPrec()
	return n
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

// adjustPrec sets the precision to the lowest possible value that still fits
// the number.
func (x *DecFloatPointNumber) adjustPrec() {
	if x.x.Sign() == 0 {
		x.x.x = intZero
		x.x.p = 0
		return
	}
	s := x.x.x.String()
	z := uint8(0)
	for i := len(s) - 1; ; i-- {
		if i < 0 {
			z = x.x.p // zero number
			break
		}
		if z >= x.x.p {
			break // all decimals are zeros
		}
		if s[i] != '0' {
			break
		}
		z++
	}
	p := x.x.p - z
	if x.x.p == p {
		return
	}
	x.x.x = bigIntSetPrec(x.x.x, uint32(x.x.p), uint32(p))
	x.x.p = p
}
