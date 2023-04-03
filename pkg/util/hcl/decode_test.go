package hcl

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/ptrutil"
)

type textUnmarshaler struct {
	Val string
}

func (t *textUnmarshaler) UnmarshalText(text []byte) error {
	t.Val = string(text)
	return nil
}

func TestDecode(t *testing.T) {
	type basicTypes struct {
		Var1 string         `hcl:"var1"`
		Var2 int            `hcl:"var2"`
		Var3 bool           `hcl:"var3"`
		Var4 []int          `hcl:"var4"`
		Var5 map[string]int `hcl:"var5"`
	}
	type block struct {
		Label string `hcl:"label,label"`
		Var1  string `hcl:"var1,optional"`
	}
	type blocks struct {
		Single   block             `hcl:"single,block"`
		Optional *block            `hcl:"optional,block"`
		Slice    []block           `hcl:"slice,block"`
		SlicePtr []*block          `hcl:"slice_ptr,block"`
		Map      map[string]block  `hcl:"map,block"`
		MapPtr   map[string]*block `hcl:"map_ptr,block"`
	}
	type optionalAttr struct {
		Var1 string  `hcl:"var1,optional"`
		Var2 *string `hcl:"var2,optional"`
	}
	type mapToCty struct {
		Var1 cty.Value `hcl:"var1"`
	}
	type textUnmarshalerLabel struct {
		Label textUnmarshaler `hcl:"label,label"`
	}
	type textUnmarshalerLabelBlock struct {
		Block textUnmarshalerLabel `hcl:"block,block"`
	}
	type requiredBlock struct {
		Block block `hcl:"block,block"`
	}
	type optionalBlock struct {
		Block *block `hcl:"block,block"`
	}
	type ignoredAttr struct {
		Ignored string `hcl:"ignored,ignore"`
	}
	tests := []struct {
		input   string
		target  any
		want    any
		wantErr bool
	}{
		// Basic types
		{
			input: `
				var1 = "foo"
				var2 = 1
				var3 = true
				var4 = [1, 2, 3]
				var5 = {
					"foo" = 1
					"bar" = 2
				}
			`,
			target: &basicTypes{},
			want: &basicTypes{
				Var1: "foo",
				Var2: 1,
				Var3: true,
				Var4: []int{1, 2, 3},
				Var5: map[string]int{
					"foo": 1,
					"bar": 2,
				},
			},
		},
		// Blocks
		{
			input: `
				single "foo" {
					var1 = "foo"
				}
				optional "bar" {
					var1 = "bar"
				}
				slice "foo" {
					var1 = "foo"
				}
				slice "bar" {
					var1 = "bar"
				}
				slice_ptr "foo" {
					var1 = "foo"
				}
				slice_ptr "bar" {
					var1 = "bar"
				}
				map "foo" {
					var1 = "foo"
				}
				map_ptr "foo" {
					var1 = "foo"
				}
			`,
			target: &blocks{},
			want: &blocks{
				Single: block{
					Label: "foo",
					Var1:  "foo",
				},
				Optional: &block{
					Label: "bar",
					Var1:  "bar",
				},
				Slice: []block{
					{
						Label: "foo",
						Var1:  "foo",
					},
					{
						Label: "bar",
						Var1:  "bar",
					},
				},
				SlicePtr: []*block{
					{
						Label: "foo",
						Var1:  "foo",
					},
					{
						Label: "bar",
						Var1:  "bar",
					},
				},
				Map: map[string]block{
					"foo": {
						Label: "foo",
						Var1:  "foo",
					},
				},
				MapPtr: map[string]*block{
					"foo": {
						Label: "foo",
						Var1:  "foo",
					},
				},
			},
		},
		// Optional attr (present)
		{
			input: `
				var1 = "foo"
				var2 = "bar"
			`,
			target: &optionalAttr{},
			want: &optionalAttr{
				Var1: "foo",
				Var2: ptrutil.Ptr("bar"),
			},
		},
		// Optional attr (absent)
		{
			input:  ``,
			target: &optionalAttr{},
			want: &optionalAttr{
				Var1: "",
				Var2: nil,
			},
		},
		// Map to cty.Value
		{
			input: `
				var1 = "foo"
			`,
			target: &mapToCty{},
			want: &mapToCty{
				Var1: cty.StringVal("foo"),
			},
		},
		// Label that implements TextUnmarshaler
		{
			input: `
				block "foo" { }
			`,
			target: &textUnmarshalerLabelBlock{},
			want: &textUnmarshalerLabelBlock{
				Block: textUnmarshalerLabel{Label: textUnmarshaler{Val: "foo"}},
			},
		},
		// String to int
		{
			input: `
				var1 = "1"
			`,
			target: &struct {
				Var1 int `hcl:"var1"`
			}{},
			wantErr: true,
		},
		// Int to string
		{
			input: `
				var1 = 1
			`,
			target: &struct {
				Var1 string `hcl:"var1"`
			}{},
			wantErr: true,
		},
		// Required block
		{
			input:   ``,
			target:  &requiredBlock{},
			wantErr: true,
		},
		// Optional block
		{
			input:  ``,
			target: &optionalBlock{},
			want: &optionalBlock{
				Block: nil,
			},
		},
		// Extraneous block
		{
			input: `
				block "foo" { }
			    block "bar" { }
			`,
			target:  &requiredBlock{},
			wantErr: true,
		},
		// Ignored attr
		{
			input: `
				ignored = "foo"
			`,
			target: &ignoredAttr{},
			want:   &ignoredAttr{},
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n), func(t *testing.T) {
			file, diags := hclsyntax.ParseConfig([]byte(tt.input), "test.hcl", hcl.Pos{})
			if diags.HasErrors() {
				assert.Fail(t, "parse config failed", diags)
			}
			diags = Decode(&hcl.EvalContext{}, file.Body, tt.target)
			if tt.wantErr {
				assert.True(t, diags.HasErrors())
				return
			}
			if diags.HasErrors() {
				assert.Fail(t, "decode failed", diags)
			}
			assert.Equal(t, tt.want, tt.target)
		})
	}
}

func TestSpecialTags(t *testing.T) {
	type config struct {
		Remain  hcl.Body        `hcl:",remain"`
		Body    hcl.Body        `hcl:",body"`
		Content hcl.BodyContent `hcl:",content"`
		Schema  hcl.BodySchema  `hcl:",schema"`
		Range   hcl.Range       `hcl:",range"`
	}
	var dest config
	file, diags := hclsyntax.ParseConfig([]byte(``), "test.hcl", hcl.Pos{})
	if diags.HasErrors() {
		assert.Fail(t, "parse config failed", diags)
	}
	diags = Decode(&hcl.EvalContext{}, file.Body, &dest)
	require.False(t, diags.HasErrors(), diags.Error())
}

func TestRecursiveSchema(t *testing.T) {
	type recur struct {
		Recur []recur `hcl:"Recur,block"`
	}
	type config struct {
		Recur recur `hcl:"Recur,block"`
	}
	var data = `Recur {}`
	var dest config
	file, diags := hclsyntax.ParseConfig([]byte(data), "test.hcl", hcl.Pos{})
	if diags.HasErrors() {
		assert.Fail(t, "parse config failed", diags)
	}
	diags = Decode(&hcl.EvalContext{}, file.Body, &dest)
	require.False(t, diags.HasErrors())
}
