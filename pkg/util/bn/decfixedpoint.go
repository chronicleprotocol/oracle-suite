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
		return x.SetPrec(n)
	case *DecFixedPointNumber:
		return x.SetPrec(n)
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
	return &DecFixedPointNumber{p: n, x: x}
}

// DecFixedPointNumber represents a fixed-point decimal number with fixed
// precision.
//
// Internally, the number is stored as a *big.Int, scaled by 10^prec.
type DecFixedPointNumber struct {
	p uint8
	x *big.Int
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
	return new(big.Float).Quo(new(big.Float).SetInt(x.x), new(big.Float).SetInt(pow10(x.p)))
}

// String returns the 10-base string representation of the DecFixedPointNumber.
func (x *DecFixedPointNumber) String() string {
	if x.x.Sign() == 0 {
		return "0"
	}
	if x.p == 0 {
		return x.x.String()
	}
	n := new(big.Int).Abs(x.x).String()
	s := strings.Builder{}
	if x.x.Sign() < 0 {
		s.WriteString("-")
	}
	if len(n) <= int(x.p) {
		s.WriteString("0.")
		s.WriteString(strings.Repeat("0", int(x.p)-len(n)))
		s.WriteString(strings.TrimRight(n, "0")) // remove trailing zeros
	} else {
		intPart := n[:len(n)-int(x.p)]
		fractPart := strings.TrimRight(n[len(n)-int(x.p):], "0") // remove trailing zeros
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

// Prec returns the precision of the DecFixedPointNumber.
//
// Prec is the number of decimal digits in the fractional part.
func (x *DecFixedPointNumber) Prec() uint8 {
	return x.p
}

// SetPrec returns a new DecFixedPointNumber with the given precision.
//
// Prec is the number of decimal digits in the fractional part.
//
// If precision is decreased, the number is rounded.
func (x *DecFixedPointNumber) SetPrec(prec uint8) *DecFixedPointNumber {
	if x.p == prec {
		return x
	}
	if x.x.Sign() == 0 {
		return &DecFixedPointNumber{p: prec, x: intZero}
	}
	return &DecFixedPointNumber{
		p: prec,
		x: bigIntSetPrec(x.x, uint32(x.p), uint32(prec)),
	}
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (x *DecFixedPointNumber) Sign() int {
	return x.x.Sign()
}

// Add adds y to the number and returns the result.
//
// Before the addition, the precision of x and y is increased to the larger of
// the two precisions. The precision of the result is set back to the precision
// of x.
func (x *DecFixedPointNumber) Add(y *DecFixedPointNumber) *DecFixedPointNumber {
	p := max(x.p, y.p)
	xi := bigIntSetPrec(x.x, uint32(x.p), uint32(p))
	yi := bigIntSetPrec(y.x, uint32(y.p), uint32(p))
	z := bigIntSetPrec(new(big.Int).Add(xi, yi), uint32(p), uint32(x.p))
	return &DecFixedPointNumber{x: z, p: x.p}
}

// Sub subtracts y from the number and returns the result.
//
// Before the subtraction, the precision of x and y is increased to the larger
// of the two precisions. The precision of the result is set back to the
// precision of x.
func (x *DecFixedPointNumber) Sub(y *DecFixedPointNumber) *DecFixedPointNumber {
	p := max(x.p, y.p)
	xi := bigIntSetPrec(x.x, uint32(x.p), uint32(p))
	yi := bigIntSetPrec(y.x, uint32(y.p), uint32(p))
	z := bigIntSetPrec(new(big.Int).Sub(xi, yi), uint32(p), uint32(x.p))
	return &DecFixedPointNumber{x: z, p: x.p}
}

// Mul multiplies the number by y and returns the result.
//
// Before the multiplication, the precision of x and y is increased to the
// sum of the precisions of x and y. The precision of the result is set back to
// the precision of x.
func (x *DecFixedPointNumber) Mul(y *DecFixedPointNumber) *DecFixedPointNumber {
	p := uint32(x.p) + uint32(y.p)
	xi := bigIntSetPrec(x.x, uint32(x.p), p)
	yi := bigIntSetPrec(y.x, uint32(y.p), p)
	z := bigIntSetPrec(new(big.Int).Mul(xi, yi), p*2, uint32(x.p))
	return &DecFixedPointNumber{x: z, p: x.p}
}

// Div divides the number by y and returns the result.
//
// Division by zero panics.
//
// Before the division, the precision of x and y is increased to the precision
// of the larger of the two values plus decGuardDigits. The precision of the
// result is set back to the precision of x.
//
// To use a different precision, use DivPrec.
func (x *DecFixedPointNumber) Div(y *DecFixedPointNumber) *DecFixedPointNumber {
	return x.DivPrec(y, uint32(max(x.p, y.p))+decGuardDigits)
}

// DivPrec divides the number by y and returns the result.
//
// Division by zero panics.
//
// Before the division, the precision of x and y is increased to the given
// precision. The precision of the result is set back to the precision of x.
func (x *DecFixedPointNumber) DivPrec(y *DecFixedPointNumber, prec uint32) *DecFixedPointNumber {
	if y.x.Sign() == 0 {
		panic("division by zero")
	}
	xi := bigIntSetPrec(x.x, uint32(x.p), prec)
	yi := bigIntSetPrec(y.x, uint32(y.p), prec)
	z := bigIntSetPrec(new(big.Int).Quo(new(big.Int).Mul(xi, pow10(prec)), yi), prec, uint32(x.p))
	return &DecFixedPointNumber{x: z, p: x.p}
}

func (x *DecFixedPointNumber) Mod(y *DecFixedPointNumber) *DecFixedPointNumber {
	if y.x.Sign() == 0 {
		panic("division by zero")
	}

	z := new(big.Int).Mod(x.x, y.x)
	n := &DecFixedPointNumber{x: z, p: x.p}
	return n
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (x *DecFixedPointNumber) Cmp(y *DecFixedPointNumber) int {
	return x.x.Cmp(bigIntSetPrec(y.x, uint32(y.p), uint32(x.p)))
}

// Abs returns the absolute number of x.
func (x *DecFixedPointNumber) Abs() *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Abs(x.x), p: x.p}
}

// Neg returns the negative number of x.
func (x *DecFixedPointNumber) Neg() *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Neg(x.x), p: x.p}
}

