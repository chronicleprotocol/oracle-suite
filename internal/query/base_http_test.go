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

package query

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/suite"
)

const serverResponse = `{"some":"another","status":1}`
const requiredHeaderKey = "X-Client-Id"
const requiredHeaderValue = "test-client"

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MakeRequestSuite struct {
	suite.Suite
	server *httptest.Server
}

func (suite *MakeRequestSuite) TearDownTest() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *MakeRequestSuite) TestMakingRequest() {
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.Write([]byte(serverResponse))
	}))

	assert.NotNil(suite.T(), suite.server)
	data, err := doMakeHTTPRequest(&HTTPRequest{URL: suite.server.URL})

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), []byte(serverResponse), data)
}

func (suite *MakeRequestSuite) TestMakingRequestToNotFound() {
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.WriteHeader(404)
	}))

	assert.NotNil(suite.T(), suite.server)
	data, err := doMakeHTTPRequest(&HTTPRequest{URL: suite.server.URL})

	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), data)
}

func (suite *MakeRequestSuite) TestMakingRequestWithHeaders() {
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.EqualValues(suite.T(), requiredHeaderValue, req.Header.Get(requiredHeaderKey))
		// Send response to be tested
		rw.Write([]byte(serverResponse))
	}))

	assert.NotNil(suite.T(), suite.server)
	headers := map[string]string{requiredHeaderKey: requiredHeaderValue}
	r := &HTTPRequest{
		URL:     suite.server.URL,
		Headers: headers,
	}
	data, err := doMakeHTTPRequest(r)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), []byte(serverResponse), data)
}

func (suite *MakeRequestSuite) TestMakingPOSTRequest() {
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(suite.T(), "POST", req.Method)
		// Send response to be tested
		rw.Write([]byte(serverResponse))
	}))

	assert.NotNil(suite.T(), suite.server)
	r := &HTTPRequest{
		URL:    suite.server.URL,
		Method: "POST",
	}
	data, err := doMakeHTTPRequest(r)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), []byte(serverResponse), data)
}

func (suite *MakeRequestSuite) TestMakeHTTPRequestWithRetryFails() {
	calls := 0
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.EqualValues(suite.T(), requiredHeaderValue, req.Header.Get(requiredHeaderKey))
		calls++
		// Send response to be tested.
		rw.WriteHeader(404)
	}))

	assert.NotNil(suite.T(), suite.server)
	headers := map[string]string{requiredHeaderKey: requiredHeaderValue}
	r := &HTTPRequest{
		URL:     suite.server.URL,
		Headers: headers,
		Retry:   3,
	}
	res := MakeHTTPRequest(r)

	assert.Error(suite.T(), res.Error)
	assert.EqualValues(suite.T(), []byte(nil), res.Body)
	assert.EqualValues(suite.T(), 3, calls)
}

func (suite *MakeRequestSuite) TestMakeHTTPRequestWithRetry() {
	calls := 0
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.EqualValues(suite.T(), requiredHeaderValue, req.Header.Get(requiredHeaderKey))
		calls++
		// Send response to be tested.
		// Successonly on 3rd call
		if calls < 3 {
			rw.WriteHeader(404)
		} else {
			rw.Write([]byte(serverResponse))
		}
	}))

	assert.NotNil(suite.T(), suite.server)
	headers := map[string]string{requiredHeaderKey: requiredHeaderValue}
	r := &HTTPRequest{
		URL:     suite.server.URL,
		Headers: headers,
		Retry:   3,
	}
	res := MakeHTTPRequest(r)

	assert.NoError(suite.T(), res.Error)
	assert.EqualValues(suite.T(), []byte(serverResponse), res.Body)
	assert.EqualValues(suite.T(), 3, calls)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMakeRequestSuite(t *testing.T) {
	suite.Run(t, new(MakeRequestSuite))
}
