package bn

import (
	"math/big"
)

var intOne = big.NewInt(1)

// Int returns the IntNumber representation of x.
//
// The argument x can be one of the following types:
//   - IntNumber
//   - FloatNumber - the fractional part is discarded
//   - big.Int
//   - big.Float - the fractional part is discarded
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64 - the fractional part is discarded
//   - string - a string accepted by big.Int.SetString, otherwise it returns nil
//   - []byte - big-endian representation of the integer
//
// If the input value is not one of the supported types, nil is returned.
func Int(x any) *IntNumber {
	switch x := x.(type) {
	case IntNumber:
		return &x
	case FloatNumber:
		return convertFloatNumberToInt(&x)
	case *IntNumber:
		return x
	case *FloatNumber:
		return convertFloatNumberToInt(x)
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

// String returns the 10-base string representation of the Int.
func (i *IntNumber) String() string {
	return i.x.String()
}

// Text returns the string representation of the Int in the given base.
func (i *IntNumber) Text(base int) string {
	return i.x.Text(base)
}

// Float returns the Float representation of the Int.
func (i *IntNumber) Float() *FloatNumber {
	return &FloatNumber{x: new(big.Float).SetInt(i.x)}
}

// BigInt returns the *big.Int representation of the Int.
func (i *IntNumber) BigInt() *big.Int {
	return new(big.Int).Set(i.x)
}

// Int64 returns the int64 representation of the Int.
func (i *IntNumber) Int64() int64 {
	return i.x.Int64()
}

// Uint64 returns the uint64 representation of the Int.
func (i *IntNumber) Uint64() uint64 {
	return i.x.Uint64()
}

// BigFloat returns the *big.Float representation of the Int.
func (i *IntNumber) BigFloat() *big.Float {
	return new(big.Float).SetInt(i.x)
}

// Sign returns:
//
//	-1 if i <  0
//	 0 if i == 0
//	+1 if i >  0
func (i *IntNumber) Sign() int {
	return i.x.Sign()
}

// Add adds x to the number and returns the result.
//
// The x argument can be any of the types accepted by Int.
func (i *IntNumber) Add(x any) *IntNumber {
	return &IntNumber{x: new(big.Int).Add(i.x, Int(x).x)}
}

// Sub subtracts x from the number and returns the result.
//
// The x argument can be any of the types accepted by Int.
func (i *IntNumber) Sub(x any) *IntNumber {
	return &IntNumber{x: new(big.Int).Sub(i.x, Int(x).x)}
}

// Mul multiplies the number by x and returns the result.
//
// The x argument can be any of the types accepted by Int.
func (i *IntNumber) Mul(x any) *IntNumber {
	return &IntNumber{x: new(big.Int).Mul(i.x, Int(x).x)}
}

// Div divides the number by x and returns the result.
//
// The x argument can be any of the types accepted by Int.
func (i *IntNumber) Div(x any) *IntNumber {
	return &IntNumber{x: new(big.Int).Div(i.x, Int(x).x)}
}

// DivRoundUp divides the number by x and returns the result rounded up.
//
// The x argument can be any of the types accepted by Int.
func (i *IntNumber) DivRoundUp(x any) *IntNumber {
	bi := Int(x)
	if new(big.Int).Rem(i.x, bi.x).Sign() > 0 {
		return &IntNumber{x: new(big.Int).Add(new(big.Int).Div(i.x, bi.x), intOne)}
	}
	return &IntNumber{x: new(big.Int).Div(i.x, bi.x)}
}

// Rem returns the remainder of the division of the number by x.
//
// The x argument can be any of the types accepted by Int.
func (i *IntNumber) Rem(x any) *IntNumber {
	return &IntNumber{x: new(big.Int).Rem(i.x, Int(x).x)}
}

// Pow returns the number raised to the power of x.
//
// The x argument can be any of the types accepted by Int.
func (i *IntNumber) Pow(x any) *IntNumber {
	return &IntNumber{x: new(big.Int).Exp(i.x, Int(x).x, nil)}
}

// Sqrt returns the square root of the number.
func (i *IntNumber) Sqrt() *IntNumber {
	return &IntNumber{x: new(big.Int).Sqrt(i.x)}
}

// Cmp compares the number to x and returns:
//
//	-1 if i <  x
//	 0 if i == x
//	+1 if i >  x
//
// The x argument can be any of the types accepted by Int.
func (i *IntNumber) Cmp(x any) int {
	return i.x.Cmp(Int(x).x)
}

// Lsh returns the number shifted left by n bits.
func (i *IntNumber) Lsh(n uint) *IntNumber {
	return &IntNumber{x: new(big.Int).Lsh(i.x, n)}
}

// Rsh returns the number shifted right by n bits.
func (i *IntNumber) Rsh(n uint) *IntNumber {
	return &IntNumber{x: new(big.Int).Rsh(i.x, n)}
}

// Abs returns the absolute number.
func (i *IntNumber) Abs() *IntNumber {
	return &IntNumber{x: new(big.Int).Abs(i.x)}
}

// Neg returns the negative number.
func (i *IntNumber) Neg() *IntNumber {
	return &IntNumber{x: new(big.Int).Neg(i.x)}
}
