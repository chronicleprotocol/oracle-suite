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

import "net/http"

// Recover recovers from panics that occurred during invoking handler's
// ServeHTTP method.
type Recover struct {
	// Recover is an optional function that will be invoked during panicking.
	// It can be nil.
	Recover func(err interface{})
}

// Handle implements the httpserver.Middleware interface.
func (c *Recover) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil && c.Recover != nil {
				c.Recover(r)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}
