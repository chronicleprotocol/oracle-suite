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
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/ptrutil"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		input   string
		target  any
		want    any
		wantErr bool
	}{
		// Basic types
		{
			input: `
				string = "foo"
				int = 1
				float = 3.14
				bool = true
				slice = [1, 2, 3]
				map = {
					"foo" = 1
					"bar" = 2
				}
				cty = "foo"
				text_unmarshaler = "foo"
				hcl_unmarshaler = "foo"
			`,
			target: &basicTypes{},
			want: &basicTypes{
				String: "foo",
				Int:    1,
				Float:  3.14,
				Bool:   true,
				Slice:  []int{1, 2, 3},
				Map: map[string]int{
					"foo": 1,
					"bar": 2,
				},
				CTY:             cty.StringVal("foo"),
				TextUnmarshaler: &textUnmarshaler{Val: "foo"},
				HCLUnmarshaler:  &hclUnmarshaler{Val: "foo"},
			},
		},
		// Blocks
		{
			input: `
				single "foo" {
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
			target: &blocks{},
			want: &blocks{
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
		},
		// Float to int
		{
			input: `
				int = 3.14
			`,
			target:  &basicTypes{},
			wantErr: true,
		},
		// Missing block label
		{
			input: `
				block {}
			`,
			target:  &singleBlock{},
			wantErr: true,
		},
		// Missing required attribute
		{
			input:   ``,
			target:  &requiredAttrs{},
			wantErr: true,
		},
		// Optional attributes (present)
		{
			input: `
				var = "foo"
				var_ptr = "foo"
			`,
			target: &optionalAttrs{},
			want: &optionalAttrs{
				Var:    "foo",
				VarPtr: ptrutil.Ptr("foo"),
			},
		},
		// Optional attributes (missing)
		{
			input:  ``,
			target: &optionalAttrs{},
			want:   &optionalAttrs{},
		},
		// Missing required block
		{
			input:   ``,
			target:  &requiredBlocks{},
			wantErr: true,
		},
		// Optional blocks (present)
		{
			input: `
				block "foo" {}
			`,
			target: &optionalBlocks{},
			want: &optionalBlocks{
				Block: &block{Label: "foo"},
			},
		},
		// Optional blocks (missing)
		{
			input:  ``,
			target: &optionalBlocks{},
			want:   &optionalBlocks{},
		},
		// Slice of blocks (present)
		{
			input: `
				slice "foo" {}
			`,
			target: &blockSlice{},
			want: &blockSlice{
				Slice: []block{{Label: "foo"}},
			},
		},
		// Slice of blocks (missing)
		{
			input:  ``,
			target: &blockSlice{},
			want:   &blockSlice{},
		},
		// Ignored field (present)
		// Ignored field must be present if they are not optional, but they
		// should not be decoded.
		{
			input: `
				var = 1
			`,
			target: &ignoredField{},
			want:   &ignoredField{},
		},
		// Ignored field (missing)
		{
			input:   ``,
			target:  &ignoredField{},
			wantErr: true,
		},
		// Any type (string)
		{
			input: `
				var = "foo"
			`,
			target: &anyField{},
			want: &anyField{
				Var: "foo",
			},
		},
		// Any type (number)
		{
			input: `
				var = 1
			`,
			target: &anyField{},
			want: &anyField{
				Var: float64(1),
			},
		},
		// Any type (bool)
		{
			input: `
				var = true
			`,
			target: &anyField{},
			want: &anyField{
				Var: true,
			},
		},
		// Any type (list)
		{
			input: `
				var = [1, 2, 3]
			`,
			target: &anyField{},
			want: &anyField{
				Var: []any{float64(1), float64(2), float64(3)},
			},
		},
		// Any type (map)
		{
			input: `
				var = {
					foo = "bar"
				}
			`,
			target: &anyField{},
			want: &anyField{
				Var: map[string]any{
					"foo": "bar",
				},
			},
		},
		// Any type (null)
		{
			input: `
				var = null
			`,
			target: &anyField{},
			want: &anyField{
				Var: nil,
			},
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
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

func TestDecodeSpecialTags(t *testing.T) {
	type config struct {
		Attr string `hcl:"attr"`

		Remain  hcl.Body        `hcl:",remain"`
		Body    hcl.Body        `hcl:",body"`
		Content hcl.BodyContent `hcl:",content"`
		Schema  hcl.BodySchema  `hcl:",schema"`
		Range   hcl.Range       `hcl:",range"`
	}
	var dest config
	file, diags := hclsyntax.ParseConfig([]byte(`attr = "foo"`), "test.hcl", hcl.Pos{})
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

func TestDecodeRecursiveSchema(t *testing.T) {
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
	require.False(t, diags.HasErrors(), diags.Error())
}

func TestDecodeEmbeddedStruct(t *testing.T) {
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
	var data = `
		block "foo" "bar" { 
			attr = "bar" 
			emb_attr = "baz" 
		}
	`
	var dest config
	file, diags := hclsyntax.ParseConfig([]byte(data), "test.hcl", hcl.Pos{})
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

func TestDecodeDuplicatedBlocks(t *testing.T) {
	type config struct {
		Block struct{} `hcl:"block,block"`
	}
	var data = `
		block {}
		block {}
	`
	var dest config
	file, diags := hclsyntax.ParseConfig([]byte(data), "test.hcl", hcl.Pos{})
	if diags.HasErrors() {
		assert.Fail(t, "parse config failed", diags)
	}
	diags = Decode(&hcl.EvalContext{}, file.Body, &dest)
	require.True(t, diags.HasErrors(), diags.Error())
}

func TestDecodeDuplicatedOptionalBlocks(t *testing.T) {
	type config struct {
		Block struct{} `hcl:"block,block,optional"`
	}
	var data = `
		block {}
		block {}
	`
	var dest config
	file, diags := hclsyntax.ParseConfig([]byte(data), "test.hcl", hcl.Pos{})
	if diags.HasErrors() {
		assert.Fail(t, "parse config failed", diags)
	}
	diags = Decode(&hcl.EvalContext{}, file.Body, &dest)
	require.True(t, diags.HasErrors(), diags.Error())
}
