package grafana

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_byPath(t *testing.T) {
	tests := []struct {
		value   interface{}
		path    string
		want    interface{}
		invalid bool
	}{
		{
			value: "test",
			path:  "",
			want:  "test",
		},
		{
			value:   "test",
			path:    "abc",
			invalid: true,
		},
		{
			value: struct {
				Field string
			}{
				Field: "test",
			},
			path: "Field",
			want: "test",
		},
		{
			value: map[string]string{"Key": "test"},
			path:  "Key",
			want:  "test",
		},
		{
			value:   map[int]string{42: "test"},
			path:    "42",
			invalid: true,
		},
		{
			value: []string{"test"},
			path:  "0",
			want:  "test",
		},
		{
			value: struct {
				Field map[string][]int
			}{
				Field: map[string][]int{"Field2": {42}},
			},
			path: "Field.Field2.0",
			want: 42,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := byPath(reflect.ValueOf(tt.value), tt.path)
			if tt.invalid {
				assert.False(t, v.IsValid())
			} else {
				assert.Equal(t, tt.want, v.Interface())
			}
		})
	}
}
