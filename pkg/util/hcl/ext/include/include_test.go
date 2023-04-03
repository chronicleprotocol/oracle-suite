package include

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
)

func TestVariables(t *testing.T) {
	tests := []struct {
		filename    string
		asserts     func(t *testing.T, body hcl.Body)
		expectedErr string
	}{
		{
			filename: "./testdata/include.hcl",
			asserts: func(t *testing.T, body hcl.Body) {
				attrs, diags := body.JustAttributes()
				require.False(t, diags.HasErrors(), diags.Error())
				assert.NotNil(t, attrs["foo"])
				assert.NotNil(t, attrs["bar"])
			},
		},
		{
			filename: "./testdata/relative-dir.hcl",
			asserts: func(t *testing.T, body hcl.Body) {
				attrs, diags := body.JustAttributes()
				require.False(t, diags.HasErrors(), diags.Error())
				assert.NotNil(t, attrs["foo"])
			},
		},
		{
			filename:    "./testdata/self-include.hcl",
			expectedErr: "self-include.hcl:1,11-3,2: Too many nested includes;",
		},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			body, diags := utilHCL.ParseFile(tt.filename, nil)
			require.False(t, diags.HasErrors(), diags.Error())

			hclCtx := &hcl.EvalContext{}
			body, diags = Include(hclCtx, body, "./testdata", 2)
			if tt.expectedErr != "" {
				require.True(t, diags.HasErrors(), diags.Error())
				assert.Contains(t, diags.Error(), tt.expectedErr)
				return
			} else {
				require.False(t, diags.HasErrors(), diags.Error())
				tt.asserts(t, body)

				// The "include" attribute should be removed from the body.
				emptySchema := &hcl.BodySchema{
					Attributes: []hcl.AttributeSchema{{Name: "foo"}, {Name: "bar"}},
				}
				_, diags = body.Content(emptySchema)
				require.False(t, diags.HasErrors(), diags.Error())
			}
		})
	}
}
