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

package rpcsplitter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rpcReq struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

type rpcRes struct {
	ID      int         `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type mockClient struct {
	t *testing.T

	currCall int            // current call idx, increases every time Call is called
	calls    []expectedCall // list of expected calls
}

type expectedCall struct {
	result func() interface{}
	method string
	params []interface{}
}

// mockCall adds expected call. If a result implements an error interface,
// then it will be returned as an error.
func (c *mockClient) mockCall(result interface{}, method string, params ...interface{}) {
	c.calls = append(c.calls, expectedCall{
		result: func() interface{} { return result },
		method: method,
		params: params,
	})
}

// mockSlowCall works just like mockCall but adds a delay to the response.
func (c *mockClient) mockSlowCall(delay time.Duration, result interface{}, method string, params ...interface{}) {
	c.calls = append(c.calls, expectedCall{
		result: func() interface{} { time.Sleep(delay); return result },
		method: method,
		params: params,
	})
}

func (c *mockClient) CallContext(ctx context.Context, result interface{}, method string, params ...interface{}) error {
	if c.currCall >= len(c.calls) {
		require.Fail(c.t, "unexpected call")
	}
	defer func() { c.currCall++ }()

	// Check if current call meets expectations.
	call := c.calls[c.currCall]
	assert.Equal(c.t, call.method, method)
	assert.True(c.t, compare(call.params, params))

	// Wait for the result:
	var callResult interface{}
	callResultCh := make(chan interface{})
	go func() { callResultCh <- call.result() }()
	select {
	case callResult = <-callResultCh:
	case <-ctx.Done():
		callResult = errors.New("context cancelled")
	}

	// Error results are treated differently, as described in mockCall.
	if err, ok := callResult.(error); ok {
		return err
	}

	// Message is marshalled and unmarshalled to verify, if marshalling is
	// implemented correctly.
	return json.Unmarshal(jsonMarshal(c.t, callResult), result)
}

type handlerTester struct {
	t *testing.T

	clients   []caller
	options   []Option
	expResult interface{}
	expMethod string
	expParams []interface{}
	expErrors []string
}

func prepareHandlerTest(t *testing.T, clients int, method string, params ...interface{}) *handlerTester {
	var callers []caller
	for i := 0; i < clients; i++ {
		callers = append(callers, &mockClient{t: t})
	}
	return &handlerTester{t: t, clients: callers, expMethod: method, expParams: params}
}

// mockClientCall mocks call on n client.
func (t *handlerTester) mockClientCall(n int, response interface{}, method string, params ...interface{}) *handlerTester {
	t.clients[n].(*mockClient).mockCall(response, method, params...)
	return t
}

// mockClientSlowCall mocks call with a delay on n client.
func (t *handlerTester) mockClientSlowCall(delay time.Duration, n int, response interface{}, method string, params ...interface{}) *handlerTester {
	t.clients[n].(*mockClient).mockSlowCall(delay, response, method, params...)
	return t
}

// setRequirements is an equivalent of WithRequirements option.
func (t *handlerTester) setOptions(opts ...Option) *handlerTester {
	t.options = append(t.options, opts...)
	return t
}

// expectedResult sets expected result.
func (t *handlerTester) expectedResult(res interface{}) *handlerTester {
	t.expResult = res
	return t
}

// expectedError adds an error as an expectation. If msg is a non-empty string,
// a returned error must contain msg. If msg is empty, any error will match.
func (t *handlerTester) expectedError(msg string) *handlerTester {
	t.expErrors = append(t.expErrors, msg)
	return t
}

func (t *handlerTester) test() {
	// Prepare server.
	callers := map[string]caller{}
	for n, c := range t.clients {
		callers[fmt.Sprintf("%d", n)] = c
	}
	h, err := NewServer(append([]Option{withCallers(callers)}, t.options...)...)
	require.NoError(t.t, err)

	// Prepare request.
	id := rand.Int()
	msg := jsonMarshal(t.t, rpcReq{
		ID:      id,
		JSONRPC: "2.0",
		Method:  t.expMethod,
		Params:  t.expParams,
	})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(msg))
	r.Header.Set("Content-Type", "application/json")

	// Do request.
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	// Unmarshall response.
	res := &rpcRes{}
	jsonUnmarshal(t.t, rw.Body.Bytes(), res)

	// Verify response.
	assert.Equal(t.t, id, res.ID)
	assert.Equal(t.t, "2.0", res.JSONRPC)
	if len(t.expErrors) > 0 {
		for _, e := range t.expErrors {
			if e == "" {
				assert.NotEmpty(t.t, res.Error.Message)
			} else {
				assert.Contains(t.t, res.Error.Message, e)
			}
		}
	} else {
		assert.Equal(t.t, 0, res.Error.Code)
		assert.Empty(t.t, res.Error.Message)
		assert.JSONEq(t.t, string(jsonMarshal(t.t, t.expResult)), string(jsonMarshal(t.t, res.Result)))
	}
}

func jsonMarshal(t *testing.T, v interface{}) []byte {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func jsonUnmarshal(t *testing.T, b []byte, v interface{}) interface{} {
	require.NoError(t, json.Unmarshal(b, v))
	return v
}
