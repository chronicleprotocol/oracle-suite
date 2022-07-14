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
type KucoinSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *KucoinSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *KucoinSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Kucoin{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *KucoinSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KucoinSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Kucoin)
	suite.EqualValues("BTC-ETH", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.NotEqual("BTC-USDC", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *KucoinSuite) TestFailOnWrongInput() {
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
	suite.origin.ExchangeHandler.(Kucoin).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Kucoin).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	for n, r := range [][]byte{
		// invalid price
		[]byte(`{
			"code":"200000",
			"data": {
				"time":1596632420791,
				"sequence":"1594320230985",
				"price":"err",
				"size":"0.129",
				"bestBid":"139.55",
				"bestBidSize": "0.2866",
				"bestAsk":"139.7",
				"bestAskSize":"0.2863"
			}
		}`),
		// invalid bid price
		[]byte(`{
			"code":"200000",
			"data": {
				"time":1596632420791,
				"sequence":"1594320230985",
				"price":"1.23",
				"size":"0.129",
				"bestBid":"err",
				"bestBidSize": "0.2866",
				"bestAsk":"139.7",
				"bestAskSize":"0.2863"
			}
		}`),
		// invalid ask price
		[]byte(`{
			"code":"200000",
			"data": {
				"time":1596632420791,
				"sequence":"1594320230985",
				"price":"1.23",
				"size":"0.129",
				"bestBid":"139.55",
				"bestBidSize": "0.2866",
				"bestAsk":"err",
				"bestAskSize":"0.2863"
			}
		}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.origin.ExchangeHandler.(Kucoin).Pool().(*query.MockWorkerPool).MockResp(resp)
			cr = suite.origin.Fetch([]Pair{pair})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *KucoinSuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{
			"code":"200000",
			"data": {
				"time":1596632420791,
				"sequence":"1594320230985",
				"price":"1.23",
				"size":"0.123",
				"bestBid":"1.2",
				"bestBidSize": "0.2866",
				"bestAsk":"1.3",
				"bestAskSize":"0.2863"
			}
		}`),
	}
	suite.origin.ExchangeHandler.(Kucoin).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(int64(1596632420), cr[0].Price.Timestamp.Unix())
	suite.Equal(1.23, cr[0].Price.Price)
	suite.Equal(1.3, cr[0].Price.Bid)
	suite.Equal(1.2, cr[0].Price.Ask)
}

func (suite *KucoinSuite) TestRealAPICall() {
	testRealAPICall(
		suite,
		NewBaseExchangeHandler(Kucoin{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		"ETH",
		"BTC",
	)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKucoinSuite(t *testing.T) {
	suite.Run(t, new(KucoinSuite))
}
