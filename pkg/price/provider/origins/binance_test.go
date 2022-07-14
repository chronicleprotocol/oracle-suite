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

type BinanceSuite struct {
	suite.Suite
	origin *BaseExchangeHandler
}

func (suite *BinanceSuite) Origin() Handler {
	return suite.origin
}

func (suite *BinanceSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Binance{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *BinanceSuite) TestLocalPair() {
	binance := suite.origin.ExchangeHandler.(Binance)
	suite.EqualValues("BTCETH", binance.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.NotEqual("BTCUSDC", binance.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *BinanceSuite) TestFailOnWrongInput() {
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

	suite.origin.ExchangeHandler.(Binance).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Binance).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Error during converting price to a number
	resp = &query.HTTPResponse{
		Body: []byte(`
			[
			   {
				  "symbol":"BTCETH",
				  "lastPrice":"abc",
				  "bidPrice":"0",
				  "askPrice":"0",
				  "volume":"0",
				  "closeTime":"10000"
			   }
			]
		`),
	}
	suite.origin.ExchangeHandler.(Binance).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Unable to find a pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			[
			   {
				  "symbol":"AAABBB",
				  "lastPrice":"0",
				  "bidPrice":"0",
				  "askPrice":"0",
				  "volume":"0",
				  "closeTime":"10000"
			   }
			]
		`),
	}
	suite.origin.ExchangeHandler.(Binance).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *BinanceSuite) TestSuccessResponse() {
	pairBTCETH := Pair{Base: "BTC", Quote: "ETH"}
	pairBTCUSD := Pair{Base: "BTC", Quote: "USD"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			[
			   {
				  "symbol":"BTCETH",
				  "lastPrice":"1.1",
				  "bidPrice":"1.0",
				  "askPrice":"1.3",
				  "volume":"10.1",
				  "closeTime":10000
			   },
			   {
				  "symbol":"BTCUSD",
				  "lastPrice":"2.1",
				  "bidPrice":"2.0",
				  "askPrice":"2.3",
				  "volume":"20.1",
				  "closeTime":10000
			   },
			   {
				  "symbol":"BTCEUR",
				  "lastPrice":"3.1",
				  "bidPrice":"3.0",
				  "askPrice":"3.3",
				  "volume":"30.1",
				  "closeTime":10000
			   }
			]
		`),
	}
	suite.origin.ExchangeHandler.(Binance).Pool().(*query.MockWorkerPool).MockResp(resp)
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

func (suite *BinanceSuite) TestRealAPICall() {
	testRealBatchAPICall(
		suite,
		NewBaseExchangeHandler(Binance{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		[]Pair{
			{Base: "BAT", Quote: "BTC"},
			{Base: "COMP", Quote: "USDT"},
			{Base: "ETH", Quote: "BTC"},
			// {Base: "GNT", Quote: "BTC"},
			{Base: "KNC", Quote: "BTC"},
			// {Base: "LEND", Quote: "BTC"},
			{Base: "LINK", Quote: "BTC"},
			{Base: "LRC", Quote: "BTC"},
			{Base: "MANA", Quote: "BTC"},
			{Base: "OMG", Quote: "BTC"},
			{Base: "POLY", Quote: "BTC"},
			{Base: "REP", Quote: "BTC"},
			{Base: "SNT", Quote: "BTC"},
			{Base: "BTC", Quote: "USDT"},
			{Base: "ZRX", Quote: "BTC"},
		},
	)
}

func TestBinanceSuite(t *testing.T) {
	suite.Run(t, new(BinanceSuite))
}
