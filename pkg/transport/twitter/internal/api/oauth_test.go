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
