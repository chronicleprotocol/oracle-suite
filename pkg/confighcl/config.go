package confighcl

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func LoadFile(config any, path string) error {
	ctx := &hcl.EvalContext{
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
	return hclsimple.DecodeFile(path, ctx, config)
}

func getEnvVars() cty.Value {
	envs := map[string]cty.Value{}
	for _, env := range os.Environ() {
		idx := strings.Index(env, "=")
		envs[env[:idx]] = cty.StringVal(env[idx+1:])
	}
	return cty.ObjectVal(envs)
}
