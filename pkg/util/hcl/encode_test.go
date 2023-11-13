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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/ptrutil"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		input  any
		target string
	}{
		// Basic Types
		{
			input: &basicTypes{
				String: "foo",
				Int:    1,
				Float:  1.25,
				Bool:   true,
				Slice:  []int{1, 2, 3},
				Map: map[string]int{
					"foo": 1,
					"bar": 2,
				},
				CTY: cty.StringVal("foo"),
			},
			target: `
string = "foo"
int    = 1
float  = 1.25
bool   = true
slice  = [1, 2, 3]
map = {
  bar = 2
  foo = 1
}
cty = "foo"
`,
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
						Label: "bar",
						Attr:  "bar",
					},
					{
						Label: "foo",
						Attr:  "foo",
					},
				},
				SlicePtr: []*block{
					{
						Label: "bar",
						Attr:  "bar",
					},
					{
						Label: "foo",
						Attr:  "foo",
					},
				},
				Map: map[string]block{
					"bar": {
						Label: "bar",
						Attr:  "bar",
					},
					"foo": {
						Label: "foo",
						Attr:  "foo",
					},
				},
				MapPtr: map[string]*block{
					"bar": {
						Label: "bar",
						Attr:  "bar",
					},
					"foo": {
						Label: "foo",
						Attr:  "foo",
					},
				},
				PtrSlice: &[]block{
					{
						Label: "bar",
						Attr:  "bar",
					},
					{
						Label: "foo",
						Attr:  "foo",
					},
				},
				PtrSlicePtr: &[]*block{
					{
						Label: "bar",
						Attr:  "bar",
					},
					{
						Label: "foo",
						Attr:  "foo",
					},
				},
				PtrMap: &map[string]block{
					"bar": {
						Label: "bar",
						Attr:  "bar",
					},
					"foo": {
						Label: "foo",
						Attr:  "foo",
					},
				},
				PtrMapPtr: &map[string]*block{
					"bar": {
						Label: "bar",
						Attr:  "bar",
					},
					"foo": {
						Label: "foo",
						Attr:  "foo",
					},
				},
			},
			target: `
single "foo" { attr = "foo" }
single_ptr "foo" { attr = "foo" }
slice "bar" { attr = "bar" }
slice "foo" { attr = "foo" }
slice_ptr "bar" { attr = "bar" }
slice_ptr "foo" { attr = "foo" }
map "bar" { attr = "bar" }
map "foo" { attr = "foo" }
map_ptr "bar" { attr = "bar" }
map_ptr "foo" { attr = "foo" }
ptr_slice "bar" { attr = "bar" }
ptr_slice "foo" { attr = "foo" }
ptr_slice_ptr "bar" { attr = "bar" }
ptr_slice_ptr "foo" { attr = "foo" }
ptr_map "bar" { attr = "bar" }
ptr_map "foo" { attr = "foo" }
ptr_map_ptr "bar" { attr = "bar" }
ptr_map_ptr "foo" { attr = "foo" }
`,
		},
		// Optional attributes
		{
			input: &optionalAttrs{
				Var:    "foo",
				VarPtr: ptrutil.Ptr("foo"),
			},
			target: `
var     = "foo"
var_ptr = "foo"
`,
		},
		// Optional blocks
		{
			input: &optionalBlocks{
				Block: &block{Label: "foo"},
			},
			target: `
block "foo" {}
`,
		},
		// Slice of blocks (present)
		{
			input: &blockSlice{
				Slice: []block{{Label: "foo"}},
			},
			target: `
slice "foo" {}
`,
		},
		// Any type (string)
		{
			input:  &anyField{Var: "foo"},
			target: `var = "foo"`,
		},
		// Any type (number)
		{
			input:  &anyField{Var: float64(1)},
			target: `var = 1`,
		},
		// Any type (bool)
		{
			input:  &anyField{Var: true},
			target: `var = true`,
		},
		// Any type (list)
		{
			input:  &anyField{Var: []any{float64(1), float64(2), float64(3)}},
			target: `var = [1, 2, 3]`,
		},
		// Any type (map)
		{
			input: &anyField{Var: map[string]string{
				"foo": "bar",
			}},
			target: `
var = {
  foo = "bar"
}
`,
		},
		// Empty block
		{
			input:  &singleEmptyBlock{Block: emptyBlock{Label: "foo"}},
			target: `block "foo" {}`,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			body := &Block{}
			require.False(t, Encode(tt.input, body).HasErrors())
			hcl, err := body.Bytes()
			require.False(t, err.HasErrors())
			assert.Equal(t, strings.TrimSpace(tt.target), strings.TrimSpace(string(hcl)))
		})
	}
}
