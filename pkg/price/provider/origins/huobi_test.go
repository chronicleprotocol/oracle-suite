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
type HuobiSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *HuobiSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *HuobiSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Huobi{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *HuobiSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *HuobiSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Huobi)
	suite.EqualValues("btceth", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("btcusd", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *HuobiSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.origin.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Huobi).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Huobi).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"status":"error"}`),
	}
	suite.origin.ExchangeHandler.(Huobi).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"status":"success","vol":"abc"}`),
	}
	suite.origin.ExchangeHandler.(Huobi).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"status":"success","data":[],"ts":"abc"}`),
	}
	suite.origin.ExchangeHandler.(Huobi).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"status":"success","ts":2,"data":[{"bid":"abc"}]}`),
	}
	suite.origin.ExchangeHandler.(Huobi).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *HuobiSuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"status":"success","ts":2000,"data":[{"symbol":"btceth","ask":1,"bid":2.1,"vol":1.3}]}`),
	}
	suite.origin.ExchangeHandler.(Huobi).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})

	suite.NoError(cr[0].Error)
	suite.Equal(1.3, cr[0].Price.Volume24h)
	suite.Equal(1.0, cr[0].Price.Ask)
	suite.Equal(2.1, cr[0].Price.Bid)
	suite.Equal(cr[0].Price.Timestamp.Unix(), int64(2))
}

func (suite *HuobiSuite) TestRealAPICall() {
	huobi := NewBaseExchangeHandler(Huobi{WorkerPool: query.NewHTTPWorkerPool(1)}, nil)
	testRealAPICall(suite, huobi, "ETH", "BTC")
	testRealBatchAPICall(suite, huobi, []Pair{
		{Base: "SNT", Quote: "USDT"},
		{Base: "SNX", Quote: "USDT"},
		{Base: "YFI", Quote: "USDT"},
		{Base: "ETH", Quote: "BTC"},
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHuobiSuite(t *testing.T) {
	suite.Run(t, new(HuobiSuite))
}
