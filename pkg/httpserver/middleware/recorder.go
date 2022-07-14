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

package middleware

import (
	"bytes"
	"io"
	"net/http"
)

// recorder implements the http.ResponseWriter interface. It passes all calls
// to the underlying ResponseWriter and records a copies of values for a later
// inspection.
type recorder struct {
	rw      http.ResponseWriter // rw is an underlying ResponseWriter.
	code    int                 // code is the HTTP status code
	headers http.Header         // headers is the list of HTTP headers
	body    *bytes.Buffer       // body is the HTTP response body
}

func newRecorder(rw http.ResponseWriter) *recorder {
	return &recorder{
		rw:      rw,
		headers: make(http.Header),
		body:    new(bytes.Buffer),
		code:    http.StatusOK,
	}
}

func (r *recorder) Header() http.Header {
	return r.rw.Header()
}

func (r *recorder) Write(buf []byte) (int, error) {
	r.body.Write(buf)
	return r.rw.Write(buf)
}

func (r *recorder) WriteHeader(code int) {
	r.code = code
	r.rw.WriteHeader(code)
}

func readRequest(r *http.Request) []byte {
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	return b
}

func readResponse(r *recorder) []byte {
	b, _ := io.ReadAll(r.body)
	return b
}
