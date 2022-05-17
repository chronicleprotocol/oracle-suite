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
	"fmt"
	"io"
	"net/http"
)

// Transport implements the http.RoundTripper interface. It creates a virtual
// host with RPC Splitter.
type Transport struct {
	transport http.RoundTripper
	server    http.Handler
	vhost     string
}

// NewTransport returns a new instance of Transport.
func NewTransport(vhost string, transport http.RoundTripper, opts ...Option) (*Transport, error) {
	if transport == nil {
		transport = http.DefaultTransport
	}
	rpcServer, err := NewServer(opts...)
	if err != nil {
		return nil, err
	}
	return &Transport{
		transport: transport,
		vhost:     vhost,
		server:    rpcServer,
	}, nil
}

// RoundTrip implements the http.RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.isVirtualHost(req) {
		return t.transport.RoundTrip(req)
	}
	rec := newRecorder()
	t.server.ServeHTTP(rec, req)
	return t.buildResponse(rec), nil
}

func (t *Transport) isVirtualHost(req *http.Request) bool {
	return req.Host == t.vhost
}

func (t *Transport) buildResponse(res *recorder) *http.Response {
	return &http.Response{
		Status:        fmt.Sprintf("%d %s", res.code, http.StatusText(res.code)),
		StatusCode:    res.code,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(res.body.Len()),
		Header:        res.headers,
		Body:          io.NopCloser(res.body),
	}
}

// recorder is an implementation of http.ResponseWriter that
// records its mutations for later inspection.
type recorder struct {
	code    int           // code is the HTTP status code
	headers http.Header   // headers is the list of HTTP headers
	body    *bytes.Buffer // body is the HTTP response body
}

func newRecorder() *recorder {
	return &recorder{
		headers: make(http.Header),
		body:    new(bytes.Buffer),
		code:    http.StatusOK,
	}
}

func (r *recorder) Header() http.Header {
	return r.headers
}

func (r *recorder) Write(buf []byte) (int, error) {
	return r.body.Write(buf)
}

func (r *recorder) WriteHeader(code int) {
	r.code = code
}
