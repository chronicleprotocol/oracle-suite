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

package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

func TestOAuth_Sign(t *testing.T) {
	tests := []struct {
		request       *http.Request
		authorization string
	}{
		{
			request:       errutil.Must(http.NewRequest(http.MethodGet, "https://example.com/", nil)),
			authorization: `OAuth oauth_consumer_key="ck", oauth_nonce="bm9uY2U%3D", oauth_signature="EDYNfT5ocp4nmKoUmjhG2ob4110%3D", oauth_signature_method="HMAC-SHA1", oauth_timestamp="1000", oauth_token="at", oauth_version="1.0"`,
		},
		{
			request:       errutil.Must(http.NewRequest(http.MethodPost, "https://example.com/", nil)),
			authorization: `OAuth oauth_consumer_key="ck", oauth_nonce="bm9uY2U%3D", oauth_signature="%2F17%2Fd5g0ia6npSALcSTcEjylW8Q%3D", oauth_signature_method="HMAC-SHA1", oauth_timestamp="1000", oauth_token="at", oauth_version="1.0"`,
		},
		{
			request:       errutil.Must(http.NewRequest(http.MethodGet, "https://example.com/path?param1=foo&param2=bar", nil)),
			authorization: `OAuth oauth_consumer_key="ck", oauth_nonce="bm9uY2U%3D", oauth_signature="jNJRcZl9XBxlwhxN9K%2BaaWK%2B6tY%3D", oauth_signature_method="HMAC-SHA1", oauth_timestamp="1000", oauth_token="at", oauth_version="1.0"`,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			oAuth := &OAuth{
				ConsumerKey:    "ck",
				ConsumerSecret: "cs",
				AccessToken:    "at",
				AccessSecret:   "as",
				nonce:          []byte("nonce"),
				time:           time.Unix(1000, 0),
			}
			require.NoError(t, oAuth.Sign(tt.request))
			assert.Equal(t, tt.authorization, tt.request.Header.Get("Authorization"))
		})
	}
}

func TestOAuth_RandomNonce(t *testing.T) {
	oAuth := &OAuth{
		ConsumerKey:    "ck",
		ConsumerSecret: "cs",
		AccessToken:    "at",
		AccessSecret:   "as",
		time:           time.Unix(1000, 0),
	}

	request1 := errutil.Must(http.NewRequest(http.MethodGet, "https://example.com/", nil))
	request2 := errutil.Must(http.NewRequest(http.MethodGet, "https://example.com/", nil))

	require.NoError(t, oAuth.Sign(request1))
	require.NoError(t, oAuth.Sign(request2))

	assert.NotEqual(t, request1.Header.Get("Authorization"), request2.Header.Get("Authorization"))
}
