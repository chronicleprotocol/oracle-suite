//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		path           string
		check          bool
		requestPath    string
		expectedStatus int
	}{
		{
			path:           "/health",
			check:          true,
			requestPath:    "/health",
			expectedStatus: http.StatusOK,
		},
		{
			path:           "/health",
			check:          false,
			requestPath:    "/health",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			path:           "/health",
			requestPath:    "/foo",
			expectedStatus: http.StatusBadRequest,
		},
	}
	for n, test := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			h := (&HealthCheck{
				Path: test.path,
				Check: func(r *http.Request) bool {
					return test.check
				},
			}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}))
			r := httptest.NewRequest("GET", test.requestPath, nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)
			assert.Equal(t, test.expectedStatus, rw.Code)
		})
	}
}
