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

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"

	"github.com/stretchr/testify/suite"
)

type UniswapV3Suite struct {
	suite.Suite
	origin *BaseExchangeHandler
}

func (suite *UniswapV3Suite) Origin() Handler {
	return suite.origin
}

func (suite *UniswapV3Suite) SetupSuite() {
	aliases := SymbolAliases{
		"ETH": "WETH",
		"BTC": "WBTC",
	}
	addresses := ContractAddresses{
		"YFI/WETH": "0x04916039b1f59d9745bf6e0a21f191d1e0a84287",
		"YFI/WBTC": "0x04916039b1f59d9745bf6e0a21f191d1e0a84287",
		"CRV/WETH": "0x58dc5a51fe44589beb22e8ce67720b5bc5378009",
	}
	suite.origin = NewBaseExchangeHandler(
		UniswapV3{WorkerPool: query.NewMockWorkerPool(), ContractAddresses: addresses},
		aliases,
	)
}

func (suite *UniswapV3Suite) TestFailOnWrongInput() {
	pair := Pair{Base: "YFI", Quote: "ETH"}

	// Wrong pair
	fr := suite.origin.Fetch([]Pair{{}})
	suite.Error(fr[0].Error)

	// Nil as a response
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, fr[0].Error)

	// Error in a response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}

	suite.origin.ExchangeHandler.(UniswapV3).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(UniswapV3).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Error during converting price to a number
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pools": [
						{
							"id": "0x04916039b1f59d9745bf6e0a21f191d1e0a84287",
							"token0Price": "",
							"token1Price": "",
							"volumeToken0": "31.001552898956395468",
                			"volumeToken1": "-402.068307058139831158",
							"token0": {
								"symbol": "YFI"
							},
							"token1": {
								"symbol": "WETH"
							}
						}
					]
				}
			}
		`),
	}
	suite.origin.ExchangeHandler.(UniswapV3).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Unable to find a pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pools": [
						{
							"id": "0x04916039b1f59d9745bf6e0a21f191d1e0a84287",
							"token0Price": "0.06624583662031174276461684468775496",
							"token1Price": "15.09528826289120164642035869260895",
							"volumeToken0": "31.001552898956395468",
                			"volumeToken1": "-402.068307058139831158",
							"token0": {
								"symbol": "YFI"
							},
							"token1": {
								"symbol": "WBTC"
							}
						}
					]
				}
			}
		`),
	}
	suite.origin.ExchangeHandler.(UniswapV3).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *UniswapV3Suite) TestSuccessResponse() {
	pairYFIWETH := Pair{Base: "YFI", Quote: "ETH"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pools": [
						{
							"id": "0x04916039b1f59d9745bf6e0a21f191d1e0a84287",
							"token0": {
								"symbol": "YFI"
							},
							"token0Price": "0.0662",
							"token1": {
								"symbol": "WETH"
							},
							"token1Price": "15.0952",
							"volumeToken0": "31.00155",
							"volumeToken1": "-402.0683"
						}
					]
				}
			}
		`),
	}
	suite.origin.ExchangeHandler.(UniswapV3).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairYFIWETH})

	suite.Len(fr, 1)

	// SNX/WETH
	suite.NoError(fr[0].Error)
	suite.Equal(pairYFIWETH, fr[0].Price.Pair)
	suite.Equal(15.0952, fr[0].Price.Price)
	suite.Equal(15.0952, fr[0].Price.Bid)
	suite.Equal(15.0952, fr[0].Price.Ask)
	suite.Equal(31.00155, fr[0].Price.Volume24h)
	suite.Greater(fr[0].Price.Timestamp.Unix(), int64(0))

	pairCRVWETH := Pair{Base: "CRV", Quote: "ETH"}
	resp1 := &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pools": [
						{
							"id": "0x58dc5a51fe44589beb22e8ce67720b5bc5378009",
							"token0Price": "0.0006",
							"token1Price": "1560.2121",
							"volumeToken0": "142365.8321",
							"volumeToken1": "274940368.6801",
							"token0": {
								"symbol": "WETH"
							},
							"token1": {
								"symbol": "CRV"
							}
						}
					]
				}
			}
		`),
	}
	suite.origin.ExchangeHandler.(UniswapV3).Pool().(*query.MockWorkerPool).MockResp(resp1)
	fr1 := suite.origin.Fetch([]Pair{pairCRVWETH})

	suite.Len(fr1, 1)

	// CRV/WETH
	suite.NoError(fr1[0].Error)
	suite.Equal(pairCRVWETH, fr1[0].Price.Pair)
	suite.Equal(0.0006, fr1[0].Price.Price)
	suite.Equal(0.0006, fr1[0].Price.Bid)
	suite.Equal(0.0006, fr1[0].Price.Ask)
	suite.Equal(274940368.6801, fr1[0].Price.Volume24h)
	suite.Greater(fr1[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *UniswapV3Suite) TestRealAPICall() {
	aliases := SymbolAliases{
		"ETH": "WETH",
	}
	addresses := ContractAddresses{
		"YFI/WETH": "0x04916039b1f59d9745bf6e0a21f191d1e0a84287",
	}
	origin := NewBaseExchangeHandler(UniswapV3{
		WorkerPool:        query.NewHTTPWorkerPool(1),
		ContractAddresses: addresses,
	}, aliases)

	testRealBatchAPICall(
		suite,
		origin,
		[]Pair{
			{Base: "YFI", Quote: "WETH"},
		},
	)
}

func TestUniswapV3Suite(t *testing.T) {
	suite.Run(t, new(UniswapV3Suite))
}
