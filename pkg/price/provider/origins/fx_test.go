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
type FxSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *FxSuite) Origin() Handler {
	return suite.origin
}

// Setup exchange
func (suite *FxSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Fx{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *FxSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *FxSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	// nil as response
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Fx).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Fx).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error convert price to number
	resp = &query.HTTPResponse{
		Body: []byte(`{"rates":{}}`),
	}
	suite.origin.ExchangeHandler.(Fx).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error convert price to number
	resp = &query.HTTPResponse{
		Body: []byte(`{"rates":{"ETH":"abcd"}}`),
	}
	suite.origin.ExchangeHandler.(Fx).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *FxSuite) TestSuccessResponse() {
	pair := Pair{Base: "A", Quote: "B"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"rates":{"B":1,"C":2},"base":"A"}`),
	}
	suite.origin.ExchangeHandler.(Fx).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Price.Price)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *FxSuite) TestRealAPICall() {
	fx := NewBaseExchangeHandler(Fx{
		WorkerPool: query.NewHTTPWorkerPool(1),
		APIKey:     "API_KEY", // TODO: find a way to pass API kEY ?
	}, nil)

	testRealAPICall(suite, fx, "USD", "EUR")
	testRealBatchAPICall(suite, fx, []Pair{
		{Base: "EUR", Quote: "USD"},
		{Base: "EUR", Quote: "PHP"},
		{Base: "EUR", Quote: "CAD"},
		{Base: "EUR", Quote: "SEK"},
		{Base: "USD", Quote: "SEK"},
		{Base: "SEK", Quote: "EUR"},
		{Base: "SEK", Quote: "USD"},
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFxSuite(t *testing.T) {
	suite.Run(t, new(FxSuite))
}
