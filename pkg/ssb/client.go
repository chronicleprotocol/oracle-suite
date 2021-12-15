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
	"errors"

	"go.cryptoscope.co/muxrpc/v2"
	"go.cryptoscope.co/ssb"
	"go.cryptoscope.co/ssb/client"
	"go.cryptoscope.co/ssb/invite"

	ssb2 "github.com/chronicleprotocol/oracle-suite/cmd/keeman/ssb"
)

type Client struct {
	ctx    context.Context
	doneCh chan struct{}

	rpc    *client.Client
	keys   ssb.KeyPair
	shs    string
	invite invite.Token
}
type ClientConfig struct {
	Keys   ssb.KeyPair
	Shs    string
	Invite invite.Token
}

func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	rpc, err := client.NewTCP(
		cfg.Keys,
		cfg.Invite.Address,
		client.WithSHSAppKey(cfg.Shs),
		client.WithContext(ctx),
		// client.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		ctx:    ctx,
		doneCh: make(chan struct{}),
		rpc:    rpc,
	}, nil
}

func (c *Client) PublishPrice(price ssb2.FeedAssetPrice) error {
	var resp string
	return c.rpc.Async(c.ctx, &resp, muxrpc.TypeString, muxrpc.Method{"publish"}, price)
}
