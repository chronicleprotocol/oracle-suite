//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package hcl

import (
	"fmt"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/ptrutil"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func TestEncode(t *testing.T) {
	type basicTypes struct {
		String string         `hcl:"string,optional"`
		Int    int32          `hcl:"int,optional"`
		Float  float64        `hcl:"float,optional"`
		Bool   bool           `hcl:"bool,optional"`
		Slice  []int          `hcl:"slice,optional"`
		Map    map[string]int `hcl:"map,optional"`
		CTY    cty.Value      `hcl:"cty,optional"`
	}
	type block struct {
		Label string `hcl:",label"`
		Attr  string `hcl:"attr,optional"`
	}
	type blocks struct {
		Single      block              `hcl:"single,block"`
		SinglePtr   *block             `hcl:"single_ptr,block"`
		Slice       []block            `hcl:"slice,block"`
		SlicePtr    []*block           `hcl:"slice_ptr,block"`
		Map         map[string]block   `hcl:"map,block"`
		MapPtr      map[string]*block  `hcl:"map_ptr,block"`
		PtrSlice    *[]block           `hcl:"ptr_slice,block"`
		PtrSlicePtr *[]*block          `hcl:"ptr_slice_ptr,block"`
		PtrMap      *map[string]block  `hcl:"ptr_map,block"`
		PtrMapPtr   *map[string]*block `hcl:"ptr_map_ptr,block"`
	}
	type singleBlock struct {
		Block block `hcl:"block,block"`
	}
	type requiredAttrs struct {
		Var    string  `hcl:"var"`
		VarPtr *string `hcl:"var_ptr"`
	}
	type optionalAttrs struct {
		Var    string  `hcl:"var,optional"`
		VarPtr *string `hcl:"var_ptr,optional"`
	}
	type requiredBlocks struct {
		Block    block  `hcl:"block,block"`
		BlockPtr *block `hcl:"block_ptr,block"`
	}
	type optionalBlocks struct {
		Block    *block `hcl:"block,block,optional"`
		BlockPtr *block `hcl:"block_ptr,block,optional"`
	}
	type blockSlice struct {
		Slice []block `hcl:"slice,block"`
	}
	type ignoredField struct {
		Var string `hcl:"var,ignore"`
	}
	type anyField struct {
		Var any `hcl:"var"`
	}

	tests := []struct {
		input         any
		target        string
		expectedError string
	}{
		// Basic Types
		{
			input: &basicTypes{
				String: "foo",
				Int:    1,
				Float:  3.14,
				Bool:   true,
				Slice:  []int{1, 2, 3},
				Map: map[string]int{
					"foo": 1,
					"bar": 2,
				},
				CTY: cty.StringVal("foo"),
			},
			target: `string = "foo"
int    = 1
float  = 3.14
bool   = true
slice  = [1, 2, 3]
map = {
  bar = 2
  foo = 1
}
cty = "foo"
`,
			expectedError: "",
		},
		// Blocks
		{
			input: &blocks{
				Single: block{
					Label: "foo",
					Attr:  "foo",
				},
				SinglePtr: &block{
					Label: "foo",
					Attr:  "foo",
				},
				Slice: []block{
					{
						Label: "foo",
						Attr:  "foo",
					},
					{
						Label: "bar",
						Attr:  "bar",
					},
				},
				SlicePtr: []*block{
					{
						Label: "foo",
						Attr:  "foo",
					},
					{
						Label: "bar",
						Attr:  "bar",
					},
				},
				Map: map[string]block{
					"foo": {
						Label: "foo",
						Attr:  "foo",
					},
					"bar": {
						Label: "bar",
						Attr:  "bar",
					},
				},
				MapPtr: map[string]*block{
					"foo": {
						Label: "foo",
						Attr:  "foo",
					},
					"bar": {
						Label: "bar",
						Attr:  "bar",
					},
				},
				PtrSlice: &[]block{
					{
						Label: "foo",
						Attr:  "foo",
					},
					{
						Label: "bar",
						Attr:  "bar",
					},
				},
				PtrSlicePtr: &[]*block{
					{
						Label: "foo",
						Attr:  "foo",
					},
					{
						Label: "bar",
						Attr:  "bar",
					},
				},
				PtrMap: &map[string]block{
					"foo": {
						Label: "foo",
						Attr:  "foo",
					},
					"bar": {
						Label: "bar",
						Attr:  "bar",
					},
				},
				PtrMapPtr: &map[string]*block{
					"foo": {
						Label: "foo",
						Attr:  "foo",
					},
					"bar": {
						Label: "bar",
						Attr:  "bar",
					},
				},
			},
			target: `single "foo" {
  attr = "foo"
}
single_ptr "foo" {
  attr = "foo"
}
slice "foo" {
  attr = "foo"
}
slice "bar" {
  attr = "bar"
}
slice_ptr "foo" {
  attr = "foo"
}
slice_ptr "bar" {
  attr = "bar"
}
map "foo" {
  attr = "foo"
}
map "bar" {
  attr = "bar"
}
map_ptr "foo" {
  attr = "foo"
}
map_ptr "bar" {
  attr = "bar"
}
ptr_slice "foo" {
  attr = "foo"
}
ptr_slice "bar" {
  attr = "bar"
}
ptr_slice_ptr "foo" {
  attr = "foo"
}
ptr_slice_ptr "bar" {
  attr = "bar"
}
ptr_map "foo" {
  attr = "foo"
}
ptr_map "bar" {
  attr = "bar"
}
ptr_map_ptr "foo" {
  attr = "foo"
}
ptr_map_ptr "bar" {
  attr = "bar"
}
`,
			expectedError: "",
		},
		// Missing block label
		{
			input:         &singleBlock{},
			target:        "",
			expectedError: "missing block label: block",
		},
		// Missing required attribute
		{
			input:         &requiredAttrs{},
			target:        ``,
			expectedError: "missing attribute: var",
		},
		// Optional attributes (present)
		{
			input: &optionalAttrs{
				Var:    "foo",
				VarPtr: ptrutil.Ptr("foo"),
			},
			target: `var     = "foo"
var_ptr = "foo"
`,
			expectedError: "",
		},
		// Optional attributes (missing)
		{
			input:         &optionalAttrs{},
			target:        ``,
			expectedError: "",
		},
		// Missing required block
		{
			input: &requiredBlocks{
				Block: block{Label: "foo"},
			},
			target: `block "foo" {
}
`,
			expectedError: "missing block: block_ptr",
		},
		// Optional blocks (missing)
		{
			input:         &optionalBlocks{},
			target:        ``,
			expectedError: "",
		},
		// Optional blocks (present)
		{
			input: &optionalBlocks{
				Block: &block{Label: "foo"},
			},
			target: `block "foo" {
}
`,
			expectedError: "",
		},
		// Slice of blocks (present)
		{
			input: &blockSlice{
				Slice: []block{{Label: "foo"}},
			},
			target: `slice "foo" {
}
`,
			expectedError: "",
		},
		// Slice of blocks (missing)
		{
			input:         &blockSlice{},
			target:        ``,
			expectedError: "",
		},
		// Ignored field (present)
		// Ignored field must be present if they are not optional, but they
		// should not be decoded.
		{
			input:         &ignoredField{Var: "1"},
			target:        ``,
			expectedError: "",
		},
		// Ignored field (missing)
		{
			input:         &ignoredField{},
			target:        ``,
			expectedError: "",
		},
		// Any type (string)
		{
			input: &anyField{Var: "foo"},
			target: `var = "foo"
`,
			expectedError: "",
		},
		// Any type (number)
		{
			input: &anyField{Var: float64(1)},
			target: `var = 1
`,
			expectedError: "",
		},
		// Any type (bool)
		{
			input: &anyField{Var: true},
			target: `var = true
`,
			expectedError: "",
		},
		// Any type (list)
		{
			input: &anyField{Var: []any{float64(1), float64(2), float64(3)}},
			target: `var = [1, 2, 3]
`,
			expectedError: "",
		},
		// Any type (map)
		{
			input: &anyField{Var: map[string]string{
				"foo": "bar",
			}},
			target: `var = {
  foo = "bar"
}
`,
			expectedError: "",
		},
		// Any type (null)
		{
			input:         &anyField{Var: nil},
			target:        ``,
			expectedError: "missing attribute: var",
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			f := hclwrite.NewFile()
			body := f.Body()
			err := Encode(tt.input, body)
			if tt.expectedError == "" {
				assert.Nil(t, err)
			} else {
				require.NotNil(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			}
			assert.Equal(t, tt.target, string(f.Bytes()))
		})
	}
}

