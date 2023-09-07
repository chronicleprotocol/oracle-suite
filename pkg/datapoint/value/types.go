package value

import (
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// Value is a data point value.
//
// A value can be anything, e.g. a number, a string, a struct, etc.
//
// The interface must be implemented by using non-pointer receivers.
type Value interface {
	// Print returns a human-readable representation of the value.
	Print() string
}

// ValidatableValue is a data point value which can be validated.
//
// The interface must be implemented by using non-pointer receivers.
type ValidatableValue interface {
	Validate() error
}

// NumericValue is a data point value which is a number.
//
// The interface must be implemented by using non-pointer receivers.
type NumericValue interface {
	Number() *bn.FloatNumber
}
