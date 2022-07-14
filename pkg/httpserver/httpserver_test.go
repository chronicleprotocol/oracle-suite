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

package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_WithoutMiddlewares(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	srv := New(&http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("response"))
		}),
	})

	srv.ServeHTTP(rw, r)
	assert.Equal(t, "response", rw.Body.String())
}

func TestServer_WithMiddlewares(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	srv := New(&http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("response"))
		}),
	})
	srv.Use(MiddlewareFunc(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("before-"))
			handler.ServeHTTP(rw, r)
		})
	}))
	srv.Use(MiddlewareFunc(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(rw, r)
			rw.Write([]byte("-after"))
		})
	}))

	srv.ServeHTTP(rw, r)
	assert.Equal(t, "before-response-after", rw.Body.String())
}

func TestServer_UseMustPanicAfterServerHTTP(t *testing.T) {
	var panicVal interface{}
	defer func() { panicVal = recover() }()

	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	srv := New(&http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}),
	})
	srv.ServeHTTP(rw, r)
	srv.Use(MiddlewareFunc(func(handler http.Handler) http.Handler { return nil }))

	assert.NotNil(t, panicVal)
}
