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

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type DdexSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *DdexSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *DdexSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Ddex{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *DdexSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *DdexSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Ddex)
	suite.EqualValues("BTC-ETH", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.NotEqual("BTC-USDC", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *DdexSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	// nil as response
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrInvalidResponseStatus, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Ddex).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(fmt.Errorf("bad response: %w", ourErr), cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Ddex).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	for n, r := range [][]byte{
		// invalid desc
		[]byte(`{
		   "status":0,
		   "desc":"err",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[
					{
					   "price":"11691.6",
					   "amount":"0.4173"
					}
				 ],
				 "asks":[
					{
					   "price":"11719.8",
					   "amount":"0.3709"
					}
				 ]
			  }
		   }
		}`),
		// invalid ask price
		[]byte(`{
		   "status":0,
		   "desc":"success",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[
					{
					   "price":"11691.6",
					   "amount":"0.4173"
					}
				 ],
				 "asks":[
					{
					   "price":"err",
					   "amount":"0.3709"
					}
				 ]
			  }
		   }
		}`),
		// invalid bid price
		[]byte(`{
		   "status":0,
		   "desc":"success",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[
					{
					   "price":"err",
					   "amount":"0.4173"
					}
				 ],
				 "asks":[
					{
					   "price":"11719.8",
					   "amount":"0.3709"
					}
				 ]
			  }
		   }
		}`),
		// empty order book
		[]byte(`{
		   "status":0,
		   "desc":"success",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[],
				 "asks":[]
			  }
		   }
		}`),
		[]byte(`{"status":1,"desc":"failure","template":"","params":null,"data":
		{"tickers":[
		{"marketId":"ETH-USDT","price":"362.64","volume":"6.75",
		"bid":"362.57","ask":"362.64","low":"362.64","high":"374.8","updateAt":1600239124811},
		{"marketId":"ETH-USDC","price":"364.96","volume":"11.9853",
		"bid":"363.76","ask":"364.96","low":"364.96","high":"364.96","updateAt":1600250097975},
		{"marketId":"ETH-DAI","price":"322.28","volume":"4.5",
		"bid":"322.53","ask":"322.63","low":"322.63","high":"322.53","updateAt":1599331939832},
		{"marketId":"WBTC-USDT","price":"10039.8","volume":"0.8867",
		"bid":"10039.8","ask":"10048.6","low":"10048.6","high":"10109","updateAt":1599369255015},
		{"marketId":"ETH-SAI","price":"145.48","volume":"3.6783",
		"bid":"145.48","ask":"149.41","low":"149.41","high":"149.35","updateAt":1575188948775}]}}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.origin.ExchangeHandler.(Ddex).Pool().(*query.MockWorkerPool).MockResp(resp)
			cr = suite.origin.Fetch([]Pair{pair})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *DdexSuite) TestSuccessResponse() {
	pair := Pair{Base: "ETH", Quote: "USDT"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"status":0,"desc":"success","template":"","params":null,"data":
		{"tickers":[
		{"marketId":"ETH-USDT","price":"362.64","volume":"6.75",
		"bid":"362.57","ask":"362.64","low":"362.64","high":"374.8","updateAt":2000},
		{"marketId":"USDT-ETH","price":"1","volume":"2",
		"bid":"3","ask":"4","low":"5","high":"6","updateAt":123},
		{"marketId":"ETH-USDC","price":"364.96","volume":"11.9853",
		"bid":"363.76","ask":"364.96","low":"364.96","high":"364.96","updateAt":1600250097975},
		{"marketId":"ETH-DAI","price":"322.28","volume":"4.5",
		"bid":"322.53","ask":"322.63","low":"322.63","high":"322.53","updateAt":1599331939832},
		{"marketId":"WBTC-USDT","price":"10039.8","volume":"0.8867",
		"bid":"10039.8","ask":"10048.6","low":"10048.6","high":"10109","updateAt":1599369255015},
		{"marketId":"ETH-SAI","price":"145.48","volume":"3.6783",
		"bid":"145.48","ask":"149.41","low":"149.41","high":"149.35","updateAt":1575188948775}]}}`),
	}
	suite.origin.ExchangeHandler.(Ddex).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(362.64, cr[0].Price.Ask)
	suite.Equal(362.57, cr[0].Price.Bid)
	suite.Equal(362.64, cr[0].Price.Price)
	suite.Equal(6.75, cr[0].Price.Volume24h)
	suite.Equal(cr[0].Price.Timestamp.Unix(), int64(2))
}

func (suite *DdexSuite) TestRealAPICall() {
	origin := NewBaseExchangeHandler(Ddex{WorkerPool: query.NewHTTPWorkerPool(1)}, nil)
	testRealAPICall(suite, origin, "WBTC", "USDT")
	pairs := []Pair{
		{Base: "ETH", Quote: "USDT"},
		{Base: "ETH", Quote: "USDC"},
		{Base: "ETH", Quote: "DAI"},
		{Base: "WBTC", Quote: "USDT"},
	}
	testRealBatchAPICall(suite, origin, pairs)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDdexSuite(t *testing.T) {
	suite.Run(t, new(DdexSuite))
}
