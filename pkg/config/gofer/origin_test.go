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

	"github.com/stretchr/testify/assert"
)

func TestParsingOriginParamsAliasesFailParsing(t *testing.T) {
	parsed, err := parseParamsSymbolAliases(nil)
	assert.Nil(t, parsed)
	assert.Error(t, err)

	parsed, err = parseParamsSymbolAliases([]byte(""))
	assert.Nil(t, parsed)
	assert.Error(t, err)
}

func TestParsingOriginParamsAliases(t *testing.T) {
	// parsing empty aliases
	parsed, err := parseParamsSymbolAliases([]byte(`{}`))
	assert.NoError(t, err)
	assert.Nil(t, parsed)

	// Parsing only apiKey
	key, err := parseParamsAPIKey([]byte(`{"apiKey":"test"}`))
	assert.NoError(t, err)
	assert.Equal(t, "test", key)

	// Parsing contracts
	contracts, err := parseParamsContracts([]byte(`{"contracts":{"BTC/ETH":"0x00000"}}`))
	assert.NoError(t, err)
	assert.NotNil(t, contracts)
	assert.Equal(t, "0x00000", contracts["BTC/ETH"])

	// Parsing symbol aliases
	aliases, err := parseParamsSymbolAliases([]byte(`{"symbolAliases":{"ETH":"WETH"}}`))
	assert.NoError(t, err)
	assert.NotNil(t, aliases)
	assert.Equal(t, "WETH", aliases["ETH"])
}
