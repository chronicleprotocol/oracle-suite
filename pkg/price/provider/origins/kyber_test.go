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
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

type KyberSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *KyberSuite) Origin() Handler {
	return suite.origin
}

func (suite *KyberSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Kyber{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *KyberSuite) TearDownTest() {
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KyberSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "WBTC", Quote: "ETH"}
	var cr []FetchResult

	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrInvalidResponseStatus, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.ExchangeHandler.(Kyber).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(fmt.Errorf("bad response: %w", ourErr), cr[0].Error)

	for n, r := range [][]byte{
		[]byte(""),
		[]byte("{}"),
		[]byte("[]"),
		[]byte(`[ {
		"timestamp": 1600331875531,
		"token_symbol": "WBTC",
		"token_name": "Wrapped BTC",
		"token_address": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
		"token_decimal": 8,
		"rate_eth_now": 30.11825982131223,
		"change_eth_24h": -2.17,
		"rate_usd_now": 11375.32395734396,
		"change_usd_24h": 2.27
		}]`),
		[]byte(`{"ETH_WBTC": {
		"timestamp": "",
		"token_symbol": "WBTC",
		"token_name": "Wrapped BTC",
		"token_address": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
		"token_decimal": 8,
		"rate_eth_now": 30.11825982131223,
		"change_eth_24h": -2.17,
		"rate_usd_now": 11375.32395734396,
		"change_usd_24h": 2.27
		}}`),
		[]byte(`{"ETH_WBTC": {
		"timestamp": 1600331875531,
		"token_symbol": 0,
		"token_name": "Wrapped BTC",
		"token_address": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
		"token_decimal": 8,
		"rate_eth_now": 30.11825982131223,
		"change_eth_24h": -2.17,
		"rate_usd_now": 11375.32395734396,
		"change_usd_24h": 2.27
		}}`),
		[]byte(`{"ETH_WBTC": {
		"timestamp": 1600331875531,
		"token_symbol": "",
		"token_name": "Wrapped BTC",
		"token_address": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
		"token_decimal": 8,
		"rate_eth_now": 30.11825982131223,
		"change_eth_24h": -2.17,
		"rate_usd_now": 11375.32395734396,
		"change_usd_24h": 2.27
		}}`),
		[]byte(`{"ETH_WBTC": {
		"timestamp": 1600331875531,
		"token_symbol": "WBTC",
		"token_name": 0,
		"token_address": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
		"token_decimal": 8,
		"rate_eth_now": 30.11825982131223,
		"change_eth_24h": -2.17,
		"rate_usd_now": 11375.32395734396,
		"change_usd_24h": 2.27
		}}`),
		[]byte(`{"ETH_WBTC": {
		"timestamp": 1600331875531,
		"token_symbol": "WBTC",
		"token_name": "Wrapped BTC",
		"token_address": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
		"token_decimal": 1.1,
		"rate_eth_now": 30.11825982131223,
		"change_eth_24h": -2.17,
		"rate_usd_now": 11375.32395734396,
		"change_usd_24h": 2.27
		}}`),
		[]byte(`{"ETH_WBTC": {
		"timestamp": 1600331875531,
		"token_symbol": "WBTC",
		"token_name": "Wrapped BTC",
		"token_address": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
		"token_decimal": 8,
		"rate_eth_now": "",
		"change_eth_24h": -2.17,
		"rate_usd_now": 11375.32395734396,
		"change_usd_24h": 2.27
		}}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.origin.ExchangeHandler.(Kyber).Pool().(*query.MockWorkerPool).MockResp(resp)
			cr = suite.origin.Fetch([]Pair{pair})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *KyberSuite) TestSuccessResponse() {
	pair := Pair{Base: "WBTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"ETH_WBTC": {
			"timestamp": 1600331875531,
			"token_symbol": "WBTC",
			"token_name": "Wrapped BTC",
			"token_address": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
			"token_decimal": 8,
			"rate_eth_now": 30.11825982131223,
			"change_eth_24h": -2.17,
			"rate_usd_now": 11375.32395734396,
			"change_usd_24h": 2.27
			}}
		`),
	}

	suite.origin.ExchangeHandler.(Kyber).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(30.11825982131223, cr[0].Price.Price)
	suite.Equal(time.Unix(1600331875, 0).Unix(), cr[0].Price.Timestamp.Unix())
}

func (suite *KyberSuite) TestRealAPICall() {
	origin := NewBaseExchangeHandler(Kyber{WorkerPool: query.NewHTTPWorkerPool(1)}, nil)

	testRealAPICall(suite, origin, "WBTC", "ETH")
	pairs := []Pair{
		{Base: "WBTC", Quote: "ETH"},
		{Base: "WETH", Quote: "ETH"},
		{Base: "DAI", Quote: "ETH"},
	}
	testRealBatchAPICall(suite, origin, pairs)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKyberSuiteSuite(t *testing.T) {
	suite.Run(t, new(KyberSuite))
}
