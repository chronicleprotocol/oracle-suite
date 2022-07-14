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

type OkexSuite struct {
	suite.Suite
	origin *BaseExchangeHandler
}

func (suite *OkexSuite) Origin() Handler {
	return suite.origin
}

func (suite *OkexSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Okex{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *OkexSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Okex)
	suite.EqualValues("BTC-ETH", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.NotEqual("BTC-USDC", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *OkexSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}

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

	suite.origin.ExchangeHandler.(Okex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Okex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Error during converting price to a number
	resp = &query.HTTPResponse{
		Body: []byte(`
			[
			   {
				  "instrument_id":"BTC-ETH",
				  "last":"abc",
				  "best_bid":"0",
				  "best_ask":"0",
				  "base_volume_24h":"0",
				  "timestamp":"2020-09-24T14:02:39.877Z"
			   }
			]
		`),
	}
	suite.origin.ExchangeHandler.(Okex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Unable to find a pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			[
			   {
				  "instrument_id":"AAA-BBB",
				  "last":"0",
				  "best_bid":"0",
				  "best_ask":"0",
				  "base_volume_24h":"0",
				  "timestamp":"2020-09-24T14:02:39.877Z"
			   }
			]
		`),
	}
	suite.origin.ExchangeHandler.(Okex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *OkexSuite) TestSuccessResponse() {
	pairBTCETH := Pair{Base: "BTC", Quote: "ETH"}
	pairBTCUSD := Pair{Base: "BTC", Quote: "USD"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			[
			   {
				  "instrument_id":"BTC-ETH",
				  "last":"1.1",
				  "best_bid":"1.0",
				  "best_ask":"1.3",
				  "base_volume_24h":"10.1",
				  "timestamp":"2020-09-24T14:02:39.877Z"
			   },
			   {
				  "instrument_id":"BTC-USD",
				  "last":"2.1",
				  "best_bid":"2.0",
				  "best_ask":"2.3",
				  "base_volume_24h":"20.1",
				  "timestamp":"2020-09-24T14:02:39.877Z"
			   },
			   {
				  "instrument_id":"BTC-EUR",
				  "last":"3.1",
				  "best_bid":"3.0",
				  "best_ask":"3.3",
				  "base_volume_24h":"30.1",
				  "timestamp":"2020-09-24T14:02:39.877Z"
			   }
			]
		`),
	}
	suite.origin.ExchangeHandler.(Okex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairBTCETH, pairBTCUSD})

	suite.Len(fr, 2)

	// BTC/ETH
	suite.NoError(fr[0].Error)
	suite.Equal(pairBTCETH, fr[0].Price.Pair)
	suite.Equal(1.1, fr[0].Price.Price)
	suite.Equal(1.0, fr[0].Price.Bid)
	suite.Equal(1.3, fr[0].Price.Ask)
	suite.Equal(10.1, fr[0].Price.Volume24h)
	suite.Greater(fr[0].Price.Timestamp.Unix(), int64(0))

	// BTC/USD
	suite.NoError(fr[1].Error)
	suite.Equal(pairBTCUSD, fr[1].Price.Pair)
	suite.Equal(2.1, fr[1].Price.Price)
	suite.Equal(2.0, fr[1].Price.Bid)
	suite.Equal(2.3, fr[1].Price.Ask)
	suite.Equal(20.1, fr[1].Price.Volume24h)
	suite.Greater(fr[1].Price.Timestamp.Unix(), int64(0))
}

func (suite *OkexSuite) TestRealAPICall() {
	testRealBatchAPICall(
		suite,
		NewBaseExchangeHandler(Okex{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		[]Pair{
			{Base: "LRC", Quote: "USDT"},
			{Base: "MKR", Quote: "BTC"},
			{Base: "ZRX", Quote: "BTC"},
			{Base: "COMP", Quote: "USDT"},
			{Base: "SNT", Quote: "USDT"},
			{Base: "BTC", Quote: "USDT"},
		},
	)
}

func TestOkexSuite(t *testing.T) {
	suite.Run(t, new(OkexSuite))
}
