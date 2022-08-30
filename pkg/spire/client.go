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

package spire

import (
	"context"
	"errors"
	"net/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Client struct {
	ctx    context.Context
	waitCh chan error

	rpc    *rpc.Client
	addr   string
	signer ethereum.Signer
}

type ClientConfig struct {
	Signer  ethereum.Signer
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
func (c *Client) Wait() chan error {
	return c.waitCh
}

func (c *Client) PublishPrice(price *messages.Price) error {
	err := c.rpc.Call("API.PublishPrice", PublishPriceArg{Price: price}, &Nothing{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) PullPrices(assetPair string, feeder string) ([]*messages.Price, error) {
	resp := &PullPricesResp{}
	err := c.rpc.Call("API.PullPrices", PullPricesArg{FilterAssetPair: assetPair, FilterFeeder: feeder}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}

func (c *Client) PullPrice(assetPair string, feeder string) (*messages.Price, error) {
	resp := &PullPriceResp{}
	err := c.rpc.Call("API.PullPrice", PullPriceArg{AssetPair: assetPair, Feeder: feeder}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Price, nil
}

func (c *Client) contextCancelHandler() {
	defer func() { close(c.waitCh) }()
	<-c.ctx.Done()
	if err := c.rpc.Close(); err != nil {
		c.waitCh <- err
	}
}
