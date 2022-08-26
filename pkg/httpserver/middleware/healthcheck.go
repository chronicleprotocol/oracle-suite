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
	"strings"
)

// HealthCheck is a middleware that checks the health of the service, it may
// be used with Kubernetes' liveliness probe.
//
// It returns a 200 response if the service is healthy, otherwise it returns a
// 503 response.
type HealthCheck struct {
	// Path is the path where the health check will be available.
	Path string
	// Check is a function that will be called to check the health of the service.
	Check func(r *http.Request) bool
}

// Handle implements the httpserver.Middleware interface.
func (c *HealthCheck) Handle(next http.Handler) http.Handler {
	path := "/" + strings.Trim(c.Path, "/")
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if path == strings.TrimRight(r.URL.Path, "/") {
			if c.Check != nil && !c.Check(r) {
				rw.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			rw.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(rw, r)
	})
}
