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
		{arg: 1, want: 1},
		{arg: 1.1, want: 1.1},
		{arg: "foo", want: "foo"},
		{arg: []byte{0xDE, 0xAD, 0xBE, 0xEF}, want: "0xdeadbeef"},
		{arg: []string{"foo", "bar"}, want: json.RawMessage(`["foo","bar"]`)},
		{arg: map[string]string{"foo": "bar"}, want: json.RawMessage(`{"foo":"bar"}`)},
		{arg: struct{ A int }{A: 1}, want: json.RawMessage(`{"A":1}`)},
		{arg: &struct{ A int }{A: 1}, want: json.RawMessage(`{"A":1}`)},
		{arg: errors.New("foo"), want: "foo"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n), func(t *testing.T) {
			assert.Equal(t, tt.want, Dump(tt.arg))
		})
	}
}