func TestEncodeSpecialTags(t *testing.T) {
	type config struct {
		Attr string `hcl:"attr"`

		Remain  hcl.Body        `hcl:",remain"`
		Body    hcl.Body        `hcl:",body"`
		Content hcl.BodyContent `hcl:",content"`
		Schema  hcl.BodySchema  `hcl:",schema"`
		Range   hcl.Range       `hcl:",range"`
	}
	var configVar = config{
		Attr: "foo",
	}
	var dest config
	// Encode `dest`
	f := hclwrite.NewFile()
	body := f.Body()
	err := Encode(configVar, body)
	assert.Nil(t, err)
	// Decode encodes string
	file, diags := hclsyntax.ParseConfig(f.Bytes(), "test.hcl", hcl.Pos{})
	if diags.HasErrors() {
		assert.Fail(t, "parse config failed", diags)
	}
	diags = Decode(&hcl.EvalContext{}, file.Body, &dest)
	require.False(t, diags.HasErrors(), diags.Error())
	assert.NotNil(t, dest.Remain)
	assert.NotNil(t, dest.Body)
	assert.Len(t, dest.Content.Attributes, 1)
	assert.Len(t, dest.Schema.Attributes, 1)
	assert.Equal(t, ":0,0-0", dest.Range.String())
}

