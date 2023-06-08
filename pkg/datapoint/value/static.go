package value

import (
	"math/big"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// StaticValue is a numeric value obtained from a static origin.
type StaticValue struct {
	Value *bn.FloatNumber
}

// Number implements the NumericValue interface.
func (s StaticValue) Number() *bn.FloatNumber {
	return s.Value
}

// Print implements the Value interface.
func (s StaticValue) Print() string {
	return s.Value.String()
}

// MarshalBinary implements the Value interface.
func (s StaticValue) MarshalBinary() ([]byte, error) {
	return s.Value.Mul(RealNumberPrecision).BigInt().Bytes(), nil
}

// UnmarshalBinary implements the Value interface.
func (s *StaticValue) UnmarshalBinary(bytes []byte) error {
	s.Value = bn.Float(new(big.Int).SetBytes(bytes)).Div(RealNumberPrecision)
	return nil
}
