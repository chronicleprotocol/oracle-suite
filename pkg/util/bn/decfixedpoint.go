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
	"errors"
	"math/big"
	"strings"
)

// DecFixedPoint returns the DecFixedPointNumber representation of x.
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
func DecFixedPoint(x any, n uint8) *DecFixedPointNumber {
	switch x := x.(type) {
	case IntNumber:
		return convertIntToDecFixedPoint(&x, n)
	case *IntNumber:
		return convertIntToDecFixedPoint(x, n)
	case FloatNumber:
		return convertFloatToDecFixedPoint(&x, n)
	case *FloatNumber:
		return convertFloatToDecFixedPoint(x, n)
	case DecFixedPointNumber:
		return x.SetPrecision(n)
	case *DecFixedPointNumber:
		return x.SetPrecision(n)
	case DecFloatPointNumber:
		return convertDecFloatPointToDecFixedPoint(&x, n)
	case *DecFloatPointNumber:
		return convertDecFloatPointToDecFixedPoint(x, n)
	case *big.Int:
		return convertBigIntToDecFixedPoint(x, n)
	case *big.Float:
		return convertBigFloatToDecFixedPoint(x, n)
	case int, int8, int16, int32, int64:
		return convertInt64ToDecFixedPoint(anyToInt64(x), n)
	case uint, uint8, uint16, uint32, uint64:
		return convertUint64ToDecFixedPoint(anyToUint64(x), n)
	case float32, float64:
		return convertFloat64ToDecFixedPoint(anyToFloat64(x), n)
	case string:
		return convertStringToDecFixedPoint(x, n)
	}
	return nil
}

// DecFixedPointFromRawBigInt returns the DecFixedPointNumber of x assuming it
// is already scaled by 10^prec.
func DecFixedPointFromRawBigInt(x *big.Int, n uint8) *DecFixedPointNumber {
	return &DecFixedPointNumber{prec: n, x: x}
}

// DecFixedPointNumber represents a fixed-point decimal number with fixed
// precision.
//
// Internally, the number is stored as a *big.Int, scaled by 10^prec.
type DecFixedPointNumber struct {
	x    *big.Int
	prec uint8
}

// Int returns the Int representation of the DecFixedPointNumber.
func (x *DecFixedPointNumber) Int() *IntNumber {
	return convertDecFixedPointToInt(x)
}

// Float returns the Float representation of the DecFixedPointNumber.
func (x *DecFixedPointNumber) Float() *FloatNumber {
	return convertDecFixedPointToFloat(x)
}

// DecFloatPoint returns the DecFloatPointNumber representation of the
// DecFixedPointNumber.
func (x *DecFixedPointNumber) DecFloatPoint() *DecFloatPointNumber {
	return convertDecFixedPointToDecFloatPoint(x)
}

// BigInt returns the *big.Int representation of the DecFixedPointNumber.
func (x *DecFixedPointNumber) BigInt() *big.Int {
	return bigFloatToBigInt(x.BigFloat())
}

// RawBigInt returns the internal *big.Int representation of the
// DecFixedPointNumber without scaling.
func (x *DecFixedPointNumber) RawBigInt() *big.Int {
	return x.x
}

// BigFloat returns the *big.Float representation of the DecFixedPointNumber.
func (x *DecFixedPointNumber) BigFloat() *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(x.x), new(big.Float).SetInt(pow10(x.prec)))
}

// String returns the 10-base string representation of the DecFixedPointNumber.
func (x *DecFixedPointNumber) String() string {
	n := x.x.String()
	if x.prec == 0 {
		return n
	}
	s := strings.Builder{}
	if x.x.Sign() < 0 {
		s.WriteString("-")
		n = n[1:]
	}
	if len(n) <= int(x.prec) {
		s.WriteString("0.")
		s.WriteString(strings.Repeat("0", int(x.prec)-len(n)))
		s.WriteString(strings.TrimRight(n, "0")) // remove trailing zeros
	} else {
		intPart := n[:len(n)-int(x.prec)]
		fractPart := strings.TrimRight(n[len(n)-int(x.prec):], "0") // remove trailing zeros
		s.WriteString(intPart)
		if len(fractPart) > 0 {
			s.WriteString(".")
			s.WriteString(fractPart)
		}
	}
	return s.String()
}

// Text returns the string representation of the DecFixedPointNumber.
// The format and prec arguments are the same as in big.Float.Text.
//
// For any format other than 'f' and prec of -1, the result may be rounded.
func (x *DecFixedPointNumber) Text(format byte, prec int) string {
	if format == 'f' && prec < 0 {
		return x.String()
	}
	return x.BigFloat().Text(format, prec)
}

