package hcl

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ParseFiles parses the HCL configuration files at the given paths. It returns
// a merged hcl.Body. The subject argument is optional. It is used to provide a
// range for the returned diagnostics.
func ParseFiles(paths []string, subject *hcl.Range) (hcl.Body, hcl.Diagnostics) {
	bodies := make([]hcl.Body, len(paths))
	for n, path := range paths {
		body, diags := ParseFile(path, subject)
		if diags.HasErrors() {
			return nil, diags
		}
		bodies[n] = body
	}
	return hcl.MergeBodies(bodies), nil
}

// ParseFile parses the given path into a hcl.File. The subject argument is
// optional. It is used to provide a range for the returned diagnostics.
func ParseFile(path string, subject *hcl.Range) (hcl.Body, hcl.Diagnostics) {
	src, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Failed to read configuration",
				Detail:   fmt.Sprintf("Cannot read file %s: file does not exist.", path),
				Subject:  subject,
			}}
		}
		return nil, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Failed to read configuration",
			Detail:   fmt.Sprintf("Cannot read file %s: %s.", path, err),
			Subject:  subject,
		}}
	}
	file, diags := hclsyntax.ParseConfig(src, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	return file.Body, nil
}
