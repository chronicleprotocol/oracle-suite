package bn

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

var intOne = big.NewInt(1)

type Int struct {
	x *big.Int
}

func IntFromString(s string, base int) *Int {
	x, ok := new(big.Int).SetString(s, base)
	if !ok {
		return nil
	}
	return &Int{x: x}
}

func IntFromInt64(x int64) *Int {
	return &Int{x: big.NewInt(x)}
}

func IntFromUint64(x uint64) *Int {
	return &Int{x: new(big.Int).SetUint64(x)}
}

func IntFromFloat64(x float64) *Int {
	i, _ := big.NewFloat(x).Int(nil)
	return &Int{x: i}
}

func IntFromBigInt(x *big.Int) *Int {
	return &Int{x: x}
}

func IntFromBigFloat(x *big.Float) *Int {
	bi, _ := x.Int(nil)
	return &Int{x: bi}
}

func IntFromFloat(x *Float) *Int {
	return &Int{x: x.BigInt()}
}

func IntFromBytes(x []byte) *Int {
	return &Int{x: new(big.Int).SetBytes(x)}
}

func (i *Int) String() string {
	return i.x.String()
}

func (i *Int) Text(base int) string {
	return i.x.Text(base)
}

func (i *Int) Float() *Float {
	return FloatFromInt(i)
}

func (i *Int) BigInt() *big.Int {
	return i.x
}

func (i *Int) Int64() int64 {
	return i.x.Int64()
}

func (i *Int) Uint64() uint64 {
	return i.x.Uint64()
}

func (i *Int) BigFloat() *big.Float {
	return new(big.Float).SetInt(i.x)
}

func (i *Int) Sign() int {
	return i.x.Sign()
}

func (i *Int) Add(x *Int) *Int {
	return &Int{x: new(big.Int).Add(i.x, x.x)}
}

func (i *Int) Sub(x *Int) *Int {
	return &Int{x: new(big.Int).Sub(i.x, x.x)}
}

func (i *Int) Mul(x *Int) *Int {
	return &Int{x: new(big.Int).Mul(i.x, x.x)}
}

func (i *Int) Div(x *Int) *Int {
	return &Int{x: new(big.Int).Div(i.x, x.x)}
}

func (i *Int) DivRoundUp(x *Int) *Int {
	if new(big.Int).Rem(i.x, x.x).Sign() > 0 {
		return &Int{x: new(big.Int).Add(new(big.Int).Div(i.x, x.x), intOne)}
	}
	return &Int{x: new(big.Int).Div(i.x, x.x)}
}

func (i *Int) Rem(x *Int) *Int {
	return &Int{x: new(big.Int).Rem(i.x, x.x)}
}

func (i *Int) Pow(x *Int) *Int {
	return &Int{x: new(big.Int).Exp(i.x, x.x, nil)}
}

func (i *Int) Sqrt() *Int {
	return &Int{x: new(big.Int).Sqrt(i.x)}
}

func (i *Int) Cmp(x *Int) int {
	return i.x.Cmp(x.x)
}

func (i *Int) Lsh(n uint) *Int {
	return &Int{x: new(big.Int).Lsh(i.x, n)}
}

func (i *Int) Rsh(n uint) *Int {
	return &Int{x: new(big.Int).Rsh(i.x, n)}
}

func (i *Int) Abs() *Int {
	return &Int{x: new(big.Int).Abs(i.x)}
}

func (i *Int) Neg() *Int {
	return &Int{x: new(big.Int).Neg(i.x)}
}

func (i *Int) Inv() *Float {
	return FloatFromFloat64(1).Div(i.Float())
}

// GobEncode implements the gob.GobEncoder interface.
func (i *Int) GobEncode() ([]byte, error) {
	return i.x.GobEncode()
}

// GobDecode implements the gob.GobDecoder interface.
func (i *Int) GobDecode(b []byte) error {
	i.x = new(big.Int)
	return i.x.GobDecode(b)
}

// Value implements the driver.Valuer interface.
func (i *Int) Value() (driver.Value, error) {
	return i.GobEncode()
}

// Scan implements the sql.Scanner interface.
func (i *Int) Scan(v interface{}) error {
	switch v := v.(type) {
	case nil:
		return fmt.Errorf("nil")
	case []byte:
		return i.GobDecode(v)
	case string:
		return i.GobDecode([]byte(v))
	}
	return fmt.Errorf("unsupported type: %T", v)
}
