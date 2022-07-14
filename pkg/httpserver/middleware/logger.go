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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const httpRequestLog = "HTTP request"

// Logger prints logs for each request. If the log level is set to debug, it
// will print the contents of requests and responses.
type Logger struct {
	// Log is an instance of a log.Logger. It cannot be nil, otherwise code will panic.
	Log log.Logger
}

// Handle implements the httpserver.Middleware interface.
func (l *Logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		t := time.Now()
		e := l.Log
		if l.Log.Level() >= log.Debug {
			rw = newRecorder(rw)
			e = e.WithField("request", string(readRequest(r)))
		}
		defer func() {
			e = e.WithFields(log.Fields{
				"remoteAddr": r.RemoteAddr,
				"duration":   time.Since(t),
				"method":     r.Method,
				"url":        r.URL.String(),
			})
			if l.Log.Level() >= log.Debug {
				e = e.WithFields(log.Fields{
					"response": string(readResponse(rw.(*recorder))),
					"status":   rw.(*recorder).code,
				})
				e.Debug(httpRequestLog)
			} else {
				e.Info(httpRequestLog)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}
