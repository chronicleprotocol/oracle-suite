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

const successResponse = `{
	"tickers":[
		[
			"LRC-ETH",
			"1618137071822",
			"1021326019000000000000000",
			"270371208810000000000",
			"0.00026440",
			"0.00026900",
			"0.00026191",
			"0.00026700",
			"727",
			"0.00026699",
			"0.00026940",
			"",""
		],
		[
			"LRC-USDT",
			"1618137071822",
			"766306355200000000000000",
			"438371710590",
			"0.5688",
			"0.5880",
			"0.5501",
			"0.5742",
			"694",
			"0.5743",
			"0.5757",
			"",""
		]
	]
}`

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type LoopringSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *LoopringSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *LoopringSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Loopring{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *LoopringSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *LoopringSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Loopring)
	suite.EqualValues("USDT-DAI", ex.localPairName(Pair{Base: "USDT", Quote: "DAI"}))
	suite.EqualValues("ETH-DAI", ex.localPairName(Pair{Base: "ETH", Quote: "DAI"}))
}

func (suite *LoopringSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.origin.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "LRC", Quote: "USDT"}
	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Loopring).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Loopring).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	suite.origin.ExchangeHandler.(Loopring).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error wrong code
	resp = &query.HTTPResponse{
		Body: []byte(`{"tickers":{}}`),
	}
	suite.origin.ExchangeHandler.(Loopring).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error wrong message
	resp = &query.HTTPResponse{
		Body: []byte(`{"tickers":[]}`),
	}
	suite.origin.ExchangeHandler.(Loopring).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error no data
	resp = &query.HTTPResponse{
		Body: []byte(`{"tickers":[[]]}`),
	}
	suite.origin.ExchangeHandler.(Loopring).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
	// Error no pair in data
	resp = &query.HTTPResponse{
		Body: []byte(`{"tickers":[
			["LRC-USDT"]
		]}`),
	}
	suite.origin.ExchangeHandler.(Loopring).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *LoopringSuite) TestSuccessResponse() {
	pair := Pair{Base: "LRC", Quote: "ETH"}
	pair2 := Pair{Base: "LRC", Quote: "USDT"}

	resp := &query.HTTPResponse{
		Body: []byte(successResponse),
	}
	suite.origin.ExchangeHandler.(Loopring).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair, pair2})

	suite.NoError(cr[0].Error)
	suite.Equal(0.000267, cr[0].Price.Price)
	suite.Equal(0.0002694, cr[0].Price.Ask)
	suite.Equal(0.00026699, cr[0].Price.Bid)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(2))

	suite.NoError(cr[1].Error)
	suite.Equal(0.5742, cr[1].Price.Price)
	suite.Equal(0.5757, cr[1].Price.Ask)
	suite.Equal(0.5743, cr[1].Price.Bid)
	suite.Greater(cr[1].Price.Timestamp.Unix(), int64(2))
}

func (suite *LoopringSuite) TestRealAPICall() {
	testRealAPICall(
		suite,
		NewBaseExchangeHandler(Loopring{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		"LRC",
		"ETH",
	)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestLoopringSuite(t *testing.T) {
	suite.Run(t, new(LoopringSuite))
}
