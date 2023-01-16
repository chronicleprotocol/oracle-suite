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
	"testing"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ISharesSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *ISharesSuite) Origin() Handler {
	return suite.origin
}

// Setup exchange
func (suite *ISharesSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(IShares{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *ISharesSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *ISharesSuite) TestSuccessResponse() {
	pair := Pair{Base: "IBTA", Quote: "USD"}
	resp := &query.HTTPResponse{
		Body: []byte(`<html><body><li class="navAmount " data-col="fundHeader.fundNav.navAmount" data-path="">
<span class="header-nav-label navAmount">
NAV as of 09/Jan/2023
</span>
<span class="header-nav-data">
USD 5.43
</span>
<span class="header-info-bubble">
</span>
<br>
<span class="fiftyTwoWeekData">
52 WK: 5.11 - 5.37
</span>
</li></body></html>`),
	}
	suite.origin.ExchangeHandler.(IShares).Pool().(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(5.43, cr[0].Price.Price)
}

// // In order for 'go test' to run this suite, we need to create
// // a normal test function and pass our suite to suite.Run
func TestISharesSuite(t *testing.T) {
	suite.Run(t, new(ISharesSuite))
}
