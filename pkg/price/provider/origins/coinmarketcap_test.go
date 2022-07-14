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

const successCoinmarketcapResponse = `
{
	"data": {
		"1": {
			"id": 1,
			"name": "Bitcoin",
			"symbol": "BTC",
			"slug": "bitcoin",
			"is_active": 1,
			"is_fiat": 0,
			"circulating_supply": 17199862,
			"total_supply": 17199862,
			"max_supply": 21000000,
			"date_added": "2013-04-28T00:00:00.000Z",
			"num_market_pairs": 331,
			"cmc_rank": 1,
			"last_updated": "2018-08-09T21:56:28.000Z",
			"tags": [
			"mineable"
			],
			"platform": null,
			"quote": {
				"USD": {
					"price": 6602.60701122,
					"volume_24h": 4314444687.5194,
					"percent_change_1h": 0.988615,
					"percent_change_24h": 4.37185,
					"percent_change_7d": -12.1352,
					"market_cap": 113563929433.21645,
					"last_updated": "2018-08-09T21:56:28.000Z"
				}
			}
		}
	},
	"status": {
		"timestamp": "2020-10-01T11:20:25.637Z",
		"error_code": 0,
		"error_message": "",
		"elapsed": 10,
		"credit_count": 1
	}
}
`

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CoinmarketcapSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *CoinmarketcapSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *CoinmarketcapSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(
		CoinMarketCap{WorkerPool: query.NewMockWorkerPool(), APIKey: "API_KEY"},
		nil,
	)
}

func (suite *CoinmarketcapSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *CoinmarketcapSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(CoinMarketCap)
	suite.EqualValues("825", ex.localPairName(Pair{Base: "USDT", Quote: "USD"}))
	suite.EqualValues("1", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *CoinmarketcapSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.origin.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "USDT", Quote: "USD"}
	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(CoinMarketCap).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(CoinMarketCap).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	suite.origin.ExchangeHandler.(CoinMarketCap).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error wrong code
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{}}`),
	}
	suite.origin.ExchangeHandler.(CoinMarketCap).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error wrong message
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{},"status":{error_code":1,"error_message":"Wrong"}}`),
	}
	suite.origin.ExchangeHandler.(CoinMarketCap).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error no data
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{},"status":{error_code":0,"error_message":""}}`),
	}
	suite.origin.ExchangeHandler.(CoinMarketCap).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
	// Error no pair in data
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{"1":{"quote":{}}},"status":{error_code":0,"error_message":""}}`),
	}
	suite.origin.ExchangeHandler.(CoinMarketCap).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *CoinmarketcapSuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "USD"}

	resp := &query.HTTPResponse{
		Body: []byte(successCoinmarketcapResponse),
	}
	suite.origin.ExchangeHandler.(CoinMarketCap).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})

	suite.NoError(cr[0].Error)
	suite.Equal(6602.60701122, cr[0].Price.Price)
	suite.Equal(4314444687.5194, cr[0].Price.Volume24h)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(2))
}

func (suite *CoinmarketcapSuite) TestRealAPICall() {
	testRealAPICall(
		suite,
		NewBaseExchangeHandler(CoinMarketCap{
			WorkerPool: query.NewHTTPWorkerPool(1),
			APIKey:     "API_KEY", // TODO: Somehow setup API key ?
		}, nil),
		"LRC", "ETH")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCoinmarketcapSuite(t *testing.T) {
	suite.Run(t, new(CoinmarketcapSuite))
}
