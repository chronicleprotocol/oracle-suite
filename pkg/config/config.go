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

package config

import (
	"fmt"
	"os"
	"path/filepath"
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
		"can":      tryfunc.CanFunc,
		"length":   stdlib.LengthFunc,
		"range":    stdlib.RangeFunc,
		"replace":  stdlib.ReplaceFunc,
		"join":     stdlib.JoinFunc,
		"merge":    stdlib.MergeFunc,
		"concat":   stdlib.ConcatFunc,
		"try":      tryfunc.TryFunc,
		"contains": stdlib.ContainsFunc,
		"keys":     stdlib.KeysFunc,
		"distinct": stdlib.DistinctFunc,

		// Custom functions:
		"tobool":   funcs.MakeToFunc(cty.Bool),
		"tolist":   funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":    funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
		"tonumber": funcs.MakeToFunc(cty.Number),
		"toset":    funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tostring": funcs.MakeToFunc(cty.String),

		// Even more customer functions:
		"env":     envFunc,
		"explode": explodeFunc,
	},
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

var envFunc = function.New(&function.Spec{
	Description: `Returns the environment variable for a given key.`,
	Params: []function.Parameter{
		{Name: "key", Type: cty.String},
		{Name: "default", Type: cty.String},
	},
	Type: func(args []cty.Value) (ret cty.Type, err error) {
		return cty.String, nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		key := args[0].AsString()
		def := args[1].AsString()
		val, ok := os.LookupEnv(key)
		if !ok {
			val = def
		}
		return cty.StringVal(val), nil
	},
})

var explodeFunc = function.New(&function.Spec{
	Description: "Produces a list of one or more strings by splitting the given string at all instances of a given separator substring. Empty string becomes an empty List", //nolint:lll
	Params: []function.Parameter{
		{
			Name:        "separator",
			Description: "The substring that delimits the result strings.",
			Type:        cty.String,
		},
		{
			Name:        "str",
			Description: "The string to split.",
			Type:        cty.String,
		},
	},
	Type: function.StaticReturnType(cty.List(cty.String)),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[1].AsString()
		if len(str) == 0 {
			return cty.ListValEmpty(cty.String), nil
		}
		sep := args[0].AsString()
		elems := strings.Split(str, sep)
		elemVals := make([]cty.Value, len(elems))
		for i, s := range elems {
			elemVals[i] = cty.StringVal(s)
		}
		if len(elemVals) == 0 {
			return cty.ListValEmpty(cty.String), nil
		}
		return cty.ListVal(elemVals), nil
	},
})

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
	if len(paths) > 0 {
		wd = filepath.Dir(paths[0])
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

// LoadEmbeds populates config with data from []utilHCL.NamedBytes into the given config,
// and expanding dynamic blocks before decoding the HCL content.
func LoadEmbeds(config any, embeds [][]byte) (err error) {
	var body hcl.Body
	var diags hcl.Diagnostics
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if body, diags = utilHCL.ParseSources(embeds); diags.HasErrors() {
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

type HasDefaults interface {
	DefaultEmbeds() [][]byte
}
