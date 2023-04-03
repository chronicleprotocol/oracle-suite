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
			filename:    "./testdata/self-reference.hcl",
			expectedErr: `self-reference.hcl:3,9-12: Variables not allowed; Variables may not be used here.`, // Location and filename must be reported.
		},
		{
			filename:     "./testdata/empty.hcl",
			expectedVars: (map[string]cty.Value)(nil),
		},
		{
			filename:     "./testdata/missing.hcl",
			expectedVars: (map[string]cty.Value)(nil),
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
				assert.Equal(t, tt.expectedVars, hclCtx.Variables["var"].AsValueMap())

				// The "variables" block should be removed from the body.
				emptySchema := &hcl.BodySchema{}
				_, diags = body.Content(emptySchema)
				require.False(t, diags.HasErrors(), diags.Error())
			}
		})
	}
}
