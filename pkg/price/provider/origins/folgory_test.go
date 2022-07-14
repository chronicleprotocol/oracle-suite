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
type FolgorySuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *FolgorySuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *FolgorySuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Folgory{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *FolgorySuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *FolgorySuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Folgory)
	suite.EqualValues("BTC/ETH", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.NotEqual("BTC/USDC", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *FolgorySuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	var cr []FetchResult

	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrInvalidResponseStatus, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Folgory).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(fmt.Errorf("bad response: %w", ourErr), cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Folgory).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[]`),
	}
	suite.origin.ExchangeHandler.(Folgory).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"symbol":"BTC/ETH","last":"abc"}]`),
	}
	suite.origin.ExchangeHandler.(Folgory).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"symbol":"BTC/ETH","last":"1","volume":"abc"}]`),
	}
	suite.origin.ExchangeHandler.(Folgory).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *FolgorySuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`[{"symbol":"BTC/ETH","last":"1","volume":"2"}]`),
	}
	suite.origin.ExchangeHandler.(Folgory).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Price.Price)
	suite.Equal(2.0, cr[0].Price.Volume24h)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *FolgorySuite) TestRealAPICall() {
	origin := NewBaseExchangeHandler(Folgory{WorkerPool: query.NewHTTPWorkerPool(1)}, nil)
	testRealAPICall(suite, origin, "ETH", "BTC")
	pairs := []Pair{
		{Base: "ETH", Quote: "USDF"},
		{Base: "BTC", Quote: "USDT"},
		{Base: "USDF", Quote: "DAI"},
		{Base: "BTC", Quote: "TUSD"},
	}
	testRealBatchAPICall(suite, origin, pairs)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFolgorySuite(t *testing.T) {
	suite.Run(t, new(FolgorySuite))
}
