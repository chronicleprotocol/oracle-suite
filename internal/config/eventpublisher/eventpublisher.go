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

package eventpublisher

import (
	"context"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher"
	eventObserverEth "github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var eventPublisherFactory = func(ctx context.Context, cfg publisher.Config) (*publisher.EventPublisher, error) {
	return publisher.New(ctx, cfg)
}

type EventPublisher struct {
	Listeners struct {
		Wormhole wormholeListener `json:"wormhole"`
	} `json:"listeners"`
}

type wormholeListener struct {
	Enable        bool     `json:"enable"`
	Interval      int64    `json:"interval"`
	Confirmations int      `json:"confirmations"`
	MaxBlocks     int      `json:"maxBlocks"`
	Addresses     []string `json:"addresses"`
}

type Dependencies struct {
	Context        context.Context
	Signer         ethereum.Signer
	EthereumClient eventObserverEth.EthClient
	Transport      transport.Transport
	Logger         log.Logger
}

type DatastoreDependencies struct {
	Context   context.Context
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *EventPublisher) Configure(d Dependencies) (*publisher.EventPublisher, error) {
	var lis []publisher.Listener
	var sig []publisher.Signer
	if c.Listeners.Wormhole.Enable {
		var addrs []ethereum.Address
		for _, addr := range c.Listeners.Wormhole.Addresses {
			addrs = append(addrs, ethereum.HexToAddress(addr))
		}
		interval := c.Listeners.Wormhole.Interval
		if interval < 1 {
			interval = 1
		}
		lis = append(lis, eventObserverEth.NewWormholeListener(eventObserverEth.WormholeListenerConfig{
			Client:       d.EthereumClient,
			Addresses:    addrs,
			Interval:     time.Second * time.Duration(interval),
			BlocksBehind: c.Listeners.Wormhole.Confirmations,
			MaxBlocks:    c.Listeners.Wormhole.MaxBlocks,
			Log:          d.Logger,
		}))
	}
	sig = append(sig, eventObserverEth.NewSigner(d.Signer, []string{eventObserverEth.WormholeEventType}))
	cfg := publisher.Config{
		Listeners: lis,
		Signers:   sig,
		Transport: d.Transport,
		Logger:    d.Logger,
	}
	return eventPublisherFactory(d.Context, cfg)
}
