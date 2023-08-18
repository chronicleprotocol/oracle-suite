package bn

import (
	"errors"
	"math/big"
)

// DecFixedPoint returns the DecFixedPoint DecFixedPointNumber of x.
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
// is already scaled by 10^n.
func DecFixedPointFromRawBigInt(x *big.Int, n uint8) *DecFixedPointNumber {
	return &DecFixedPointNumber{n: n, x: x}
}

// DecFixedPointNumber represents a fixed-point decimal number with precision.
// Internally, the number is stored as a *big.Int, scaled by 10^n.
type DecFixedPointNumber struct {
	n uint8
	x *big.Int
}

// Int returns the Int representation of the DecFixedPointNumber.
func (d *DecFixedPointNumber) Int() *IntNumber {
	return &IntNumber{x: new(big.Int).Div(d.x, decFixedPointScale(d.n))}
}

// Float returns the Float representation of the DecFixedPointNumber.
func (d *DecFixedPointNumber) Float() *FloatNumber {
	return &FloatNumber{x: d.BigFloat()}
}

// BigInt returns the *big.Int representation of the DecFixedPointNumber.
func (d *DecFixedPointNumber) BigInt() *big.Int {
	i, _ := d.BigFloat().Int(nil)
	return i
}

// RawBigInt returns the *big.Int representation of the number without scaling.
func (d *DecFixedPointNumber) RawBigInt() *big.Int {
	return d.x
}

// BigFloat returns the *big.Float representation of the DecFixedPointNumber.
func (d *DecFixedPointNumber) BigFloat() *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(d.x), new(big.Float).SetInt(decFixedPointScale(d.n)))
}

// Float64 returns the float64 representation of the DecFixedPointNumber.
func (d *DecFixedPointNumber) Float64() float64 {
	f64, _ := d.BigFloat().Float64()
	return f64
}

// Int64 returns the int64 representation of the DecFixedPointNumber.
func (d *DecFixedPointNumber) Int64() int64 {
	return d.BigInt().Int64()
}

// Uint64 returns the uint64 representation of the DecFixedPointNumber.
func (d *DecFixedPointNumber) Uint64() uint64 {
	return d.BigInt().Uint64()
}

// String returns the 10-base string representation of the DecFixedPointNumber.
func (d *DecFixedPointNumber) String() string {
	return d.BigFloat().String()
}

// Text returns the string representation of the DecFixedPointNumber.
// The format and prec arguments are the same as in big.Float.Text.
func (d *DecFixedPointNumber) Text(format byte, prec int) string {
	return d.BigFloat().Text(format, prec)
}

func (d *DecFixedPointNumber) Precision() uint8 {
	return d.n
}

func (d *DecFixedPointNumber) SetPrecision(n uint8) *DecFixedPointNumber {
	if d.n == n {
		return d
	}
	if d.x.Sign() == 0 {
		return &DecFixedPointNumber{n: n, x: big.NewInt(0)}
	}
	if d.n < n {
		return &DecFixedPointNumber{
			n: n,
			x: new(big.Int).Mul(d.x, new(big.Int).Exp(intTen, big.NewInt(int64(n-d.n)), nil)),
		}
	}
	return &DecFixedPointNumber{
		n: n,
		x: new(big.Int).Div(d.x, new(big.Int).Exp(intTen, big.NewInt(int64(d.n-n)), nil)),
	}
}

// Sign returns:
//
//	-1 if i <  0
//	 0 if i == 0
//	+1 if i >  0
func (d *DecFixedPointNumber) Sign() int {
	return d.x.Sign()
}

// Add adds x to the number and returns the result.
//
// The x argument can be any of the types accepted by DecFixedPointNumber.
func (d *DecFixedPointNumber) Add(x any) *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Add(d.x, DecFixedPoint(x, d.n).x), n: d.n}
}

// Sub subtracts x from the number and returns the result.
//
// The x argument can be any of the types accepted by DecFixedPointNumber.
func (d *DecFixedPointNumber) Sub(x any) *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Sub(d.x, DecFixedPoint(x, d.n).x), n: d.n}
}

// Mul multiplies the number by x and returns the result.
//
// The x argument can be any of the types accepted by DecFixedPointNumber.
func (d *DecFixedPointNumber) Mul(x any) *DecFixedPointNumber {
	f := DecFixedPoint(x, d.n)
	return (&DecFixedPointNumber{x: new(big.Int).Mul(d.x, f.x), n: d.n + f.n}).SetPrecision(d.n)
}

// Div divides the number by x and returns the result.
//
// Division by zero panics.
//
// The x argument can be any of the types accepted by DecFixedPointNumber.
func (d *DecFixedPointNumber) Div(x any) *DecFixedPointNumber {
	f := DecFixedPoint(x, d.n)
	if f.x.Sign() == 0 {
		panic("division by zero")
	}
	return &DecFixedPointNumber{x: new(big.Int).Mul(new(big.Int).Div(d.x, f.x), decFixedPointScale(f.n)), n: d.n}
}

// Cmp compares the number to x and returns:
//
//	-1 if i <  x
//	 0 if i == x
//	+1 if i >  x
//
// The x argument can be any of the types accepted by DecFixedPointNumber.
func (d *DecFixedPointNumber) Cmp(x any) int {
	return d.x.Cmp(DecFixedPoint(x, d.n).x)
}

// Abs returns the absolute number.
func (d *DecFixedPointNumber) Abs() *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Abs(d.x), n: d.n}
}

// Neg returns the negative number.
func (d *DecFixedPointNumber) Neg() *DecFixedPointNumber {
	return &DecFixedPointNumber{x: new(big.Int).Neg(d.x), n: d.n}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (d *DecFixedPointNumber) MarshalBinary() (data []byte, err error) {
	// Note, that changes in this function may break backward compatibility.

	b := make([]byte, 2+(d.x.BitLen()+7)/8)
	b[0] = 0 // version, reserved for future use
	b[1] = d.n
	d.x.FillBytes(b[2:])
	return b, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (d *DecFixedPointNumber) UnmarshalBinary(data []byte) error {
	// Note, that changes in this function may break backward compatibility.

	if len(data) < 2 {
		return errors.New("DecFixedPointNumber.UnmarshalBinary: invalid data length")
	}
	if data[0] != 0 {
		return errors.New("DecFixedPointNumber.UnmarshalBinary: invalid data format")
	}
	d.n = data[1]
	d.x = new(big.Int).SetBytes(data[2:])
	return nil
}
