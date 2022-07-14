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

	"github.com/stretchr/testify/suite"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type BitfinexSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *BitfinexSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *BitfinexSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Bitfinex{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *BitfinexSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *BitfinexSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Bitfinex)
	suite.EqualValues("tBTCETH", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("tBTCUSD", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
	suite.EqualValues("tBTCUSD", ex.localPairName(Pair{Base: "BTC", Quote: "USDT"}))
}

func (suite *BitfinexSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrInvalidResponseStatus, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Bitfinex).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(fmt.Errorf("bad response: %w", ourErr), cr[0].Error)

	for n, r := range [][]byte{
		[]byte(``),
		[]byte(`{}`),
		[]byte(`["x",1,1,1,1,1,1,1,1,1,1]`),
		[]byte(`[[1,1,1,1,1,1,1,1,1,1,1]]`),
		[]byte(`[[true,1,1,1,1,1,1,1,1,1,1]]`),
		[]byte(`[["x",1,1,1,1,1,1,1,1,1,1]]`),
		[]byte(`[["x",1,1,1,1,1,1,1,1,1]]`),
		[]byte(`[["x",1,1,1,1,1,1,1,1,1,1,1]]`),
		[]byte(`[[1,"tBTCETH",1,1,1,1,1,1,1,1,1]]`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.origin.ExchangeHandler.(Bitfinex).Pool().(*query.MockWorkerPool).MockResp(resp)
			cr = suite.origin.Fetch([]Pair{pair})
			suite.Errorf(cr[0].Error, fmt.Sprintf("Case-%d", n+1))
		})
	}
}

func (suite *BitfinexSuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`[["tBTCETH",1.01,1.02,1.03,1.04,1.05,1.06,1.07,1.08,1.09,1.10]]`),
	}
	suite.origin.ExchangeHandler.(Bitfinex).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(1.01, cr[0].Price.Bid)
	suite.Equal(1.03, cr[0].Price.Ask)
	suite.Equal(1.07, cr[0].Price.Price)
	suite.Equal(1.08, cr[0].Price.Volume24h)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *BitfinexSuite) TestRealAPICall() {
	pairs := []Pair{
		{Base: "USDT", Quote: "USD"},
		{Base: "ETH", Quote: "BTC"},
		// {Base: "MKR", Quote: "ETH"},
		{Base: "ZRX", Quote: "USD"},
		{Base: "ETH", Quote: "USD"},
		// {Base: "DGX", Quote: "USDT"},
		{Base: "OMG", Quote: "USDT"},
	}
	testRealBatchAPICall(
		suite,
		NewBaseExchangeHandler(Bitfinex{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		pairs,
	)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestBitfinexSuite(t *testing.T) {
	suite.Run(t, new(BitfinexSuite))
}
