package dump

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerialize(t *testing.T) {
	tests := []struct {
		arg  interface{}
		want interface{}
	}{
		{arg: nil, want: nil},
		{arg: 1, want: 1},
		{arg: 1.1, want: 1.1},
		{arg: true, want: true},
		{arg: "foo", want: "foo"},
		{arg: stringer{}, want: "foo"},
		{arg: textMarshaler{}, want: "foo"},
		{arg: jsonMarshaler{`"foo"`}, want: "foo"},
		{arg: jsonMarshaler{`42`}, want: int64(42)},
		{arg: jsonMarshaler{`3.14`}, want: float64(3.14)},
		{arg: jsonMarshaler{`true`}, want: true},
		{arg: errors.New("foo"), want: "foo"},
		{arg: struct{ A int }{A: 1}, want: json.RawMessage(`{"A":1}`)},
		{arg: &struct{ A int }{A: 1}, want: json.RawMessage(`{"A":1}`)},
		{arg: []byte{0xDE, 0xAD, 0xBE, 0xEF}, want: "0xdeadbeef"},
		{arg: [4]byte{0xDE, 0xAD, 0xBE, 0xEF}, want: "0xdeadbeef"},
		{arg: [4]string{"foo", "bar", "baz", "qux"}, want: json.RawMessage(`["foo","bar","baz","qux"]`)},
		{arg: []string{"foo", "bar"}, want: json.RawMessage(`["foo","bar"]`)},
		{arg: map[string]string{"foo": "bar"}, want: json.RawMessage(`{"foo":"bar"}`)},
		{arg: emptyInterface(), want: nil},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n), func(t *testing.T) {
			assert.Equal(t, tt.want, Dump(tt.arg))
		})
	}
}

type stringer struct{}

func (stringer) String() string {
	return "foo"
}

type textMarshaler struct{}

func (textMarshaler) MarshalText() ([]byte, error) {
	return []byte("foo"), nil
}

type jsonMarshaler struct{ v string }

func (j jsonMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(j.v), nil
}

func emptyInterface() fmt.Stringer {
	var v *stringer
	return v
}
