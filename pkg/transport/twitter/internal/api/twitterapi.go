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
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const (
	twitterAPIV2     = "https://api.twitter.com/2/"
	twitterAPIUpload = "https://upload.twitter.com/1.1/media/upload.json"
)

type Signer interface {
	Sign(req *http.Request) error
}

// API is a Twitter API client. It implements minimal functionality required
// for the transport.
type API struct {
	Signer Signer
	Client http.Client
}

// Me returns the authenticated user.
func (a *API) Me(ctx context.Context) (*UserResponse, error) {
	res := &UserResponse{}
	if err := a.call(
		ctx,
		http.MethodGet,
		twitterAPIV2+"users/me",
		nil, nil,
		res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

// UserByUsername returns a user by name.
func (a *API) UserByUsername(ctx context.Context, r *UserByUsernameQuery) (*UserResponse, error) {
	res := &UserResponse{}
	if err := a.call(
		ctx,
		http.MethodGet,
		twitterAPIV2+"users/by/username/"+r.Username,
		nil,
		nil,
		res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

// CreateTweet creates a new tweet.
func (a *API) CreateTweet(ctx context.Context, r *CreateTweetRequest) (*CreateTweetResponse, error) {
	res := &CreateTweetResponse{}
	if err := a.call(
		ctx,
		http.MethodPost,
		twitterAPIV2+"tweets",
		nil,
		r,
		res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

// UserTweets returns a user's tweets.
func (a *API) UserTweets(ctx context.Context, q *UserTweetsQuery) (*UserTweetsResponse, error) {
	res := &UserTweetsResponse{}
	if err := a.call(
		ctx,
		http.MethodGet,
		twitterAPIV2+"users/"+q.UserID+"/tweets",
		q,
		nil,
		res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

// Upload uploads a media file that can be attached to a Tweet.
func (a *API) Upload(ctx context.Context, data io.Reader) (*UploadResponse, error) {
	// Prepare the multipart request.
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreateFormFile("media", "media")
	if err != nil {
		return nil, fmt.Errorf("multipart error: %w", err)
	}
	if _, err := io.Copy(fw, data); err != nil {
		return nil, fmt.Errorf("copy error: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("close error: %w", err)
	}
	// Do the request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, twitterAPIUpload, body)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	req.Header.Add("Content-Type", mw.FormDataContentType())
	req.Header.Add("Accept", "application/json")
	if err := a.Signer.Sign(req); err != nil {
		return nil, fmt.Errorf("sign error: %w", err)
	}
	res, err := a.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	// Decode the response.
	uploadRes := &UploadResponse{}
	if err := a.decodeJSON(res, uploadRes); err != nil {
		return nil, err
	}
	return uploadRes, nil
}

func (a *API) call(ctx context.Context, method, url string, query, request, response any) error {
	// Encode request body.
	body, err := a.encodeJSON(request)
	if err != nil {
		return err
	}
	// Do the request.
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	req.URL.RawQuery = encodeQuery(query)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	if err := a.Signer.Sign(req); err != nil {
		return fmt.Errorf("sign error: %w", err)
	}
	res, err := a.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer res.Body.Close()
	// Decode the response.
	return a.decodeJSON(res, response)
}

func (a *API) encodeJSON(v any) (io.Reader, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("request marshal error %w", err)
	}
	return bytes.NewReader(body), nil
}

func (a *API) decodeJSON(res *http.Response, v any) error {
	decoder := json.NewDecoder(res.Body)
	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		err := &ErrorResponse{StatusCode: res.StatusCode}
		if err := decoder.Decode(err); err != nil {
			return fmt.Errorf("response decode error: %w", err)
		}
		return err
	}
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("response decode error: %w", err)
	}
	return nil
}

type UserTweetsQuery struct {
	UserID          string
	Exclude         string `param:"exclude"`
	Expansions      string `param:"expansions"`
	MaxResults      int    `param:"max_results"`
	PaginationToken string `param:"pagination_token"`
	SinceID         string `param:"since_id"`
	UntilID         string `param:"until_id"`
	StartTime       string `param:"start_time"`
	EndTime         string `param:"end_time"`
	MediaFields     string `param:"media.fields"`
	PollFields      string `param:"poll.fields"`
	TweetFields     string `param:"tweet.fields"`
	PlaceFields     string `param:"place.fields"`
	UserFields      string `param:"user.fields"`
}

type UserByUsernameQuery struct {
	Username    string
	Expansions  string `param:"expansions"`
	UserFields  string `param:"user.fields"`
	TweetFields string `param:"tweet.fields"`
}

type UserResponse struct {
	Data struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"data"`
}

type UserTweetsResponse struct {
	Data []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
	Includes struct {
		Media []*MediaResponse `json:"media,omitempty"`
	} `json:"includes,omitempty"`
	Meta struct {
		OldestID    string `json:"oldest_id"`
		NewestID    string `json:"newest_id"`
		ResultCount int    `json:"result_count"`
		NextToken   string `json:"next_token"`
	} `json:"meta"`
}

type CreateTweetRequest struct {
	Text  string                   `json:"text"`
	Media *CreateTweetMediaRequest `json:"media,omitempty"`
}

type CreateTweetMediaRequest struct {
	MediaIDs []string `json:"media_ids,omitempty"`
}

type CreateTweetResponse struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}
}

type UploadResponse struct {
	MediaID          int64  `json:"media_id"`
	MediaIDString    string `json:"media_id_string"`
	MediaKey         string `json:"media_key"`
	Size             int    `json:"size"`
	ExpiresAfterSecs int    `json:"expires_after_secs"`
	Image            struct {
		ImageType string `json:"image_type"`
		W         int    `json:"w"`
		H         int    `json:"h"`
	} `json:"image"`
}

type MediaResponse struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type ErrorResponse struct {
	StatusCode int
	Errors     []struct {
		Parameters any    `json:"parameters"`
		Message    string `json:"message"`
	} `json:"errors"`
}

func (e ErrorResponse) Error() string {
	errs := make([]string, len(e.Errors))
	for n, err := range e.Errors {
		errs[n] = strings.TrimSuffix(strings.ToLower(err.Message), ".")
	}

	return fmt.Sprintf("twitter callout status %d: %s", e.StatusCode, strings.Join(errs, ","))
}

// encodeQuery encodes the given struct as a query string. The struct fields
// must be tagged with the `param` tag to be included in the query string.
// Zero values are ignored.
func encodeQuery(v any) string {
	if v == nil {
		return ""
	}
	rv := reflect.Indirect(reflect.ValueOf(v))
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		panic("not a struct")
	}
	query := url.Values{}
	for i := 0; i < rv.NumField(); i++ {
		tf := rt.Field(i)
		vf := rv.Field(i)
		tag := tf.Tag.Get("param")
		if len(tag) == 0 || vf.IsZero() {
			continue
		}
		query.Add(tag, fmt.Sprint(vf.Interface()))
	}
	return query.Encode()
}
