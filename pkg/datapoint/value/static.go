package value

import (
	"encoding/json"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// StaticValue is a numeric value obtained from a static origin.
type StaticValue struct {
	Value *bn.DecFloatPointNumber
}

// Number implements the NumericValue interface.
func (s StaticValue) Number() *bn.FloatNumber {
	if s.Value == nil {
		return nil
	}
	return s.Value.Float()
}

// Print implements the Value interface.
func (s StaticValue) Print() string {
	if s.Value == nil {
		return "<nil>"
	}
	return s.Value.Text('g', 10)
}

func (s StaticValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value.String())
}

func (s *StaticValue) UnmarshalJSON(bytes []byte) error {
	var str string
	if err := json.Unmarshal(bytes, &str); err != nil {
		return err
	}
	s.Value = bn.DecFloatPoint(str)
	return nil
}