func TestEncodeEmbeddedStruct(t *testing.T) {
	type embedded struct {
		EmbLabel string `hcl:",label"`
		EmbAttr  string `hcl:"emb_attr"`
	}
	type block struct {
		Label string `hcl:",label"`
		Attr  string `hcl:"attr"`
		embedded
	}
	type config struct {
		Block block `hcl:"block,block"`
	}

	var embeddedVar = embedded{
		EmbLabel: "bar",
		EmbAttr:  "baz",
	}
	var blockVar = block{
		Label:    "foo",
		Attr:     "bar",
		embedded: embeddedVar,
	}
	var configVar = config{
		Block: blockVar,
	}
	var dest config
	//var data = `
	//	block "foo" "bar" {
	//		attr = "bar"
	//		emb_attr = "baz"
	//	}
	//`

	// Encode `dest`
	f := hclwrite.NewFile()
	body := f.Body()
	err := Encode(configVar, body)
	assert.Nil(t, err)
	// Decode encodes string
	file, diags := hclsyntax.ParseConfig(f.Bytes(), "test.hcl", hcl.Pos{})
	if diags.HasErrors() {
		assert.Fail(t, "parse config failed", diags)
	}
	diags = Decode(&hcl.EvalContext{}, file.Body, &dest)
	require.False(t, diags.HasErrors(), diags.Error())
	assert.Equal(t, "foo", dest.Block.Label)
	assert.Equal(t, "bar", dest.Block.EmbLabel)
	assert.Equal(t, "bar", dest.Block.Attr)
	assert.Equal(t, "baz", dest.Block.EmbAttr)
}
