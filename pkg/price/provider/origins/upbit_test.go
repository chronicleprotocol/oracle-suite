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
type UpbitSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *UpbitSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *UpbitSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Upbit{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *UpbitSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *UpbitSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Upbit)
	suite.EqualValues("ETH-BTC", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.NotEqual("USDC-BTC", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *UpbitSuite) TestFailOnWrongInput() {
	var cr []FetchResult
	pair := Pair{Base: "BTC", Quote: "ETH"}

	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrInvalidResponseStatus, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Upbit).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(fmt.Errorf("bad response: %w", ourErr), cr[0].Error)

	for n, r := range [][]byte{
		[]byte(``),
		[]byte(`{}`),
		[]byte(`[]`),
		[]byte(`{"success":true}`),
		[]byte(`{"success":true,"result":{}}`),
		[]byte(`{"success":true,"result":[]}`),
		[]byte(`{"success":true,"result":[{"name":"SOME/ANOTHER"}]}`),
		[]byte(`{"success":true,"result":[{"name":"BTC/ETH","last":"1"}]}`),
		[]byte(`{"success":true,"result":[{"name":"BTC/ETH","last":1,"ask":"2"}]}`),
		[]byte(`{"success":true,"result":[{"name":"BTC/ETH","last":1,"ask":2,"bid":"3"}]}`),
		[]byte(`{"success":true,"result":[{"name":"BTC/ETH","last":1,"ask":2,"bid":3,"quoteVolume24h":"4"}]}`),
		[]byte(`[{
					"market": "BTC-ETH",
					"trade_date": "20200917",
					"trade_time": "100909",
					"trade_timestamp": 1600337349000,
					"opening_price": 0.0334175,
					"high_price": 0.03539946,
					"low_price": 0.0334175,
					"trade_price": 0.03527794,
					"prev_closing_price": 0.0333,
					"change": "RISE",
					"change_price": 0.00197794,
					"change_rate": 0.0593975976,
					"signed_change_price": 0.00197794,
					"signed_change_rate": 0.0593975976,
					"trade_volume": 0.47239852,
					"acc_trade_price": 1.16721813,
					"acc_trade_price_24h": 1.5688811943216596,
					"acc_trade_volume": 33.40917363,
					"acc_trade_volume_24h": 45.24091194,
					"highest_52_week_price": 0.0414002,
					"highest_52_week_date": "2020-09-02",
					"lowest_52_week_price": 0.0170001,
					"lowest_52_week_date": "2020-01-08",
					"timestamp": 2000
				}]`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.origin.ExchangeHandler.(Upbit).Pool().(*query.MockWorkerPool).MockResp(resp)
			cr = suite.origin.Fetch([]Pair{pair})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *UpbitSuite) TestSuccessResponse() {
	pair := Pair{Base: "ETH", Quote: "BTC"}
	resp := &query.HTTPResponse{
		Body: []byte(`[{
						"market": "BTC-ETH",
						"trade_date": "20200917",
						"trade_time": "100909",
						"trade_timestamp": 1600337349000,
						"opening_price": 0.0334175,
						"high_price": 0.03539946,
						"low_price": 0.0334175,
						"trade_price": 0.03527794,
						"prev_closing_price": 0.0333,
						"change": "RISE",
						"change_price": 0.00197794,
						"change_rate": 0.0593975976,
						"signed_change_price": 0.00197794,
						"signed_change_rate": 0.0593975976,
						"trade_volume": 0.47239852,
						"acc_trade_price": 1.16721813,
						"acc_trade_price_24h": 1.5688811943216596,
						"acc_trade_volume": 33.40917363,
						"acc_trade_volume_24h": 45.24091194,
						"highest_52_week_price": 0.0414002,
						"highest_52_week_date": "2020-09-02",
						"lowest_52_week_price": 0.0170001,
						"lowest_52_week_date": "2020-01-08",
						"timestamp": 2000
					}]`),
	}
	suite.origin.ExchangeHandler.(Upbit).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(0.03527794, cr[0].Price.Price)
	suite.Equal(45.24091194, cr[0].Price.Volume24h)
	suite.Equal(cr[0].Price.Timestamp.Unix(), int64(2))
}

func (suite *UpbitSuite) TestRealAPICall() {
	testRealAPICall(
		suite,
		NewBaseExchangeHandler(Upbit{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		"ETH", "BTC",
	)
	pairs := []Pair{
		{Base: "KNC", Quote: "KRW"},
		{Base: "MANA", Quote: "KRW"},
		{Base: "OMG", Quote: "KRW"},
		{Base: "REP", Quote: "KRW"},
		{Base: "SNT", Quote: "KRW"},
		{Base: "ZRX", Quote: "KRW"},
	}
	testRealBatchAPICall(
		suite,
		NewBaseExchangeHandler(Upbit{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		pairs,
	)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUpbitSuite(t *testing.T) {
	suite.Run(t, new(UpbitSuite))
}
