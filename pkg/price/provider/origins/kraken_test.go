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
type KrakenSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *KrakenSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *KrakenSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Kraken{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *KrakenSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KrakenSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Kraken)
	suite.EqualValues("BTC/ETH", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("BTC/USD", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *KrakenSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.origin.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "DAI", Quote: "USD"}
	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrInvalidResponseStatus, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Kraken).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(fmt.Errorf("bad response: %w", ourErr), cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Kraken).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":["abcd"]}`),
	}
	suite.origin.ExchangeHandler.(Kraken).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":[], "result":{}}`),
	}
	suite.origin.ExchangeHandler.(Kraken).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":[], "result":{"XDAIZUSD":{}})`),
	}
	suite.origin.ExchangeHandler.(Kraken).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *KrakenSuite) TestSuccessResponse() {
	pair := Pair{Base: "DAI", Quote: "USD"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"error":[],"result":{"DAI/USD":{"c":["1"],"v":["2"]}}}`),
	}
	suite.origin.ExchangeHandler.(Kraken).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Price.Price)
	suite.Equal(2.0, cr[0].Price.Volume24h)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *KrakenSuite) TestRealAPICall() {
	pairs := []Pair{
		{Base: "ETH", Quote: "BTC"},
		{Base: "ETH", Quote: "USD"},
		{Base: "BTC", Quote: "USD"},
		{Base: "LINK", Quote: "ETH"},
		{Base: "REP", Quote: "EUR"},
		{Base: "USDT", Quote: "USD"},
	}
	testRealBatchAPICall(suite, NewBaseExchangeHandler(
		Kraken{WorkerPool: query.NewHTTPWorkerPool(1)},
		nil,
	), pairs)
}

func TestKrakenSuite(t *testing.T) {
	suite.Run(t, new(KrakenSuite))
}
