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

type SushiswapSuite struct {
	suite.Suite
	origin *BaseExchangeHandler
}

func (suite *SushiswapSuite) Origin() Handler {
	return suite.origin
}

func (suite *SushiswapSuite) SetupSuite() {
	aliases := SymbolAliases{
		"ETH": "WETH",
		"BTC": "WBTC",
		"USD": "USDC",
	}
	addresses := ContractAddresses{
		"SNX/WETH": "0xa1d7b2d891e3a1f9ef4bbc5be20630c2feb1c470",
		"SNX/WBTC": "0xaabbccddeeffgghh0011223344556677889900aa",
		"CRV/WETH": "0x58dc5a51fe44589beb22e8ce67720b5bc5378009",
	}
	suite.origin = NewBaseExchangeHandler(
		Sushiswap{WorkerPool: query.NewMockWorkerPool(), ContractAddresses: addresses},
		aliases,
	)
}

func (suite *SushiswapSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "SNX", Quote: "WETH"}

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

	suite.origin.ExchangeHandler.(Sushiswap).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Sushiswap).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Error during converting price to a number
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pairs": [
						{
							"id": "0xa1d7b2d891e3a1f9ef4bbc5be20630c2feb1c470",
							"token0Price": "",
							"token1Price": "",
							"volumeToken0": "274940368.686748844780986508",
							"volumeToken1": "142365.832159709562349781",
							"token0": {
								"symbol": "SNX"
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
	suite.origin.ExchangeHandler.(Sushiswap).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Unable to find a pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pairs": [
						{
							"id": "0xaabbccddeeffgghh0011223344556677889900aa",
							"token0Price": "1560.208506522844994633814164798516",
							"token1Price": "0.0006409399742529590737926118088434103",
							"volumeToken0": "274940368.686748844780986508",
							"volumeToken1": "142365.832159709562349781",
							"token0": {
								"symbol": "SNX"
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
	suite.origin.ExchangeHandler.(Sushiswap).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *SushiswapSuite) TestSuccessResponse() {
	pairSNXWETH := Pair{Base: "SNX", Quote: "WETH"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pairs": [
						{
							"id": "0xa1d7b2d891e3a1f9ef4bbc5be20630c2feb1c470",
							"token0Price": "1560.2121",
							"token1Price": "0.0006",
							"volumeToken0": "274940368.6801",
							"volumeToken1": "142365.8321",
							"token0": {
								"symbol": "SNX"
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
	suite.origin.ExchangeHandler.(Sushiswap).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairSNXWETH})

	suite.Len(fr, 1)

	// SNX/WETH
	suite.NoError(fr[0].Error)
	suite.Equal(pairSNXWETH, fr[0].Price.Pair)
	suite.Equal(0.0006, fr[0].Price.Price)
	suite.Equal(0.0006, fr[0].Price.Bid)
	suite.Equal(0.0006, fr[0].Price.Ask)
	suite.Equal(274940368.6801, fr[0].Price.Volume24h)
	suite.Greater(fr[0].Price.Timestamp.Unix(), int64(0))

	pairCRVWETH := Pair{Base: "CRV", Quote: "WETH"}
	resp1 := &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pairs": [
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
	suite.origin.ExchangeHandler.(Sushiswap).Pool().(*query.MockWorkerPool).MockResp(resp1)
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

func (suite *SushiswapSuite) TestRealAPICall() {
	aliases := SymbolAliases{
		"ETH": "WETH",
		"BTC": "WBTC",
		"USD": "USDC",
	}
	addresses := ContractAddresses{
		"SNX/WETH": "0xa1d7b2d891e3a1f9ef4bbc5be20630c2feb1c470",
	}

	origin := NewBaseExchangeHandler(Sushiswap{
		WorkerPool:        query.NewHTTPWorkerPool(1),
		ContractAddresses: addresses,
	}, aliases)

	testRealBatchAPICall(
		suite,
		origin,
		[]Pair{
			{Base: "SNX", Quote: "ETH"},
		},
	)
}

func TestSushiswapSuite(t *testing.T) {
	suite.Run(t, new(SushiswapSuite))
}
