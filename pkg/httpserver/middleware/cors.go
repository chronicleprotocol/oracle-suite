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
)

// CORS adds a basic support for CORS preflight requests.
type CORS struct {
	// Origin is a function that returns a value of
	// an Access-Control-Allow-Origin header. It cannot be nil.
	Origin func(r *http.Request) string
	// Headers is a function that returns a value of
	// an Access-Control-Allow-Headers header. It cannot be nil.
	Headers func(r *http.Request) string
	// Methods is a function that returns a value of
	// an Access-Control-Allow-Methods. It cannot be nil.
	Methods func(r *http.Request) string
}

// Handle implements the httpserver.Middleware interface.
func (c *CORS) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		headers := rw.Header()
		headers.Set("Access-Control-Allow-Origin", c.Origin(r))
		switch r.Method {
		case "OPTIONS":
			headers.Set("Access-Control-Allow-Headers", c.Headers(r))
			headers.Set("Access-Control-Allow-Methods", c.Methods(r))
			headers.Set("Access-Control-Max-Age", "86400")
			rw.WriteHeader(http.StatusNoContent)
		default:
			next.ServeHTTP(rw, r)
		}
	})
}
