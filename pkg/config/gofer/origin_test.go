//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package gofer

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
)

func TestParsingOriginParamsAliases(t *testing.T) {
	// Empty aliases
	parsed := parseParamsSymbolAliases(hclToAny(t, `{}`))
	assert.Nil(t, parsed)

	// API key
	key := parseParamsAPIKey(hclToAny(t, `{api_key: "test"}`))
	assert.Equal(t, "test", key)

	// URL
	url := parseParamsURL(hclToAny(t, `{url: "test"}`))
	assert.Equal(t, "test", url)

	// Parse contracts
	contracts := parseParamsContracts(hclToAny(t, `{contracts: {"BTC/ETH":"0x00000"}}`))
	assert.NotNil(t, contracts)
	assert.Equal(t, "0x00000", contracts["BTC/ETH"])

	// Symbol aliases
	aliases := parseParamsSymbolAliases(hclToAny(t, `{symbol_aliases: {"ETH":"WETH"}}`))
	assert.NotNil(t, aliases)
	assert.Equal(t, "WETH", aliases["ETH"])
}

func hclToAny(t *testing.T, body string) any {
	t.Helper()
	expr, diags := hclsyntax.ParseExpression([]byte(body), "file.hcl", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("failed to parse HCL: %v", diags)
	}
	val, diags := expr.Value(config.HCLContext)
	if diags.HasErrors() {
		t.Fatalf("failed to evaluate HCL: %v", diags)
	}
	anyVal, err := ctyToAny(val)
	if err != nil {
		t.Fatalf("failed to convert HCL: %v", err)
	}
	return anyVal
}
