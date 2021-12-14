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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/callback"
)

func TestLogger_InfoLevel(t *testing.T) {
	var recordedLogMsgs []string
	var recordedLogFields []log.Fields
	l := callback.New(log.Info, func(level log.Level, fields log.Fields, msg string) {
		if level != log.Info {
			return
		}
		recordedLogMsgs = append(recordedLogMsgs, msg)
		recordedLogFields = append(recordedLogFields, fields)
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h := (&Logger{Log: l}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	h.ServeHTTP(w, r)

	require.Len(t, recordedLogMsgs, 1)
	assert.Equal(t, httpRequestLog, recordedLogMsgs[0])
	assert.Equal(t, "GET", recordedLogFields[0]["method"])
	assert.Equal(t, "/", recordedLogFields[0]["url"])
	assert.NotEmpty(t, recordedLogFields[0]["duration"])
	assert.NotEmpty(t, recordedLogFields[0]["remoteAddr"])
}

func TestLogger_DebugLevel(t *testing.T) {
	var recordedLogMsgs []string
	var recordedLogFields []log.Fields
	l := callback.New(log.Debug, func(level log.Level, fields log.Fields, msg string) {
		if level != log.Debug {
			return
		}
		recordedLogMsgs = append(recordedLogMsgs, msg)
		recordedLogFields = append(recordedLogFields, fields)
	})

	r := httptest.NewRequest("GET", "/", strings.NewReader("request"))
	w := httptest.NewRecorder()
	h := (&Logger{Log: l}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("response"))
		writer.WriteHeader(http.StatusNotFound)
	}))
	h.ServeHTTP(w, r)

	require.Len(t, recordedLogMsgs, 1)
	assert.Equal(t, httpRequestLog, recordedLogMsgs[0])
	assert.Equal(t, "GET", recordedLogFields[0]["method"])
	assert.Equal(t, "/", recordedLogFields[0]["url"])
	assert.Equal(t, http.StatusNotFound, recordedLogFields[0]["status"])
	assert.Equal(t, "request", recordedLogFields[0]["request"])
	assert.Equal(t, "response", recordedLogFields[0]["response"])
	assert.NotEmpty(t, recordedLogFields[0]["duration"])
	assert.NotEmpty(t, recordedLogFields[0]["remoteAddr"])
}
