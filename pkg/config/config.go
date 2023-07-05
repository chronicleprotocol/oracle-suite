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
		"can":     tryfunc.CanFunc,
		"length":  stdlib.LengthFunc,
		"range":   stdlib.RangeFunc,
		"replace": stdlib.ReplaceFunc,
		"split":   stdlib.SplitFunc,
		"merge":   stdlib.MergeFunc,
		"concat":  stdlib.ConcatFunc,
		"try":     tryfunc.TryFunc,

		// Custom functions:
		"tobool":   funcs.MakeToFunc(cty.Bool),
		"tolist":   funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":    funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
		"tonumber": funcs.MakeToFunc(cty.Number),
		"toset":    funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tostring": funcs.MakeToFunc(cty.String),
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
