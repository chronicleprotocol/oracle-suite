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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS_Options(t *testing.T) {
	c := &CORS{
		Origin:  func(r *http.Request) string { return "origin" },
		Headers: func(r *http.Request) string { return "header" },
		Methods: func(r *http.Request) string { return "GET" },
	}
	h := c.Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	r := httptest.NewRequest("OPTIONS", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, "origin", rw.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "header", rw.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET", rw.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "86400", rw.Header().Get("Access-Control-Max-Age"))
}

func TestCORS_Get(t *testing.T) {
	c := &CORS{
		Origin:  func(r *http.Request) string { return "origin" },
		Headers: func(r *http.Request) string { return "header" },
		Methods: func(r *http.Request) string { return "GET" },
	}
	h := c.Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "origin", rw.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rw.Header().Get("Access-Control-Allow-Headers"))
	assert.Empty(t, rw.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(t, rw.Header().Get("Access-Control-Allow-Max-Age"))
}
