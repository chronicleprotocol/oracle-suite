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
	return ParseSource(path, src)
}

// ParseSources parses the HCL configuration sources. It returns a merged
// hcl.Body.
func ParseSources(srcs [][]byte) (hcl.Body, hcl.Diagnostics) {
	bodies := make([]hcl.Body, len(srcs))
	for n := 0; n < len(srcs); n++ {
		body, diags := ParseSource(fmt.Sprintf("embeded #%d", n), srcs[n])
		if diags.HasErrors() {
			return nil, diags
		}
		bodies[n] = body
	}
	return hcl.MergeBodies(bodies), nil
}

// ParseSource parses the given source into a hcl.File.
func ParseSource(name string, src []byte) (hcl.Body, hcl.Diagnostics) {
	file, diags := hclsyntax.ParseConfig(src, name, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	return file.Body, nil
}
