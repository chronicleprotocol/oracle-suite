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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

// OAuth provides minimalistic OAuth 1.0a implementation that have only the
// necessary features for Twitter API.
type OAuth struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string

	// The nonce and time fields are used in tests only.
	nonce []byte
	time  time.Time
}

// Sign signs the given request with OAuth 1.0a.
func (o *OAuth) Sign(req *http.Request) error {
	// Prepares OAuth parameters as described in RFC5849 section 3.1.
	nonce, err := o.randomNonce()
	if err != nil {
		return err
	}
	params := map[string]string{
		"oauth_version":          "1.0",
		"oauth_consumer_key":     o.ConsumerKey,
		"oauth_token":            o.AccessToken,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        strconv.FormatInt(o.timestamp(), 10),
		"oauth_nonce":            base64.StdEncoding.EncodeToString(nonce),
	}

	// Collects all parameters as described in RFC5849 section 3.4.1.3.1.
	signatureParams := maputil.Copy(params)
	for key, value := range req.URL.Query() {
		// It should be safe to ignore duplicate query parameters.
		signatureParams[key] = value[0]
	}

	// Generate HMAC-SHA1 signature as described in RFC5849 section 3.4.2.
	mac := hmac.New(sha1.New, []byte(fmt.Sprintf("%s&%s", o.ConsumerSecret, o.AccessSecret)))
	mac.Write([]byte(signatureBase(req, signatureParams)))
	params["oauth_signature"] = base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Set the "Authorization" header as described in RFC5849 section 3.5.1.
	req.Header.Set(
		"Authorization",
		fmt.Sprintf("OAuth %s", formatParams(params, `%s="%s"`, ", ")),
	)

	return nil
}

// randomNonce generates random nonce as described in RFC5849 section 3.3.
func (o *OAuth) randomNonce() ([]byte, error) {
	if o.nonce != nil {
		return o.nonce, nil
	}
	n := make([]byte, 32)
	if _, err := rand.Read(n); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return n, nil
}

// timestamp returns the current timestamp as described in RFC5849 section 3.3.
func (o *OAuth) timestamp() int64 {
	if !o.time.IsZero() {
		return o.time.Unix()
	}
	return time.Now().Unix()
}

// signatureBase prepares signature base string as described in RFC5849 section 3.4.1.
func signatureBase(req *http.Request, params map[string]string) string {
	method := strings.ToUpper(req.Method)
	baseURL := strings.ToLower(req.URL.Scheme+"://"+req.URL.Host) + req.URL.Path
	return strings.Join([]string{method, percentEncode(baseURL), normalizeParams(params)}, "&")
}

// normalizeParams parameters normalization as described in RFC5849 section 3.4.1.3.2.
func normalizeParams(params map[string]string) string {
	return percentEncode(formatParams(params, "%s=%s", "&"))
}

// percentEncode encodes parameters as described in RFC5849 section 3.6.
func percentEncode(s string) string {
	return strings.Replace(url.QueryEscape(s), "+", "%20", -1)
}

// formatParams formats parameters list to a string. Parameters are sorted by
// key name and then joined with the given separator. The format string is used
// to format key-value pairs.
func formatParams(params map[string]string, format, sep string) string {
	pairs := make([]string, len(params))
	for i, key := range maputil.SortKeys(params, sort.Strings) {
		pairs[i] = fmt.Sprintf(format, percentEncode(key), percentEncode(params[key]))
	}
	return strings.Join(pairs, sep)
}
