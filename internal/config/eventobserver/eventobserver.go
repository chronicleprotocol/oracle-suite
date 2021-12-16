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

package eventobserver

import (
	"context"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/observer"
	eventObserverEth "github.com/chronicleprotocol/oracle-suite/pkg/event/observer/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var eventObserverFactory = func(ctx context.Context, cfg observer.Config) (*observer.EventObserver, error) {
	return observer.NewEventObserver(ctx, cfg)
}

type EventObserver struct {
	Listeners struct {
		Wormhole wormholeListener `json:"wormhole"`
	} `json:"listeners"`
}

type wormholeListener struct {
	Enable        bool     `json:"enable"`
	Interval      int64    `json:"interval"`
	Confirmations int      `json:"confirmations"`
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

func (c *EventObserver) ConfigureLeeloo(d Dependencies) (*observer.EventObserver, error) {
	var lis []observer.Listener
	var sig []observer.Signer
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
			Client:        d.EthereumClient,
			Addresses:     addrs,
			Interval:      time.Second * time.Duration(interval),
			Confirmations: c.Listeners.Wormhole.Confirmations,
		}))
	}
	sig = append(sig, eventObserverEth.NewSigner(d.Signer, []string{eventObserverEth.WormholeEventType}))
	cfg := observer.Config{
		Listeners: lis,
		Signers:   sig,
		Transport: d.Transport,
		Logger:    d.Logger,
	}
	return eventObserverFactory(d.Context, cfg)
}
