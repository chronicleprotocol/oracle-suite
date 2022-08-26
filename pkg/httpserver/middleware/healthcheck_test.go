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
