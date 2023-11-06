package dump

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDump(t *testing.T) {
	tests := []struct {
		arg  interface{}
		want interface{}
	}{
		{arg: nil, want: nil},
		{arg: 1, want: 1},
		{arg: 1.1, want: 1.1},
		{arg: true, want: true},
		{arg: "foo", want: "foo"},
		{arg: json.RawMessage(`"foo"`), want: "foo"},
		{arg: stringer{}, want: "foo"},
		{arg: textMarshaler{v: "foo"}, want: "foo"},
		{arg: textMarshaler{e: errors.New("foo")}, want: "<error: foo>"},
		{arg: jsonMarshaler{v: `"foo"`}, want: "foo"},
		{arg: jsonMarshaler{v: `42`}, want: int64(42)},
		{arg: jsonMarshaler{v: `3.14`}, want: float64(3.14)},
		{arg: jsonMarshaler{v: `true`}, want: true},
		{arg: jsonMarshaler{e: errors.New("foo")}, want: "<error: json: error calling MarshalJSON for type dump.jsonMarshaler: foo>"},
		{arg: errors.New("foo"), want: "foo"},
		{arg: struct{ A int }{A: 1}, want: json.RawMessage(`{"A":1}`)},
		{arg: &struct{ A int }{A: 1}, want: json.RawMessage(`{"A":1}`)},
		{arg: []byte{0xDE, 0xAD, 0xBE, 0xEF}, want: "0xdeadbeef"},
		{arg: [4]byte{0xDE, 0xAD, 0xBE, 0xEF}, want: "0xdeadbeef"},
		{arg: [4]string{"foo", "bar", "baz", "qux"}, want: json.RawMessage(`["foo","bar","baz","qux"]`)},
		{arg: []string{"foo", "bar"}, want: json.RawMessage(`["foo","bar"]`)},
		{arg: map[string]string{"foo": "bar"}, want: json.RawMessage(`{"foo":"bar"}`)},
		{arg: emptyInterface(), want: nil},
		{arg: panicer{}, want: "<panic: foo>"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n), func(t *testing.T) {
			assert.Equal(t, tt.want, Dump(tt.arg))
		})
	}
}

func TestMaxDepth(t *testing.T) {
	r := recursive{}
	r.A = &r
	assert.Contains(t, fmt.Sprintf("%s", Dump(r)), "max depth reached")
}

type stringer struct{}

func (stringer) String() string {
	return "foo"
}

type textMarshaler struct {
	v string
	e error
}

func (t textMarshaler) MarshalText() ([]byte, error) {
	return []byte(t.v), t.e
}

type jsonMarshaler struct {
	v string
	e error
}

func (j jsonMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(j.v), j.e
}

type recursive struct {
	A *recursive
}

type panicer struct{}

func (panicer) String() string {
	panic("foo")
}

func emptyInterface() fmt.Stringer {
	var v *stringer
	return v
}