// Precision returns the precision of the DecFixedPointNumber.
//
// Precision is the number of decimal digits in the fractional part.
func (x *DecFixedPointNumber) Precision() uint8 {
	return x.prec
}

// SetPrecision returns a new DecFixedPointNumber with the given precision.
//
// Precision is the number of decimal digits in the fractional part.
func (x *DecFixedPointNumber) SetPrecision(prec uint8) *DecFixedPointNumber {
	if x.prec == prec {
		return x
	}
	if x.x.Sign() == 0 {
		return &DecFixedPointNumber{prec: prec, x: intZero}
	}
	if x.prec < prec {
		return &DecFixedPointNumber{
			prec: prec,
			x:    new(big.Int).Mul(x.x, new(big.Int).Exp(intTen, big.NewInt(int64(prec-x.prec)), nil)),
		}
	}
	return &DecFixedPointNumber{
		prec: prec,
		x:    new(big.Int).Quo(x.x, new(big.Int).Exp(intTen, big.NewInt(int64(x.prec-prec)), nil)),
	}
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
//
// The y argument can be any of the types accepted by DecFloatPointNumber.
func (x *DecFixedPointNumber) Sign() int {
	return x.x.Sign()
}

// Add adds y to the number and returns the result.
//
// The y argument can be any of the types accepted by DecFixedPointNumber.
//
// Before addition, y is converted to the precision of x.
func (x *DecFixedPointNumber) Add(y any) *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Add(x.x, DecFixedPoint(y, x.prec).x), prec: x.prec}
}

// Sub subtracts y from the number and returns the result.
//
// The y argument can be any of the types accepted by DecFixedPointNumber.
//
// Before subtraction, y is converted to the precision of x.
func (x *DecFixedPointNumber) Sub(y any) *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Sub(x.x, DecFixedPoint(y, x.prec).x), prec: x.prec}
}

// Mul multiplies the number by y and returns the result.
//
// The y argument can be any of the types accepted by DecFixedPointNumber.
//
// Before multiplication, y is converted to the precision of x.
func (x *DecFixedPointNumber) Mul(y any) *DecFixedPointNumber {
	f := DecFixedPoint(y, x.prec)
	p := int64(x.prec) + int64(f.prec)
	z := new(big.Int).Mul(x.x, f.x)
	z = new(big.Int).Quo(z, new(big.Int).Exp(intTen, big.NewInt(p-int64(x.prec)), nil))
	return &DecFixedPointNumber{x: z, prec: x.prec}
}

// Div divides the number by y and returns the result.
//
// Division by zero panics.
//
// The y argument can be any of the types accepted by DecFixedPointNumber.
//
// Before division, y is converted to the precision of x.
func (x *DecFixedPointNumber) Div(y any) *DecFixedPointNumber {
	f := DecFixedPoint(y, x.prec)
	if f.x.Sign() == 0 {
		panic("division by zero")
	}
	return &DecFixedPointNumber{x: new(big.Int).Quo(new(big.Int).Mul(x.x, pow10(x.prec)), f.x), prec: x.prec}
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
//
// The y argument can be any of the types accepted by DecFixedPointNumber.
func (x *DecFixedPointNumber) Cmp(y any) int {
	return x.x.Cmp(DecFixedPoint(y, x.prec).x)
}

// Abs returns the absolute number of x.
func (x *DecFixedPointNumber) Abs() *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Abs(x.x), prec: x.prec}
}

// Neg returns the negative number of x.
func (x *DecFixedPointNumber) Neg() *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Neg(x.x), prec: x.prec}
}

// Inv returns the inverse value of the number of x.
//
// If x is zero, Inv panics.
func (x *DecFixedPointNumber) Inv() *DecFixedPointNumber {
	if x.x.Sign() == 0 {
		panic("division by zero")
	}
	return &DecFixedPointNumber{x: new(big.Int).Quo(pow10(x.prec*2), x.x), prec: x.prec}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (x *DecFixedPointNumber) MarshalBinary() (data []byte, err error) {
	// Note, that changes in this function may break backward compatibility.

	b := make([]byte, 2+(x.x.BitLen()+7)/8)
	b[0] = 0 // version, reserved for future use
	b[1] = x.prec
	x.x.FillBytes(b[2:])
	return b, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (x *DecFixedPointNumber) UnmarshalBinary(data []byte) error {
	// Note, that changes in this function may break backward compatibility.

	if len(data) < 2 {
		return errors.New("DecFixedPointNumber.UnmarshalBinary: invalid data length")
	}
	if data[0] != 0 {
		return errors.New("DecFixedPointNumber.UnmarshalBinary: invalid data format")
	}
	x.prec = data[1]
	x.x = new(big.Int).SetBytes(data[2:])
	return nil
}
