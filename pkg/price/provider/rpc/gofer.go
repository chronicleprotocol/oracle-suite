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

package rpc

import (
	"context"
	"errors"
	"net/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

var ErrNotStarted = errors.New("price provider RPC client is not started")

// Provider implements the provider.Provider interface. It uses a remote RPC
// server to fetch prices and models.
type Provider struct {
	ctx    context.Context
	waitCh chan error

	rpc     *rpc.Client
	network string
	address string
}

// NewProvider returns a new Provider instance.
func NewProvider(network, address string) (*Provider, error) {
	return &Provider{
		waitCh:  make(chan error),
		network: network,
		address: address,
	}, nil
}

// Start implements the supervisor.Service interface.
func (g *Provider) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	g.ctx = ctx
	client, err := rpc.DialHTTP(g.network, g.address)
	if err != nil {
		return err
	}
	g.rpc = client
	go g.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (g *Provider) Wait() chan error {
	return g.waitCh
}

// Models implements the provider.Provider interface.
func (g *Provider) Models(pairs ...provider.Pair) (map[provider.Pair]*provider.Model, error) {
	if g.rpc == nil {
		return nil, ErrNotStarted
	}
	resp := &NodesResp{}
	err := g.rpc.Call("API.Models", NodesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}

// Price implements the provider.Provider interface.
func (g *Provider) Price(pair provider.Pair) (*provider.Price, error) {
	if g.rpc == nil {
		return nil, ErrNotStarted
	}
	resp, err := g.Prices(pair)
	if err != nil {
		return nil, err
	}
	return resp[pair], nil
}

// Prices implements the provider.Provider interface.
func (g *Provider) Prices(pairs ...provider.Pair) (map[provider.Pair]*provider.Price, error) {
	if g.rpc == nil {
		return nil, ErrNotStarted
	}
	resp := &PricesResp{}
	err := g.rpc.Call("API.Prices", PricesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}

// Pairs implements the provider.Provider interface.
func (g *Provider) Pairs() ([]provider.Pair, error) {
	if g.rpc == nil {
		return nil, ErrNotStarted
	}
	resp := &PairsResp{}
	err := g.rpc.Call("API.Pairs", &Nothing{}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}

func (g *Provider) contextCancelHandler() {
	<-g.ctx.Done()
	g.waitCh <- g.rpc.Close()
}
