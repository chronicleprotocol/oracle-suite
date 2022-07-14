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

package origins

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type OriginsSuite struct {
	suite.Suite
	pool *query.MockWorkerPool
	set  *Set
}

// Setup origin
func (suite *OriginsSuite) SetupSuite() {
	pool := query.NewMockWorkerPool()

	suite.pool = pool
	suite.set = NewSet(map[string]Handler{
		"binance": NewBaseExchangeHandler(Binance{WorkerPool: pool}, nil),
	})
}

func (suite *OriginsSuite) TestCallWithMissingOrigin() {
	cr := suite.set.Fetch(map[string][]Pair{"x": {{}}})
	assert.Error(suite.T(), cr["x"][0].Error)

	pair := Pair{Quote: "A", Base: "B"}
	cr = suite.set.Fetch(map[string][]Pair{"x": {pair}})

	assert.Equal(suite.T(), pair, cr["x"][0].Price.Pair)
	assert.Error(suite.T(), cr["x"][0].Error)
}

func (suite *OriginsSuite) TestFailWithNilResponseForBinance() {
	resp := &query.HTTPResponse{
		Body:  []byte{},
		Error: nil,
	}

	suite.pool.MockResp(resp)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	cr := suite.set.Fetch(map[string][]Pair{"binance": {pair}})

	assert.Error(suite.T(), cr["binance"][0].Error)
}

func (suite *OriginsSuite) TestSuccessBinance() {
	price := 0.024361
	json := fmt.Sprintf(`[{"symbol":"ETHBTC","lastPrice":"%f"}]`, price)
	resp := &query.HTTPResponse{
		Body:  []byte(json),
		Error: nil,
	}

	suite.pool.MockResp(resp)

	pair := Pair{Quote: "BTC", Base: "ETH"}
	cr := suite.set.Fetch(map[string][]Pair{"binance": {pair}})

	assert.NoError(suite.T(), cr["binance"][0].Error)
	assert.EqualValues(suite.T(), pair, cr["binance"][0].Price.Pair)
	assert.EqualValues(suite.T(), price, cr["binance"][0].Price.Price)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestOriginsSuite(t *testing.T) {
	suite.Run(t, new(OriginsSuite))
}

type mockExchangeHandler struct{}

func (u mockExchangeHandler) Pool() query.WorkerPool {
	return nil
}

func (u mockExchangeHandler) PullPrices(pairs []Pair) []FetchResult {
	var results []FetchResult
	for _, pair := range pairs {
		results = append(results, FetchResult{
			Price: Price{
				Pair:      pair,
				Price:     1,
				Ask:       2,
				Bid:       3,
				Volume24h: 4,
				Timestamp: time.Now(),
			},
		})
	}
	return results
}

func TestNewBaseExchangeHandlerWithoutAliases(t *testing.T) {
	pair := Pair{Base: "BTC", Quote: "ETH"}

	eh := NewBaseExchangeHandler(mockExchangeHandler{}, nil)
	assert.Nil(t, eh.aliases)

	results := eh.Fetch([]Pair{pair})
	assert.Len(t, results, 1)

	result := results[0]
	assert.NotNil(t, result)
	assert.Equal(t, pair, result.Price.Pair)
}

func TestBaseExchangeHandlerReplacement(t *testing.T) {
	aliases := SymbolAliases{"ETH": "WETH"}
	pair := Pair{Base: "BTC", Quote: "ETH"}

	mockHandler := mockExchangeHandler{}

	handler := BaseExchangeHandler{
		ExchangeHandler: mockHandler,
		aliases:         aliases,
	}

	results := handler.Fetch([]Pair{pair})
	assert.Len(t, results, 1)

	result := results[0]
	assert.NotNil(t, result)
	assert.Equal(t, pair, result.Price.Pair)
}

func TestGettingContractAddressByBaseAndQuote(t *testing.T) {
	contract := "0x0000"
	contracts := ContractAddresses{"BTC/ETH": contract}

	// Not existing pair
	address, _, ok := contracts.ByPair(Pair{Base: "BTC", Quote: "USD"})
	assert.Empty(t, address)
	assert.False(t, ok)

	// Existing direct pair
	address, _, ok = contracts.ByPair(Pair{Base: "BTC", Quote: "ETH"})
	assert.True(t, ok)
	assert.Equal(t, contract, address)

	// Existing reversed pair
	address, _, ok = contracts.ByPair(Pair{Base: "ETH", Quote: "BTC"})
	assert.True(t, ok)
	assert.Equal(t, contract, address)
}

func TestReplacingSymbolsUsingAliases(t *testing.T) {
	aliases := SymbolAliases{"ETH": "WETH"}

	// Should not be changed
	symbol, replaced := aliases.replaceSymbol("BTC")
	assert.Equal(t, "BTC", symbol)
	assert.False(t, replaced)

	// Should be replaced
	symbol, replaced = aliases.replaceSymbol("ETH")
	assert.Equal(t, "WETH", symbol)
	assert.True(t, replaced)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	replacedPair := aliases.replacePair(pair)
	assert.Equal(t, "BTC", replacedPair.Base)
	assert.Equal(t, "WETH", replacedPair.Quote)

	assert.False(t, replacedPair.baseReplaced)
	assert.True(t, replacedPair.quoteReplaced)
}

func TestReplacementAndRevertingUsingAliases(t *testing.T) {
	aliases := SymbolAliases{"ETH": "WETH"}

	// Symbol should be replaced
	symbol, replaced := aliases.replaceSymbol("ETH")
	assert.Equal(t, "WETH", symbol)
	assert.True(t, replaced)
	assert.Equal(t, "ETH", aliases.revertSymbol(symbol))

	// Replacing pair
	pair := Pair{Base: "BTC", Quote: "ETH"}
	replacedPair := aliases.replacePair(pair)
	assert.Equal(t, "BTC", replacedPair.Base)
	assert.Equal(t, "WETH", replacedPair.Quote)

	reverted := aliases.revertPair(replacedPair)
	assert.Equal(t, "BTC", reverted.Base)
	assert.Equal(t, "ETH", reverted.Quote)

	// Do not revert newly created pair
	reverted = aliases.revertPair(Pair{Base: "BTC", Quote: "WETH"})
	assert.Equal(t, "BTC", reverted.Base)
	assert.Equal(t, "WETH", reverted.Quote)
}
