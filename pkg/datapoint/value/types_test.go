package value

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func RemoveTestTypes() {
	for k, v := range registeredTypes {
		if v >= uint32(0x80000000) {
			delete(registeredTypes, k)
		}
	}
}

type mockValue struct {
	bin []byte
}

func (m mockValue) Print() string {
	return "MockValue"
}

func (m mockValue) MarshalBinary() ([]byte, error) {
	return m.bin, nil
}

func (m *mockValue) UnmarshalBinary(data []byte) error {
	m.bin = data
	return nil
}

func TestMain(m *testing.M) {
	RemoveTestTypes()
	RegisterType(&mockValue{}, 0x80000003)
	m.Run()
}

func TestMarshalBinary(t *testing.T) {
	m := &mockValue{}
	_, err := MarshalBinary(m)
	assert.NoError(t, err)
}

func TestUnmarshalBinary(t *testing.T) {
	data := []byte{128, 0, 0, 3, 1, 2, 3, 4}
	m, err := UnmarshalBinary(data)
	assert.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3, 4}, m.(mockValue).bin)
}

func TestUnmarshalBinary_InvalidData(t *testing.T) {
	data := []byte{128, 0, 0} // Less than 4 bytes
	_, err := UnmarshalBinary(data)
	assert.Error(t, err)
}

func TestUnmarshalBinary_UnknownType(t *testing.T) {
	data := []byte{128, 0, 0, 5, 1, 2, 3, 4} // Type ID 5 is not registered
	_, err := UnmarshalBinary(data)
	assert.Error(t, err)
}
