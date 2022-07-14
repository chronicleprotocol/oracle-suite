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

type PoloniexSuite struct {
	suite.Suite
	origin *BaseExchangeHandler
}

func (suite *PoloniexSuite) Origin() Handler {
	return suite.origin
}

func (suite *PoloniexSuite) SetupSuite() {
	aliases := SymbolAliases{
		"REP": "REPV2",
	}
	suite.origin = NewBaseExchangeHandler(Poloniex{WorkerPool: query.NewMockWorkerPool()}, aliases)
}

func (suite *PoloniexSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Poloniex)
	suite.EqualValues("ETH_BTC", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
}

func (suite *PoloniexSuite) TestFailOnWrongInput() {
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

	suite.origin.ExchangeHandler.(Poloniex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Poloniex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Error during converting price to a number
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"ETH_BTC":{
					"last":"abc",
					"lowestAsk":"1.3",
					"highestBid":"1.0",
					"baseVolume":"10.1",
					"isFrozen": "0"
				}
			}
		`),
	}
	suite.origin.ExchangeHandler.(Poloniex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Frozen pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"ETH_BTC":{
					"last":"abc",
					"lowestAsk":"1.3",
					"highestBid":"1.0",
					"baseVolume":"10.1",
					"isFrozen": "1"
				}
			}
		`),
	}
	suite.origin.ExchangeHandler.(Poloniex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Unable to find pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"AAA_BBB":{
					"last":"1.1",
					"lowestAsk":"1.3",
					"highestBid":"1.0",
					"baseVolume":"10.1",
					"isFrozen": "0"
				}
			}
		`),
	}
	suite.origin.ExchangeHandler.(Poloniex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *PoloniexSuite) TestSuccessResponse() {
	pairBTCETH := Pair{Base: "BTC", Quote: "ETH"}
	pairBTCUSD := Pair{Base: "BTC", Quote: "USD"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			{
				"ETH_BTC":{
					"last":"1.1",
					"lowestAsk":"1.3",
					"highestBid":"1.0",
					"baseVolume":"10.1",
					"isFrozen": "0"
				},
				"USD_BTC":{
					"last":"2.1",
					"lowestAsk":"2.3",
					"highestBid":"2.0",
					"baseVolume":"20.1",
					"isFrozen": "0"
				},
				"EUR_BTC":{
					"last":"3.1",
					"lowestAsk":"3.3",
					"highestBid":"3.0",
					"baseVolume":"30.1",
					"isFrozen": "0"
				}
			}
		`),
	}
	suite.origin.ExchangeHandler.(Poloniex).Pool().(*query.MockWorkerPool).MockResp(resp)
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

func (suite *PoloniexSuite) TestRealAPICall() {
	testRealBatchAPICall(
		suite,
		NewBaseExchangeHandler(Poloniex{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		[]Pair{
			{Base: "ETH", Quote: "BTC"},
			{Base: "REP", Quote: "BTC"},
			{Base: "BTC", Quote: "USDT"},
		},
	)
}

func TestPoloniexSuite(t *testing.T) {
	suite.Run(t, new(PoloniexSuite))
}
