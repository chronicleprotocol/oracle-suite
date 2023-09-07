package value

import (
	"encoding/json"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// StaticNumberPrecision is a precision of static numbers.
const StaticNumberPrecision = 18

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

func (s StaticValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value.String())
}

func (s *StaticValue) UnmarshalJSON(bytes []byte) error {
	var str string
	if err := json.Unmarshal(bytes, &str); err != nil {
		return err
	}
	s.Value = bn.Float(str)
	return nil
}
