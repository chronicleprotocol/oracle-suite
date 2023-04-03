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
)

var hclContext = &hcl.EvalContext{
	Variables: map[string]cty.Value{
		"env": getEnvVars(),
	},
	Functions: map[string]function.Function{
		// HCL extension functions:
		"try": tryfunc.TryFunc,
		"can": tryfunc.CanFunc,

		// Stdlib functions taken from:
		// https://github.com/hashicorp/terraform/blob/4fd832280200f57a747ea3f8c5a10f17c6e69ccc/internal/lang/functions.go
		// TODO(mdobak): Not sure if we need all of these.
		"abs":             stdlib.AbsoluteFunc,
		"ceil":            stdlib.CeilFunc,
		"chomp":           stdlib.ChompFunc,
		"coalescelist":    stdlib.CoalesceListFunc,
		"compact":         stdlib.CompactFunc,
		"concat":          stdlib.ConcatFunc,
		"contains":        stdlib.ContainsFunc,
		"csvdecode":       stdlib.CSVDecodeFunc,
		"distinct":        stdlib.DistinctFunc,
		"element":         stdlib.ElementFunc,
		"chunklist":       stdlib.ChunklistFunc,
		"flatten":         stdlib.FlattenFunc,
		"floor":           stdlib.FloorFunc,
		"format":          stdlib.FormatFunc,
		"formatdate":      stdlib.FormatDateFunc,
		"formatlist":      stdlib.FormatListFunc,
		"indent":          stdlib.IndentFunc,
		"index":           stdlib.IndexFunc,
		"join":            stdlib.JoinFunc,
		"jsondecode":      stdlib.JSONDecodeFunc,
		"jsonencode":      stdlib.JSONEncodeFunc,
		"keys":            stdlib.KeysFunc,
		"log":             stdlib.LogFunc,
		"lower":           stdlib.LowerFunc,
		"max":             stdlib.MaxFunc,
		"merge":           stdlib.MergeFunc,
		"min":             stdlib.MinFunc,
		"parseint":        stdlib.ParseIntFunc,
		"pow":             stdlib.PowFunc,
		"range":           stdlib.RangeFunc,
		"regex":           stdlib.RegexFunc,
		"regexall":        stdlib.RegexAllFunc,
		"reverse":         stdlib.ReverseListFunc,
		"setintersection": stdlib.SetIntersectionFunc,
		"setproduct":      stdlib.SetProductFunc,
		"setsubtract":     stdlib.SetSubtractFunc,
		"setunion":        stdlib.SetUnionFunc,
		"signum":          stdlib.SignumFunc,
		"slice":           stdlib.SliceFunc,
		"sort":            stdlib.SortFunc,
		"split":           stdlib.SplitFunc,
		"strrev":          stdlib.ReverseFunc,
		"substr":          stdlib.SubstrFunc,
		"timeadd":         stdlib.TimeAddFunc,
		"title":           stdlib.TitleFunc,
		"trim":            stdlib.TrimFunc,
		"trimprefix":      stdlib.TrimPrefixFunc,
		"trimspace":       stdlib.TrimSpaceFunc,
		"trimsuffix":      stdlib.TrimSuffixFunc,
		"upper":           stdlib.UpperFunc,
		"values":          stdlib.ValuesFunc,
		"zipmap":          stdlib.ZipmapFunc,
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
