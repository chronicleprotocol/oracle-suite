package bn

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

type Float struct {
	x *big.Float
}

func FloatFromString(s string) *Float {
	x, ok := new(big.Float).SetString(s)
	if !ok {
		return nil
	}
	return &Float{x: x}
}

func FloatFromFloat64(x float64) *Float {
	return &Float{x: big.NewFloat(x)}
}

func FloatFromInt64(x int64) *Float {
	return &Float{x: new(big.Float).SetInt64(x)}
}

func FloatFromUint64(x uint64) *Float {
	return &Float{x: new(big.Float).SetUint64(x)}
}

func FloatFromBigFloat(x *big.Float) *Float {
	return &Float{x: x}
}

func FloatFromBigInt(x *big.Int) *Float {
	return &Float{x: new(big.Float).SetInt(x)}
}

func FloatFromInt(x *Int) *Float {
	return &Float{x: x.BigFloat()}
}

func (f *Float) String() string {
	return f.x.String()
}

func (f *Float) Text(format byte, prec int) string {
	return f.x.Text(format, prec)
}

func (f *Float) Int() *Int {
	return IntFromFloat(f)
}

func (f *Float) BigFloat() *big.Float {
	return f.x
}

func (f *Float) BigInt() *big.Int {
	bi, _ := f.x.Int(nil)
	return bi
}

func (f *Float) Float64() float64 {
	f64, _ := f.x.Float64()
	return f64
}

func (f *Float) Sign() int {
	return f.x.Sign()
}

func (f *Float) Add(x *Float) *Float {
	return &Float{x: new(big.Float).Add(f.x, x.x)}
}

func (f *Float) Sub(x *Float) *Float {
	return &Float{x: new(big.Float).Sub(f.x, x.x)}
}

func (f *Float) Mul(x *Float) *Float {
	return &Float{x: new(big.Float).Mul(f.x, x.x)}
}

func (f *Float) Div(x *Float) *Float {
	return &Float{x: new(big.Float).Quo(f.x, x.x)}
}

// TODO: pow

func (f *Float) Sqrt() *Float {
	return &Float{x: new(big.Float).Sqrt(f.x)}
}

func (f *Float) Cmp(x *Float) int {
	return f.x.Cmp(x.x)
}

func (f *Float) Abs() *Float {
	return &Float{x: new(big.Float).Abs(f.x)}
}

func (f *Float) Neg() *Float {
	return &Float{x: new(big.Float).Neg(f.x)}
}

func (f *Float) Inv() *Float {
	return FloatFromFloat64(1).Div(f)
}

// GobEncode implements the gob.GobEncoder interface.
func (f *Float) GobEncode() ([]byte, error) {
	return f.x.GobEncode()
}

// GobDecode implements the gob.GobDecoder interface.
func (f *Float) GobDecode(b []byte) error {
	f.x = new(big.Float)
	return f.x.GobDecode(b)
}

// Value implements the driver.Valuer interface.
func (f *Float) Value() (driver.Value, error) {
	return f.GobEncode()
}

// Scan implements the sql.Scanner interface.
func (f *Float) Scan(v interface{}) error {
	switch v := v.(type) {
	case nil:
		return fmt.Errorf("nil")
	case []byte:
		return f.GobDecode(v)
	case string:
		return f.GobDecode([]byte(v))
	}
	return fmt.Errorf("unsupported type: %T", v)
}
