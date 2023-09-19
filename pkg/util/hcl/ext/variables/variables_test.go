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

package variables

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"

	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
)

func TestVariables(t *testing.T) {
	tests := []struct {
		filename     string
		expectedVars map[string]cty.Value
		expectedErr  string
	}{
		{
			filename: "./testdata/variables.hcl",
			expectedVars: map[string]cty.Value{
				"foo": cty.StringVal("foo"),
			},
		},
		{
			filename:    "./testdata/circular-reference.hcl",
			expectedErr: `./testdata/circular-reference.hcl:3,9-16: Circular reference detected`, // Location and filename must be reported.
		},
		{
			filename:     "./testdata/empty.hcl",
			expectedVars: (map[string]cty.Value)(nil),
		},
		{
			filename:     "./testdata/missing.hcl",
			expectedVars: (map[string]cty.Value)(nil),
		},
		{
			filename: "./testdata/self-reference.hcl",
			expectedVars: map[string]cty.Value{
				"word_one": cty.StringVal("hello"),
				"word_two": cty.StringVal("world"),
				"greeting": cty.StringVal("hello world"),
				"map_values": cty.ObjectVal(map[string]cty.Value{
					"integer_a": cty.NumberIntVal(1),
					"integer_b": cty.NumberIntVal(2),
					"nested_map": cty.ObjectVal(map[string]cty.Value{
						"integer_d": cty.NumberIntVal(3),
						"integer_e": cty.NumberIntVal(4),
					}),
					"nested_list": cty.TupleVal([]cty.Value{
						cty.NumberIntVal(5),
						cty.NumberIntVal(6),
					}),
				}),
				"list_values": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
					cty.TupleVal([]cty.Value{
						cty.NumberIntVal(3),
						cty.NumberIntVal(4),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"key_x": cty.NumberIntVal(5),
						"key_y": cty.NumberIntVal(6),
					}),
				}),
				"simple_ref1": cty.NumberIntVal(1),
				"simple_ref2": cty.NumberIntVal(2),
				"simple_ref3": cty.NumberIntVal(1),
				"simple_ref4": cty.NumberIntVal(2),
				"simple_ref5": cty.StringVal("hello 1"),
				"simple_ref6": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				}),
				"simple_ref7": cty.ObjectVal(map[string]cty.Value{
					"ref_a": cty.NumberIntVal(1),
					"ref_b": cty.NumberIntVal(2),
				}),
				"complex_ref1": cty.ObjectVal(map[string]cty.Value{
					"nested_map_ref": cty.ObjectVal(map[string]cty.Value{
						"ref_x": cty.NumberIntVal(3),
						"ref_y": cty.NumberIntVal(4),
					}),
					"nested_list_ref": cty.TupleVal([]cty.Value{
						cty.NumberIntVal(5),
						cty.NumberIntVal(6),
					}),
				}),
				"complex_ref2": cty.ObjectVal(map[string]cty.Value{
					"nested_list_map_ref": cty.ObjectVal(map[string]cty.Value{
						"key_x": cty.NumberIntVal(5),
						"key_y": cty.NumberIntVal(6),
					}),
					"nested_list_item_ref": cty.NumberIntVal(4),
				}),
				"complex_ref3": cty.ObjectVal(map[string]cty.Value{
					"string_interpolation_ref": cty.StringVal("Value: 3, List Item: 3"),
					"map_interpolation_ref": cty.ObjectVal(map[string]cty.Value{
						"ref_a": cty.StringVal("X: 5, Y: 6"),
					}),
				}),
				"complex_ref4": cty.ObjectVal(map[string]cty.Value{
					"integer_a": cty.NumberIntVal(1),
					"integer_b": cty.NumberIntVal(1),
				}),
			},
		},
		{
			filename: "./testdata/conditional-expression.hcl",
			expectedVars: map[string]cty.Value{
				"conditional_var": cty.StringVal("true_branch"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			body, diags := utilHCL.ParseFile(tt.filename, nil)
			require.False(t, diags.HasErrors(), diags.Error())

			hclCtx := &hcl.EvalContext{}
			body, diags = Variables(hclCtx, body)
			if tt.expectedErr != "" {
				require.True(t, diags.HasErrors(), diags.Error())
				assert.Contains(t, diags.Error(), tt.expectedErr)
				return
			} else {
				require.False(t, diags.HasErrors(), diags.Error())
				vars := map[string]cty.Value{}
				if !hclCtx.Variables["var"].IsNull() {
					vars = hclCtx.Variables["var"].AsValueMap()
				}
				for k, v := range tt.expectedVars {
					assert.True(t, vars[k].RawEquals(v), "%s: expected %s to equal %s", k, vars[k].GoString(), v.GoString())
				}
				// The "variables" block should be removed from the body.
				emptySchema := &hcl.BodySchema{}
				_, diags = body.Content(emptySchema)
				require.False(t, diags.HasErrors(), diags.Error())
			}
		})
	}
}

func BenchmarkVariables(b *testing.B) {
	for _, filename := range []string{
		"./testdata/variables.hcl",
		"./testdata/empty.hcl",
		"./testdata/self-reference.hcl",
	} {
		hclCtx := &hcl.EvalContext{}
		body, diags := utilHCL.ParseFile(filename, nil)
		require.False(b, diags.HasErrors(), diags.Error())
		b.Run(filename, func(b *testing.B) {
			_, _ = Variables(hclCtx, body)
		})
	}
}
