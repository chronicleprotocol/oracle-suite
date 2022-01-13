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

package ssb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.cryptoscope.co/muxrpc/v2"
	"go.cryptoscope.co/ssb"
	"go.cryptoscope.co/ssb/client"
	"go.cryptoscope.co/ssb/invite"
	"go.cryptoscope.co/ssb/message"
	refs "go.mindeco.de/ssb-refs"
)

type Client struct {
	ctx    context.Context
	doneCh chan struct{}

	rpc    *client.Client
	keys   ssb.KeyPair
	shs    string
	invite invite.Token
}

func (c *Client) Publish(v interface{}) error {
	var resp string
	defer func() { fmt.Println(resp) }()
	return c.rpc.Async(c.ctx, &resp, muxrpc.TypeString, muxrpc.Method{"publish"}, v)
}

func (c *Client) Log() error {
	src, err := c.rpc.Source(c.ctx, muxrpc.TypeJSON, muxrpc.Method{"createLogStream"}, message.CreateLogArgs{
		CommonArgs: message.CommonArgs{
			Live: true,
		},
		StreamArgs: message.StreamArgs{
			Limit:   -1,
			Reverse: false,
		},
	})
	if err != nil {
		return err
	}
	for nxt := src.Next(c.ctx); nxt; nxt = src.Next(c.ctx) {
		b, err := src.Bytes()
		if err != nil {
			return err
		}

		fmt.Println(string(b))
	}
	return nil
}

func (c *Client) Hist() error {
	src, err := c.rpc.Source(c.ctx, muxrpc.TypeBinary, muxrpc.Method{"createHistoryStream"}, message.CreateHistArgs{
		CommonArgs: message.CommonArgs{
			Live: true,
		},
		StreamArgs: message.StreamArgs{
			Limit:   -1,
			Reverse: false,
		},
	})
	if err != nil {
		return err
	}
	for nxt := src.Next(c.ctx); nxt; nxt = src.Next(c.ctx) {
		b, err := src.Bytes()
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	}
	return nil
}
func (c *Client) Whoami() ([]byte, error) {
	var resp []byte
	err := c.rpc.Async(c.ctx, &resp, muxrpc.TypeBinary, muxrpc.Method{"whoami"})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Last(id, contentType string, limit int64) ([]byte, error) {
	feedRef, err := refs.ParseFeedRef(id)
	if err != nil {
		return nil, err
	}
	src, err := c.rpc.Source(c.ctx, muxrpc.TypeJSON, muxrpc.Method{"createUserStream"}, message.CreateHistArgs{
		CommonArgs: message.CommonArgs{
			Keys: true,
		},
		StreamArgs: message.StreamArgs{
			Limit:   limit,
			Reverse: true,
		},
		ID: feedRef,
	})
	if err != nil {
		return nil, err
	}
	var data struct {
		Value struct {
			Content FeedAssetPrice `json:"content"`
		} `json:"value"`
	}
	for nxt := src.Next(c.ctx); nxt; nxt = src.Next(c.ctx) {
		bytes, err := src.Bytes()
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(bytes, &data); err != nil {
			return nil, err
		}
		if contentType == "" || data.Value.Content.Type == contentType {
			return bytes, nil
		}
	}
	return nil, errors.New("no data in the stream")
}
