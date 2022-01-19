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
	"fmt"
	"log"

	"go.cryptoscope.co/muxrpc/v2"
	"go.cryptoscope.co/ssb/client"
	"go.cryptoscope.co/ssb/message"
	refs "go.mindeco.de/ssb-refs"
)

const methodPublish = "publish"
const methodWhoAmI = "whoami"
const methodCreateHistoryStream = "createHistoryStream"
const methodCreateLogStream = "createLogStream"
const methodCreateUserStream = "createUserStream"

type Client struct {
	ctx context.Context
	rpc *client.Client
}

func (c *Client) Transmit(v interface{}) ([]byte, error) {
	var resp []byte
	// TODO Add Rate Limiter
	return resp, c.rpc.Async(c.ctx, &resp, muxrpc.TypeBinary, muxrpc.Method{methodPublish}, v)
}

func (c *Client) WhoAmI() ([]byte, error) {
	var resp []byte
	return resp, c.rpc.Async(c.ctx, &resp, muxrpc.TypeBinary, muxrpc.Method{methodWhoAmI})
}

func (c *Client) ReceiveLast(id, contentType string, limit int64) ([]byte, error) {
	feedRef, err := refs.ParseFeedRef(id)
	if err != nil {
		return nil, err
	}
	ch, err := c.callSSB(methodCreateUserStream, message.CreateHistArgs{
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
	for b := range ch {
		if err = json.Unmarshal(b, &data); err != nil {
			return nil, err
		}
		t := data.Value.Content.Type
		if contentType == "" || t == contentType {
			return b, nil
		}
	}
	if contentType != "" {
		return nil, fmt.Errorf("no data of type %s in the stream for ref: %s", contentType, feedRef.Ref())
	}
	return nil, fmt.Errorf("no data in the stream for ref: %s", feedRef.Ref())
}

func (c *Client) LogStream() (chan []byte, error) {
	return c.callSSB(methodCreateLogStream, message.CreateLogArgs{
		CommonArgs: message.CommonArgs{
			Live: true,
		},
		StreamArgs: message.StreamArgs{
			Limit:   -1,
			Reverse: false,
		},
	})
}

func (c *Client) HistoryStream() (chan []byte, error) {
	return c.callSSB(methodCreateHistoryStream, message.CreateHistArgs{
		CommonArgs: message.CommonArgs{
			Live: true,
		},
		StreamArgs: message.StreamArgs{
			Limit:   -1,
			Reverse: false,
		},
	})
}

func (c *Client) callSSB(method string, arg interface{}) (chan []byte, error) {
	src, err := c.rpc.Source(c.ctx, muxrpc.TypeBinary, muxrpc.Method{method}, arg)
	if err != nil {
		return nil, err
	}
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		for nxt := src.Next(c.ctx); nxt; nxt = src.Next(c.ctx) {
			b, err := src.Bytes()
			if err != nil {
				log.Println(err)
				return
			}
			ch <- b
		}
	}()
	return ch, err
}
