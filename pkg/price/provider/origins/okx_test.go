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

type OkxSuite struct {
	suite.Suite
	origin *BaseExchangeHandler
}

func (suite *OkxSuite) Origin() Handler {
	return suite.origin
}

func (suite *OkxSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Okx{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *OkxSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Okx)
	suite.EqualValues("BTC-ETH-SWAP", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("BTC-USD-SWAP", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *OkxSuite) TestFailOnWrongInput() {
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

	suite.origin.ExchangeHandler.(Okx).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Okx).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Error during converting price to a number
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"code":"0",
				"msg":"",
				"data":[
				 {
						"instType":"SWAP",
						"instId":"BTC-ETH-SWAP",
						"last":"abcd",
						"askPx":"9999.99",
						"bidPx":"8888.88",
						"vol24h":"2222",
						"ts":"1597026383085"
					}
				]
			}
		`),
	}
	suite.origin.ExchangeHandler.(Okx).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *OkxSuite) TestSuccessResponse() {
	pairBTCUSD := Pair{Base: "BTC", Quote: "USD"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			{
				"code":"0",
				"msg":"",
				"data":[
				 {
						"instType":"SWAP",
						"instId":"BTC-USD-SWAP",
						"last":"9999.99",
						"askPx":"9999.99",
						"bidPx":"8888.88",
						"vol24h":"2222",
						"ts":"1597026383085"
					}
				]
			}
		`),
	}
	suite.origin.ExchangeHandler.(Okx).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairBTCUSD})

	suite.Len(fr, 1)

	// BTC/USD
	suite.NoError(fr[0].Error)
	suite.Equal(pairBTCUSD, fr[0].Price.Pair)
	suite.Equal(9999.99, fr[0].Price.Price)
	suite.Equal(8888.88, fr[0].Price.Bid)
	suite.Equal(9999.99, fr[0].Price.Ask)
	suite.Equal(float64(2222), fr[0].Price.Volume24h)
	suite.Greater(fr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *OkxSuite) TestRealAPICall() {
	testRealBatchAPICall(
		suite,
		NewBaseExchangeHandler(Okx{WorkerPool: query.NewHTTPWorkerPool(1)}, nil),
		[]Pair{
			{Base: "BTC", Quote: "USD"},
		},
	)
}

func TestOkxSuite(t *testing.T) {
	suite.Run(t, new(OkxSuite))
}
