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

// Int returns the IntNumber representation of x.
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
// - string - a string accepted by big.Int.SetString, otherwise it returns nil
// - []byte - a byte slice accepted by big.Int.SetBytes, otherwise it returns nil
//
// If the input value is not one of the supported types, nil is returned.
func Int(x any) *IntNumber {
	switch x := x.(type) {
	case IntNumber:
		return &x
	case *IntNumber:
		return x
	case FloatNumber:
		return convertFloatToInt(&x)
	case *FloatNumber:
		return convertFloatToInt(x)
	case DecFixedPointNumber:
		return convertDecFixedPointToInt(&x)
	case *DecFixedPointNumber:
		return convertDecFixedPointToInt(x)
	case DecFloatPointNumber:
		return convertDecFloatPointToInt(&x)
	case *DecFloatPointNumber:
		return convertDecFloatPointToInt(x)
	case *big.Int:
		return convertBigIntToInt(x)
	case *big.Float:
		return convertBigFloatToInt(x)
	case int, int8, int16, int32, int64:
		return convertInt64ToInt(anyToInt64(x))
	case uint, uint8, uint16, uint32, uint64:
		return convertUint64ToInt(anyToUint64(x))
	case float32, float64:
		return convertFloat64ToInt(anyToFloat64(x))
	case string:
		return convertStringToInt(x)
	case []byte:
		return convertBytesToInt(x)
	}
	return nil
}

// IntNumber represents an arbitrary-precision integer.
type IntNumber struct {
	x *big.Int
}

// Float returns the Float representation of the Int.
func (x *IntNumber) Float() *FloatNumber {
	return convertIntToFloat(x)
}

// DecFixedPoint returns the DecFixedPoint representation of the Int.
func (x *IntNumber) DecFixedPoint(n uint8) *DecFixedPointNumber {
	return convertIntToDecFixedPoint(x, n)
}

// DecFloatPoint returns the DecFloatPoint representation of the Int.
func (x *IntNumber) DecFloatPoint() *DecFloatPointNumber {
	return convertIntToDecFloatPoint(x)
}

// BigInt returns the *big.Int representation of the Int.
func (x *IntNumber) BigInt() *big.Int {
	return new(big.Int).Set(x.x)
}

// BigFloat returns the *big.Float representation of the Int.
func (x *IntNumber) BigFloat() *big.Float {
	return new(big.Float).SetInt(x.x)
}

// String returns the 10-base string representation of the Int.
func (x *IntNumber) String() string {
	return x.x.String()
}

// Text returns the string representation of the Int in the given base.
func (x *IntNumber) Text(base int) string {
	return x.x.Text(base)
}

// Sign returns:
//
//	-1 if i <  0
//	 0 if i == 0
//	+1 if i >  0
func (x *IntNumber) Sign() int {
	return x.x.Sign()
}

// Add adds y to the number and returns the result.
//
// The y argument can be any of the types accepted by Int.
func (x *IntNumber) Add(y any) *IntNumber {
	return &IntNumber{x: new(big.Int).Add(x.x, Int(y).x)}
}

// Sub subtracts y from the number and returns the result.
//
// The y argument can be any of the types accepted by Int.
func (x *IntNumber) Sub(y any) *IntNumber {
	return &IntNumber{x: new(big.Int).Sub(x.x, Int(y).x)}
}

// Mul multiplies the number by y and returns the result.
//
// The y argument can be any of the types accepted by Int.
func (x *IntNumber) Mul(y any) *IntNumber {
	return &IntNumber{x: new(big.Int).Mul(x.x, Int(y).x)}
}

// Div divides the number by y and returns the result.
//
// The y argument can be any of the types accepted by Int.
func (x *IntNumber) Div(y any) *IntNumber {
	return &IntNumber{x: new(big.Int).Div(x.x, Int(y).x)}
}

// DivRoundUp divides the number by y and returns the result rounded up.
//
// The y argument can be any of the types accepted by Int.
func (x *IntNumber) DivRoundUp(y any) *IntNumber {
	bi := Int(y)
	if new(big.Int).Rem(x.x, bi.x).Sign() > 0 {
		return &IntNumber{x: new(big.Int).Add(new(big.Int).Div(x.x, bi.x), intOne)}
	}
	return &IntNumber{x: new(big.Int).Div(x.x, bi.x)}
}

// Rem returns the remainder of the division of the number by y.
//
// The y argument can be any of the types accepted by Int.
func (x *IntNumber) Rem(y any) *IntNumber {
	return &IntNumber{x: new(big.Int).Rem(x.x, Int(y).x)}
}

// Pow returns the number raised to the power of y.
//
// The x argument can be any of the types accepted by Int.
func (x *IntNumber) Pow(y any) *IntNumber {
	return &IntNumber{x: new(big.Int).Exp(x.x, Int(y).x, nil)}
}

// Sqrt returns the square root of the number.
func (x *IntNumber) Sqrt() *IntNumber {
	return &IntNumber{x: new(big.Int).Sqrt(x.x)}
}

// Cmp compares the number to y and returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
//
// The y argument can be any of the types accepted by Int.
func (x *IntNumber) Cmp(y any) int {
	return x.x.Cmp(Int(y).x)
}

// Lsh returns the number shifted left by n bits.
func (x *IntNumber) Lsh(n uint) *IntNumber {
	return &IntNumber{x: new(big.Int).Lsh(x.x, n)}
}

// Rsh returns the number shifted right by n bits.
func (x *IntNumber) Rsh(n uint) *IntNumber {
	return &IntNumber{x: new(big.Int).Rsh(x.x, n)}
}

// Abs returns the absolute number of x.
func (x *IntNumber) Abs() *IntNumber {
	return &IntNumber{x: new(big.Int).Abs(x.x)}
}

// Neg returns the negative number of x.
func (x *IntNumber) Neg() *IntNumber {
	return &IntNumber{x: new(big.Int).Neg(x.x)}
}
