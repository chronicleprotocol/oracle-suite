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

type BittrexSuite struct {
	suite.Suite
	origin *BaseExchangeHandler
}

func (suite *BittrexSuite) Origin() Handler {
	return suite.origin
}

func (suite *BittrexSuite) SetupSuite() {
	aliases := SymbolAliases{
		"REP": "REPV2",
	}
	suite.origin = NewBaseExchangeHandler(Bittrex{WorkerPool: query.NewMockWorkerPool()}, aliases)
}

func (suite *BittrexSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Bittrex)
	suite.EqualValues("ETH-BTC", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
}

func (suite *BittrexSuite) TestFailOnWrongInput() {
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

	suite.origin.ExchangeHandler.(Bittrex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.ExchangeHandler.(Bittrex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Price as string
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
			   "success":true,
			   "message":"",
			   "result": {
				 "Last":"1.1",
				 "Bid":1.0,
				 "Ask":1.3
			  }
			}
		`),
	}
	suite.origin.ExchangeHandler.(Bittrex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Unable to find pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
			   "success":true,
			   "message":"",
			   "result": {
				 "Last":"1.1",
				 "Bid":1.0,
				 "Ask":1.3
			  }
			}
		`),
	}
	suite.origin.ExchangeHandler.(Bittrex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *BittrexSuite) TestSuccessResponse() {
	pairBTCETH := Pair{Base: "BTC", Quote: "ETH"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			{
			   "success":true,
			   "message":"",
			   "result": {
				 "Last":1.1,
				 "Bid":1.0,
				 "Ask":1.3
			  }
			}
		`),
	}
	suite.origin.ExchangeHandler.(Bittrex).Pool().(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairBTCETH})

	suite.Len(fr, 1)

	// BTC/ETH
	suite.NoError(fr[0].Error)
	suite.Equal(pairBTCETH, fr[0].Price.Pair)
	suite.Equal(1.1, fr[0].Price.Price)
	suite.Equal(1.0, fr[0].Price.Bid)
	suite.Equal(1.3, fr[0].Price.Ask)
	suite.Greater(fr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *BittrexSuite) TestRealAPICall() {
	aliases := SymbolAliases{
		"REP": "REPV2",
	}
	testRealBatchAPICall(
		suite,
		NewBaseExchangeHandler(Bittrex{WorkerPool: query.NewHTTPWorkerPool(1)}, aliases),
		[]Pair{
			{Base: "MANA", Quote: "BTC"},
			{Base: "BAT", Quote: "BTC"},
			{Base: "BTC", Quote: "USD"},
		},
	)
}

func TestBittrexSuite(t *testing.T) {
	suite.Run(t, new(BittrexSuite))
}
