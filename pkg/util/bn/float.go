package bn

import (
	"math/big"
)

// Float returns the FloatNumber representation of x.
//
// The argument x can be one of the following types:
//   - IntNumber
//   - FloatNumber
//   - big.Int
//   - big.Float
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - string - a string accepted by big.Float.SetString, otherwise it returns nil
//
// If the input value is not one of the supported types, nil is returned.
func Float(x any) *FloatNumber {
	switch x := x.(type) {
	case IntNumber:
		return convertIntNumberToFloat(&x)
	case FloatNumber:
		return &x
	case *IntNumber:
		return convertIntNumberToFloat(x)
	case *FloatNumber:
		return x
	case DecFixedPointNumber:
		return convertDecFixedPointToFloat(&x)
	case *DecFixedPointNumber:
		return convertDecFixedPointToFloat(x)
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
// The fractional part is discarded.
func (f *FloatNumber) Int() *IntNumber {
	return &IntNumber{x: f.BigInt()}
}

// DecFixedPoint returns the DecFixedPointNumber representation of the Float.
func (f *FloatNumber) DecFixedPoint(n uint8) *DecFixedPointNumber {
	i, _ := new(big.Float).Mul(f.x, new(big.Float).SetInt(decFixedPointScale(n))).Int(nil)
	return &DecFixedPointNumber{x: i, n: n}
}

// BigInt returns the *big.Int representation of the Float.
// The fractional part is discarded.
func (f *FloatNumber) BigInt() *big.Int {
	bi, _ := f.x.Int(nil)
	return bi
}

// BigFloat returns the *big.Float representation of the Float.
func (f *FloatNumber) BigFloat() *big.Float {
	return new(big.Float).Set(f.x)
}

// Float64 returns the float64 representation of the Float.
func (f *FloatNumber) Float64() float64 {
	f64, _ := f.x.Float64()
	return f64
}

// String returns the 10-base string representation of the Float.
func (f *FloatNumber) String() string {
	return f.x.String()
}

// Text returns the string representation of the Float.
// The format and prec arguments are the same as in big.Float.Text.
func (f *FloatNumber) Text(format byte, prec int) string {
	return f.x.Text(format, prec)
}

// Sign returns:
//
//	-1 if f <  0
//	 0 if f == 0
//	+1 if f >  0
func (f *FloatNumber) Sign() int {
	return f.x.Sign()
}

// Add adds x to the number and returns the result.
//
// The x argument can be any of the types accepted by Float.
func (f *FloatNumber) Add(x any) *FloatNumber {
	return &FloatNumber{x: new(big.Float).Add(f.x, Float(x).x)}
}

// Sub subtracts x from the number and returns the result.
//
// The x argument can be any of the types accepted by Float.
func (f *FloatNumber) Sub(x any) *FloatNumber {
	return &FloatNumber{x: new(big.Float).Sub(f.x, Float(x).x)}
}

// Mul multiplies the number by x and returns the result.
//
// The x argument can be any of the types accepted by Float.
func (f *FloatNumber) Mul(x any) *FloatNumber {
	return &FloatNumber{x: new(big.Float).Mul(f.x, Float(x).x)}
}

// Div divides the number by x and returns the result.
//
// The x argument can be any of the types accepted by Float.
func (f *FloatNumber) Div(x any) *FloatNumber {
	return &FloatNumber{x: new(big.Float).Quo(f.x, Float(x).x)}
}

// Sqrt returns the square root of the number.
func (f *FloatNumber) Sqrt() *FloatNumber {
	return &FloatNumber{x: new(big.Float).Sqrt(f.x)}
}

// Cmp compares the number with x and returns:
//
//	-1 if x <  y
//	 0 if x == y (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
//	+1 if x >  y
//
// The x argument can be any of the types accepted by Float.
func (f *FloatNumber) Cmp(x any) int {
	return f.x.Cmp(Float(x).x)
}

// Abs returns the absolute value of the number.
func (f *FloatNumber) Abs() *FloatNumber {
	return &FloatNumber{x: new(big.Float).Abs(f.x)}
}

// Neg returns the negative value of the number.
func (f *FloatNumber) Neg() *FloatNumber {
	return &FloatNumber{x: new(big.Float).Neg(f.x)}
}

// Inv returns the inverse value of the number.
func (f *FloatNumber) Inv() *FloatNumber {
	return (&FloatNumber{x: floatOne}).Div(f)
}

// IsInf reports whether the number is an infinity.
func (f *FloatNumber) IsInf() bool {
	return f.x.IsInf()
}
