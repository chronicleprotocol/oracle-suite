package variables

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Variables is a custom block type that allows to define variables in the
// "variables" block. Variables are then available in "var" object.
func Variables(ctx *hcl.EvalContext, body hcl.Body) (hcl.Body, hcl.Diagnostics) {
	// Decode the "include" attribute.
	content, remain, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "variables"}},
	})
	if diags.HasErrors() {
		return nil, diags
	}

	// Evaluate the variables.
	variables := map[string]cty.Value{}
	for _, block := range content.Blocks.OfType("variables") {
		attributes, diags := block.Body.JustAttributes()
		if diags.HasErrors() {
			return nil, diags
		}
		for name, attr := range attributes {
			if variables[name], diags = attr.Expr.Value(ctx); diags.HasErrors() {
				return nil, diags
			}
		}
	}

	if ctx.Variables == nil {
		ctx.Variables = map[string]cty.Value{}
	}
	ctx.Variables["var"] = cty.ObjectVal(variables)
	return remain, nil
}
