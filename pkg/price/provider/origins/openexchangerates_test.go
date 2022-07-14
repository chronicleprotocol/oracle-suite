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
type OpenExchangeRatesSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *OpenExchangeRatesSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *OpenExchangeRatesSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(OpenExchangeRates{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *OpenExchangeRatesSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *OpenExchangeRatesSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.origin.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "KRW", Quote: "USD"}
	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(OpenExchangeRates).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(OpenExchangeRates).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error getting quote
	resp = &query.HTTPResponse{
		Body: []byte(`{"rates":{}}`),
	}
	suite.origin.ExchangeHandler.(OpenExchangeRates).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error  getting quote
	resp = &query.HTTPResponse{
		Body: []byte(`{"rates":{"EUR":0}}`),
	}
	suite.origin.ExchangeHandler.(OpenExchangeRates).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *OpenExchangeRatesSuite) TestSuccessResponse() {
	pair := Pair{Base: "KRW", Quote: "USD"}
	resp := &query.HTTPResponse{
		Body: []byte(`{
		  "disclaimer": "Usage subject to terms: https://openexchangerates.org/terms",
		  "license": "https://openexchangerates.org/license",
		  "timestamp": 1621947600,
		  "base": "KRW",
		  "rates": {
			"USD": 0.000891
		  }
		}`),
	}
	suite.origin.ExchangeHandler.(OpenExchangeRates).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(0.000891, cr[0].Price.Price)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(2))
}

func (suite *OpenExchangeRatesSuite) TestRealAPICall() {
	origin := NewBaseExchangeHandler(OpenExchangeRates{
		WorkerPool: query.NewHTTPWorkerPool(1),
		APIKey:     "KEY_HERE",
	}, nil)
	testRealAPICall(suite, origin, "KRW", "USD")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestOpenExchangeRatesSuite(t *testing.T) {
	suite.Run(t, new(OpenExchangeRatesSuite))
}
