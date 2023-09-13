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

// Float returns the FloatNumber representation of x.
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
func Float(x any) *FloatNumber {
	switch x := x.(type) {
	case IntNumber:
		return convertIntToFloat(&x)
	case FloatNumber:
		return &x
	case *IntNumber:
		return convertIntToFloat(x)
	case *FloatNumber:
		return x
	case DecFixedPointNumber:
		return convertDecFixedPointToFloat(&x)
	case *DecFixedPointNumber:
		return convertDecFixedPointToFloat(x)
	case DecFloatPointNumber:
		return convertDecFloatPointToFloat(&x)
	case *DecFloatPointNumber:
		return convertDecFloatPointToFloat(x)
	case *big.Int:
		return convertBigIntToFloat(x)
	case *big.Float:
		return convertBigFloatToFloat(x)
	case int, int8, int16, int32, int64:
		return convertInt64ToFloat(anyToInt64(x))
	case uint, uint8, uint16, uint32, uint64:
		return convertUint64ToFloat(anyToUint64(x))
	case float32, float64:
		return convertFloat64ToFloat(anyToFloat64(x))
	case string:
		return convertStringToFloat(x)
	}
	return nil
}

// FloatNumber represents a floating-point number.
type FloatNumber struct {
	x *big.Float
}

// Int returns the IntNumber representation of the Float.
//
// The fractional part is discarded and the number is rounded.
func (x *FloatNumber) Int() *IntNumber {
	return convertFloatToInt(x)
}

// DecFixedPoint returns the DecFixedPointNumber representation of the Float.
func (x *FloatNumber) DecFixedPoint(n uint8) *DecFixedPointNumber {
	return convertFloatToDecFixedPoint(x, n)
}

// BigInt returns the *big.Int representation of the Float.
//
// The fractional part is discarded and the number is rounded.
func (x *FloatNumber) BigInt() *big.Int {
	return bigFloatToBigInt(x.x)
}

// BigFloat returns the *big.Float representation of the Float.
func (x *FloatNumber) BigFloat() *big.Float {
	return new(big.Float).Set(x.x)
}

// String returns the 10-base string representation of the Float.
func (x *FloatNumber) String() string {
	return x.x.String()
}

// Text returns the string representation of the Float.
// The format and prec arguments are the same as in big.Float.Text.
func (x *FloatNumber) Text(format byte, prec int) string {
	return x.x.Text(format, prec)
}

// Precision returns the precision of the Float.
//
// It is wrapper around big.Float.Prec.
func (x *FloatNumber) Precision() uint {
	return x.x.Prec()
}

// SetPrecision sets the precision of the Float.
//
// It is wrapper around big.Float.SetPrec.
func (x *FloatNumber) SetPrecision(prec uint) *FloatNumber {
	x.x.SetPrec(prec)
	return x
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (x *FloatNumber) Sign() int {
	return x.x.Sign()
}

// Add adds y to the number and returns the result.
func (x *FloatNumber) Add(y *FloatNumber) *FloatNumber {
	return &FloatNumber{x: new(big.Float).Add(x.x, y.x)}
}

// Sub subtracts y from the number and returns the result.
func (x *FloatNumber) Sub(y *FloatNumber) *FloatNumber {
	return &FloatNumber{x: new(big.Float).Sub(x.x, y.x)}
}

// Mul multiplies the number by y and returns the result.
func (x *FloatNumber) Mul(y *FloatNumber) *FloatNumber {
	return &FloatNumber{x: new(big.Float).Mul(x.x, y.x)}
}

// Div divides the number by y and returns the result.
func (x *FloatNumber) Div(y *FloatNumber) *FloatNumber {
	return &FloatNumber{x: new(big.Float).Quo(x.x, y.x)}
}

// Sqrt returns the square root of the number.
func (x *FloatNumber) Sqrt() *FloatNumber {
	return &FloatNumber{x: new(big.Float).Sqrt(x.x)}
}

// Cmp compares the number to y and returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (x *FloatNumber) Cmp(y *FloatNumber) int {
	return x.x.Cmp(y.x)
}

// Abs returns the absolute number of x.
func (x *FloatNumber) Abs() *FloatNumber {
	return &FloatNumber{x: new(big.Float).Abs(x.x)}
}

// Neg returns the negative number of x.
func (x *FloatNumber) Neg() *FloatNumber {
	return &FloatNumber{x: new(big.Float).Neg(x.x)}
}

// Inv returns the inverse number of x.
func (x *FloatNumber) Inv() *FloatNumber {
	return (&FloatNumber{x: floatOne}).Div(x)
}

// IsInf reports whether the number is an infinity.
func (x *FloatNumber) IsInf() bool {
	return x.x.IsInf()
}
