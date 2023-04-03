package include

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"

	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
)

// Include merges the contents of multiple HCL files specified in the "include"
// attribute using glob patterns.
func Include(ctx *hcl.EvalContext, body hcl.Body, wd string, maxDeep int) (hcl.Body, hcl.Diagnostics) {
	// Decode the "include" attribute.
	content, remain, diags := body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{{Name: "include"}},
	})
	attribute := content.Attributes["include"]
	if diags.HasErrors() || attribute == nil {
		return body, diags
	}
	var include []string
	if diags = utilHCL.DecodeExpression(ctx, attribute.Expr, &include); diags.HasErrors() {
		return nil, diags
	}

	// Check for too many nested includes.
	if maxDeep <= 0 {
		return nil, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Too many nested includes",
			Detail:   "Too many nested includes. Check your glob patterns.",
			Subject:  attribute.Expr.Range().Ptr(),
		}}
	}

	// Iterate over the glob patterns.
	var bodies []hcl.Body
	for _, pattern := range include {
		// Find all files matching the glob pattern.
		paths, err := glob(pattern)
		if err != nil {
			return nil, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Invalid glob pattern",
				Detail:   fmt.Sprintf("Invalid glob pattern %s: %s.", pattern, err),
				Subject:  attribute.Expr.Range().Ptr(),
			}}
		}

		// Iterate over the files from the glob pattern.
		for _, path := range paths {
			path = relativePath(wd, path)

			// Parse the file.
			fileBody, diags := utilHCL.ParseFile(path, attribute.Expr.Range().Ptr())
			if diags.HasErrors() {
				return nil, diags
			}

			// Recursively include files.
			body, diags := Include(ctx, fileBody, filepath.Dir(path), maxDeep-1)
			if diags.HasErrors() {
				return nil, diags
			}
			bodies = append(bodies, body)
		}
	}

	// Merge the body of the main file with the bodies of the included files.
	return hcl.MergeBodies([]hcl.Body{remain, hcl.MergeBodies(bodies)}), diags
}

// relativePath returns an absolute path for the given path relative to the
// given working directory.
func relativePath(wd, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(wd, path)
}

// glob is a wrapper around filepath.Glob that returns the given pattern
// if it does not contain any glob characters.
func glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "*") {
		return []string{pattern}, nil
	}
	return filepath.Glob(pattern)
}
