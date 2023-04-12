package funcs

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestMakeToFunc(t *testing.T) {
	testCases := []struct {
		name    string
		wantTyp cty.Type
		input   cty.Value
		want    cty.Value
		wantErr bool
	}{
		{
			name:    "convert string to number",
			wantTyp: cty.Number,
			input:   cty.StringVal("42"),
			want:    cty.NumberIntVal(42),
			wantErr: false,
		},
		{
			name:    "convert number to string",
			wantTyp: cty.String,
			input:   cty.NumberIntVal(42),
			want:    cty.StringVal("42"),
			wantErr: false,
		},
		{
			name:    "convert bool to string",
			wantTyp: cty.String,
			input:   cty.BoolVal(true),
			want:    cty.StringVal("true"),
			wantErr: false,
		},
		{
			name:    "unsupported conversion",
			wantTyp: cty.Bool,
			input:   cty.StringVal("not-a-bool"),
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			toFunc := MakeToFunc(tt.wantTyp)
			output, err := toFunc.Call([]cty.Value{tt.input})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, output.RawEquals(tt.want), "expected output %#v, but got %#v", tt.want, output)
			}
		})
	}
}
