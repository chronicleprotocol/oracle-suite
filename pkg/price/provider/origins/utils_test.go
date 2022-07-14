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
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Suite interface {
	suite.TestingSuite

	Assert() *assert.Assertions
	Origin() Handler
}

func testRealAPICall(suite Suite, origin *BaseExchangeHandler, base, quote string) {
	testRealBatchAPICall(suite, origin, []Pair{{Base: base, Quote: quote}})
}

func testRealBatchAPICall(suite Suite, origin *BaseExchangeHandler, pairs []Pair) {
	if os.Getenv("GOFER_TEST_API_CALLS") == "" {
		suite.T().SkipNow()
	}

	suite.Assert().IsType(suite.Origin(), origin)

	crs := origin.Fetch(pairs)

	for _, cr := range crs {
		suite.Assert().NoErrorf(cr.Error, "%q", cr.Price.Pair)
		suite.Assert().Greater(cr.Price.Price, float64(0))
	}
}
