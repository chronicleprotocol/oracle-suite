package value

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// Mapping from types to their unique IDs.
// The 0 ID is reserved for unknown types.
// Values larger than 0x80000000 are reserved for testing.
// Changing these IDs will break backwards compatibility.
//
//nolint:gomnd
var registeredTypes = map[reflect.Type]uint32{
	reflect.TypeOf((*StaticValue)(nil)): 0x00000001,
	reflect.TypeOf((*Tick)(nil)):        0x00000002,
}

// Value is a data point value.
//
// A value can be anything, e.g. a number, a string, a struct, etc.
//
// To be able to send values over the network, they must implement the
// MarshalableValue and UnmarshalableValue interfaces.
//
// The interface must be implemented by using non-pointer receivers.
type Value interface {
	// Print returns a human-readable representation of the value.
	Print() string
}

// MarshalableValue is a data point value which can be serialized to binary.
//
// The interface must be implemented by using non-pointer receivers.
type MarshalableValue interface {
	encoding.BinaryMarshaler
	Value
}

// UnmarshalableValue is a data point value which can be deserialized from binary.
//
// The UnmarshalBinary must be implemented by using pointer receivers.
type UnmarshalableValue interface {
	encoding.BinaryUnmarshaler
	Value
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

// RegisterType registers a new value type.
//
// The type ID must be unique and not equal to 0.
func RegisterType(value Value, id uint32) {
	rt := reflect.TypeOf(value)
	if rt.Kind() != reflect.Ptr {
		rt = reflect.PtrTo(rt)
	}
	registeredTypes[rt] = id
}

// MarshalBinary serializes a Value to binary.
//
// The output includes the type ID followed by the binary representation of the
// value.
func MarshalBinary(value Value) ([]byte, error) {
	rt := reflect.TypeOf(value)
	if rt.Kind() != reflect.Ptr {
		rt = reflect.PtrTo(rt)
	}
	if rID, ok := registeredTypes[rt]; ok {
		value, ok := value.(MarshalableValue)
		if !ok {
			return nil, fmt.Errorf("value type %s does not implement MarshalableValue", rt)
		}
		data, err := value.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(uint32ToBytes(rID), data...), nil
	}
	return nil, fmt.Errorf("unknown value type: %s", rt)
}

// UnmarshalBinary deserializes a Value from binary.
//
// The input is expected to start with the type ID, followed by the binary
// representation of the value.
func UnmarshalBinary(data []byte) (Value, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid value data")
	}
	tID := bytesToUint32(data[:4])
	for rt, rID := range registeredTypes {
		if tID == rID {
			value, ok := reflect.New(rt.Elem()).Interface().(UnmarshalableValue)
			if !ok {
				return nil, fmt.Errorf("value type %s does not implement UnmarshalableValue", rt)
			}
			if err := value.UnmarshalBinary(data[4:]); err != nil {
				return nil, err
			}
			return reflect.ValueOf(value).Elem().Interface().(Value), nil
		}
	}
	return nil, fmt.Errorf("unknown value type: %d", tID)
}

// uint32ToBytes takes an uint32 and returns its byte representation.
func uint32ToBytes(value uint32) []byte {
	return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)} //nolint:gomnd
}

// bytesToUint32 takes a byte slice and returns its uint32 representation.
func bytesToUint32(bytes []byte) uint32 {
	return uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
}
