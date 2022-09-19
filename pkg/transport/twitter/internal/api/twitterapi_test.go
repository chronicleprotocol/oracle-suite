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
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummySigner struct {
	err error // optional error to return
}

func (s *dummySigner) Sign(req *http.Request) error {
	if s.err != nil {
		return s.err
	}
	req.Header.Set("Authorization", "oauth")
	return nil
}

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type testableAPI struct {
	API
	req *http.Request
	res *http.Response
}

func newTestableAPI() *testableAPI {
	t := &testableAPI{}
	t.API = API{
		Signer: &dummySigner{},
		Client: http.Client{
			Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				t.req = req
				return t.res, nil
			}),
		},
	}
	return t
}

func newResponse(statusCode int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func TestAPI_Me(t *testing.T) {
	body := []byte(`
		{
		  "data": {
			"created_at": "2013-12-14T04:35:55.000Z",
			"username": "TwitterDev",
			"pinned_tweet_id": "1255542774432063488",
			"id": "2244994945",
			"name": "Twitter Dev"
		  }
		}
	`)

	api := newTestableAPI()
	api.res = newResponse(http.StatusOK, body)
	res, err := api.Me(context.Background())
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, api.req.Method)
	assert.Equal(t, "https://api.twitter.com/2/users/me", api.req.URL.String())
	assert.Equal(t, "oauth", api.req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", api.req.Header.Get("Accept"))
	assert.Equal(t, "Twitter Dev", res.Data.Name)
	assert.Equal(t, "TwitterDev", res.Data.Username)
	assert.Equal(t, "2244994945", res.Data.ID)
}

func TestAPI_UserByUsername(t *testing.T) {
	body := []byte(`
		{
		  "data": {
			"created_at": "2013-12-14T04:35:55.000Z",
			"username": "TwitterDev",
			"pinned_tweet_id": "1255542774432063488",
			"id": "2244994945",
			"name": "Twitter Dev"
		  }
		}
	`)

	api := newTestableAPI()
	api.res = newResponse(http.StatusOK, body)
	res, err := api.UserByUsername(context.Background(), &UserByUsernameQuery{Username: "TwitterDev"})
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, api.req.Method)
	assert.Equal(t, "https://api.twitter.com/2/users/by/username/TwitterDev", api.req.URL.String())
	assert.Equal(t, "oauth", api.req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", api.req.Header.Get("Accept"))
	assert.Equal(t, "Twitter Dev", res.Data.Name)
	assert.Equal(t, "TwitterDev", res.Data.Username)
	assert.Equal(t, "2244994945", res.Data.ID)
}

func TestAPI_UserTweets(t *testing.T) {
	body := []byte(`
		{
			"data": [
				{
					"id": "1344801000000000000",
					"text": "Twitter API v2"
				},
				{
					"id": "1344801000000000001",
					"text": "Twitter API v2"
				}
			],
			"includes": {
				"media": [
					{
						"media_key": "3_1344801000000000000",
						"type": "photo",
						"url": "https://pbs.twimg.com/media/Ep0vZ8fVgAEY1Y9?format=jpg&name=small"
					}
				]
			}
		}
	`)

	api := newTestableAPI()
	api.res = newResponse(http.StatusOK, body)
	res, err := api.UserTweets(context.Background(), &UserTweetsQuery{
		UserID:      "2244994945",
		Expansions:  "attachments.media_keys",
		MediaFields: "url",
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, api.req.Method)
	assert.Equal(t, "https://api.twitter.com/2/users/2244994945/tweets?expansions=attachments.media_keys&media.fields=url", api.req.URL.String())
	assert.Equal(t, "oauth", api.req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", api.req.Header.Get("Accept"))
	assert.Equal(t, "1344801000000000000", res.Data[0].ID)
	assert.Equal(t, "1344801000000000001", res.Data[1].ID)
	assert.Equal(t, "photo", res.Includes.Media[0].Type)
	assert.Equal(t, "https://pbs.twimg.com/media/Ep0vZ8fVgAEY1Y9?format=jpg&name=small", res.Includes.Media[0].URL)
}

func TestAPI_CreateTweet(t *testing.T) {
	body := []byte(`
		{
			"data": {
				"id": "1344801000000000000",
				"text": "Twitter API v2"
			}
		}
	`)

	api := newTestableAPI()
	api.res = newResponse(http.StatusOK, body)
	res, err := api.CreateTweet(context.Background(), &CreateTweetRequest{
		Text: "Twitter API v2",
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, api.req.Method)
	assert.Equal(t, "https://api.twitter.com/2/tweets", api.req.URL.String())
	assert.Equal(t, "oauth", api.req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", api.req.Header.Get("Accept"))
	assert.Equal(t, "1344801000000000000", res.Data.ID)
	assert.Equal(t, "Twitter API v2", res.Data.Text)
}

func TestAPI_Upload(t *testing.T) {
	body := []byte(`
		{
			"media_id": 1344801000000000000,
			"media_id_string": "1344801000000000000",
			"media_key": "3_1344801000000000000",
			"size": 12345,
			"expires_after_secs": 86400,
			"processing_info": {
				"state": "pending",
				"check_after_secs": 5
			}
		}
	`)

	api := newTestableAPI()
	api.res = newResponse(http.StatusOK, body)
	res, err := api.Upload(context.Background(), io.NopCloser(strings.NewReader("test")))
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, api.req.Method)
	assert.Equal(t, "https://upload.twitter.com/1.1/media/upload.json", api.req.URL.String())
	assert.Equal(t, "oauth", api.req.Header.Get("Authorization"))
	assert.True(t, strings.HasPrefix(api.req.Header.Get("Content-Type"), "multipart/form-data; boundary="))
	assert.Equal(t, "3_1344801000000000000", res.MediaKey)
	assert.Equal(t, "1344801000000000000", res.MediaIDString)
}

func TestAPI_Error(t *testing.T) {
	body := []byte(`
		{
			"errors": [
				{
					"code": 89,
					"message": "Invalid or expired token."
				}
			]
		}
	`)

	api := newTestableAPI()
	api.res = newResponse(http.StatusUnauthorized, body)
	_, err := api.Me(context.Background())
	require.Error(t, err)

	assert.Equal(t, http.MethodGet, api.req.Method)
	assert.Equal(t, "https://api.twitter.com/2/users/me", api.req.URL.String())
	assert.Equal(t, "oauth", api.req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", api.req.Header.Get("Accept"))
	assert.Equal(t, "twitter callout status 401: invalid or expired token", err.Error())
}

func TestAPI_SignError(t *testing.T) {
	api := newTestableAPI()
	api.API.Signer = &dummySigner{err: errors.New("err")}
	_, err := api.Me(context.Background())
	require.Error(t, err)
	assert.Equal(t, "sign error: err", err.Error())
}
