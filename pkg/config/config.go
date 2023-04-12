package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"

	"github.com/hashicorp/hcl/v2/ext/dynblock"

	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/hcl/ext/include"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/hcl/ext/variables"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/hcl/funcs"
)

var hclContext = &hcl.EvalContext{
	Variables: map[string]cty.Value{
		"env": getEnvVars(),
	},
	Functions: map[string]function.Function{
		// Standard library functions:
		"try":    tryfunc.TryFunc,
		"can":    tryfunc.CanFunc,
		"range":  stdlib.RangeFunc,
		"length": stdlib.LengthFunc,
		"split":  stdlib.SplitFunc,

		// Custom functions:
		"tostring": funcs.MakeToFunc(cty.String),
		"tonumber": funcs.MakeToFunc(cty.Number),
		"tobool":   funcs.MakeToFunc(cty.Bool),
		"toset":    funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tolist":   funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":    funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
	},
}

// LoadFiles loads the given paths into the given config, merging contents of
// multiple HCL files specified by the "include" attribute using glob patterns,
// and expanding dynamic blocks before decoding the HCL content.
func LoadFiles(config any, paths []string) error {
	var body hcl.Body
	var diags hcl.Diagnostics
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if body, diags = utilHCL.ParseFiles(paths, nil); diags.HasErrors() {
		return diags
	}
	if body, diags = include.Include(hclContext, body, wd, 10); diags.HasErrors() {
		return diags
	}
	if body, diags = variables.Variables(hclContext, body); diags.HasErrors() {
		return diags
	}
	if diags = utilHCL.Decode(hclContext, dynblock.Expand(body, hclContext), config); diags.HasErrors() {
		return diags
	}
	return nil
}

// getEnvVars retrieves environment variables from the system and returns
// them as a cty object type, where keys are variable names and values are
// their corresponding values.
func getEnvVars() cty.Value {
	envVars := make(map[string]cty.Value)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		envVars[parts[0]] = cty.StringVal(parts[1])
	}
	return cty.ObjectVal(envVars)
}
