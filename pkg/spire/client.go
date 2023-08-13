//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package spire

import (
	"context"
	"errors"
	"net/rpc"

	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Client struct {
	ctx    context.Context
	waitCh chan error

	rpc    *rpc.Client
	addr   string
	signer wallet.Key
}

type ClientConfig struct {
	Signer  wallet.Key
	Address string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	return &Client{
		waitCh: make(chan error),
		addr:   cfg.Address,
		signer: cfg.Signer,
	}, nil
}

func (c *Client) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	c.ctx = ctx
	client, err := rpc.DialHTTP("tcp", c.addr)
	if err != nil {
		return err
	}
	c.rpc = client
	go c.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (c *Client) Wait() <-chan error {
	return c.waitCh
}

func (c *Client) Publish(dataPoint *messages.DataPoint) error {
	err := c.rpc.Call("API.Publish", PublishArg{DataPoint: dataPoint}, &Nothing{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) PullPrices(assetPair string, feed string) ([]*messages.DataPoint, error) {
	resp := &PullDataPointsResp{}
	err := c.rpc.Call("API.PullPoints", PullPricesArg{FilterAssetPair: assetPair, FilterFeed: feed}, resp)
	if err != nil {
		return nil, err
	}
	return resp.DataPoints, nil
}

func (c *Client) PullPrice(assetPair string, feed string) (*messages.DataPoint, error) {
	resp := &PullDataPointResp{}
	err := c.rpc.Call("API.PullPoint", PullPriceArg{AssetPair: assetPair, Feed: feed}, resp)
	if err != nil {
		return nil, err
	}
	return resp.DataPoint, nil
}

func (c *Client) contextCancelHandler() {
	defer func() { close(c.waitCh) }()
	<-c.ctx.Done()
	if err := c.rpc.Close(); err != nil {
		c.waitCh <- err
	}
}
