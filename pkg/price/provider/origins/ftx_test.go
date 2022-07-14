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
type FtxSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *FtxSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *FtxSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Ftx{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *FtxSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *FtxSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Ftx)
	suite.EqualValues("BTC/ETH", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("BTC/USDC", ex.localPairName(Pair{Base: "BTC", Quote: "USDC"}))
}

func (suite *FtxSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	var cr []FetchResult
	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrInvalidResponseStatus, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Ftx).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(fmt.Errorf("bad response: %w", ourErr), cr[0].Error)

	for n, r := range [][]byte{
		// invalid response
		[]byte(``),
		// invalid response
		[]byte(`{}`),
		// invalid success
		[]byte(`{"success":false}`),
		// invalid response
		[]byte(`{"success":true}`),
		// invalid response
		[]byte(`{"success":true,"result":{}}`),
		[]byte(`{"success":true,"result":[]}`),
		// invalid name
		[]byte(`{"success":true,"result":[{"name":"SOME/ANOTHER"}]}`),
		// invalid price (string)
		[]byte(`{"success":true,"result":[{"name":"BTC/ETH","last":"1"}]}`),
		// invalid ask (string)
		[]byte(`{"success":true,"result":[{"name":"BTC/ETH","last":1,"ask":"2"}]}`),
		// invalid bid (string)
		[]byte(`{"success":true,"result":[{"name":"BTC/ETH","last":1,"ask":2,"bid":"3"}]}`),
		// invalid volume (string)
		[]byte(`{"success":true,"result":[{"name":"BTC/ETH","last":1,"ask":2,"bid":3,"quoteVolume24h":"4"}]}`),
		// invalid success with normal result
		[]byte(`{"success":false,"result":[{"name":"BTC/ETH","last":1,"ask":2,"bid":3,"quoteVolume24h":4}]}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.origin.ExchangeHandler.(Ftx).Pool().(*query.MockWorkerPool).MockResp(resp)
			cr = suite.origin.Fetch([]Pair{pair})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *FtxSuite) TestSuccessResponse() {
	pair := Pair{Base: "ETH", Quote: "USD"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"result": [{
"ask": 380.38,
"baseCurrency": "ETH",
"bid": 380.25,
"change1h": 0.004915563307698406,
"change24h": 0.04470025825594813,
"changeBod": 0.04550453670607644,
"enabled": true,
"last": 380.23,
"minProvideSize": 0.001,
"name": "ETH/USD",
"postOnly": false,
"price": 380.25,
"priceIncrement": 0.01,
"quoteCurrency": "USD",
"quoteVolume24h": 12467473.8244,
"restricted": false,
"sizeIncrement": 0.001,
"type": "spot",
"underlying": null,
"volumeUsd24h": 12467473.8244
}],"success":true}`),
	}
	suite.origin.ExchangeHandler.(Ftx).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(380.23, cr[0].Price.Price)
	suite.Equal(380.38, cr[0].Price.Ask)
	suite.Equal(380.25, cr[0].Price.Bid)
	suite.Equal(12467473.8244, cr[0].Price.Volume24h)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *FtxSuite) TestRealAPICall() {
	origin := NewBaseExchangeHandler(Ftx{WorkerPool: query.NewHTTPWorkerPool(1)}, nil)

	testRealAPICall(suite, origin, "ETH", "BTC")
	pairs := []Pair{
		{Base: "ETH", Quote: "USDT"},
		{Base: "BTC", Quote: "USDT"},
	}
	testRealBatchAPICall(suite, origin, pairs)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFtxSuite(t *testing.T) {
	suite.Run(t, new(FtxSuite))
}
