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
type CryptoCompareSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *CryptoCompareSuite) Origin() Handler {
	return suite.origin
}

// Setup exchange
func (suite *CryptoCompareSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(CryptoCompare{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *CryptoCompareSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *CryptoCompareSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(CryptoCompare).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	for n, r := range [][]byte{
		// invalid response
		[]byte(``),
		// invalid response
		[]byte(`{}`),
		// invalid quote
		[]byte(`{"NON":1.1}`),
		// invalid price
		[]byte(`{"ETH":"1.1"}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.origin.ExchangeHandler.(CryptoCompare).Pool().(*query.MockWorkerPool).MockResp(resp)
			cr = suite.origin.Fetch([]Pair{pair})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *CryptoCompareSuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"RAW":{"BTC":{"ETH":{
		"FROMSYMBOL": "BTC",
		"TOSYMBOL": "ETH",
		"PRICE": 0.04687,
		"VOLUME24HOUR": 0,
		"LASTUPDATE": 1599982420
		}}}}`),
	}
	suite.origin.ExchangeHandler.(CryptoCompare).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(0.04687, cr[0].Price.Price)
	suite.Equal(cr[0].Price.Timestamp.Unix(), int64(1599982420))
}

func (suite *CryptoCompareSuite) TestRealAPICall() {
	origin := NewBaseExchangeHandler(CryptoCompare{WorkerPool: query.NewHTTPWorkerPool(1)}, nil)

	testRealAPICall(suite, origin, "ETH", "BTC")
	var pairs []Pair
	for _, s := range []string{"BTC", "ETH", "MKR", "POLY"} {
		pairs = append(pairs, Pair{Base: s, Quote: "USD"})
	}
	testRealBatchAPICall(suite, origin, pairs)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCryptoCompareSuite(t *testing.T) {
	suite.Run(t, new(CryptoCompareSuite))
}