// Inv returns the inverse value of the number of x.
//
// If x is zero, Inv panics.
func (x *DecFixedPointNumber) Inv() *DecFixedPointNumber {
	if x.x.Sign() == 0 {
		panic("division by zero")
	}
	return &DecFixedPointNumber{x: bigIntDivRound(pow10(uint32(x.p)*2), x.x), p: x.p}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (x *DecFixedPointNumber) MarshalBinary() (data []byte, err error) {
	// Note, that changes in this function may break backward compatibility.

	b := make([]byte, 2+(x.x.BitLen()+7)/8)
	b[0] = 0 // version, reserved for future use
	b[1] = x.p
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
	x.p = data[1]
	x.x = new(big.Int).SetBytes(data[2:])
	return nil
}

func bigIntDivRound(x, y *big.Int) *big.Int {
	quo, rem := new(big.Int).QuoRem(x, y, new(big.Int))
	if rem.Sign() == 0 {
		return quo
	}
	if rem.Cmp(new(big.Int).Quo(y, big.NewInt(2))) >= 0 {
		return new(big.Int).Add(quo, big.NewInt(1))
	}
	return quo
}

func bigIntSetPrec(x *big.Int, currPrec, newPrec uint32) *big.Int {
	if currPrec == newPrec || x.Sign() == 0 {
		return x
	}
	if currPrec < newPrec {
		return new(big.Int).Mul(x, new(big.Int).Exp(intTen, big.NewInt(int64(newPrec-currPrec)), nil))
	}
	return bigIntDivRound(x, new(big.Int).Exp(intTen, big.NewInt(int64(currPrec-newPrec)), nil))
}
