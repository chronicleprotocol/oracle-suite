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

package feed

import (
	"context"
	"errors"

	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/marshal"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "FEED"

// Feed is a service which periodically fetches prices and then sends them to
// the Oracle network using transport layer.
type Feed struct {
	ctx    context.Context
	waitCh chan error

	priceProvider provider.Provider
	signer        wallet.Key
	transport     transport.TransportService
	interval      *timeutil.Ticker
	pairs         []provider.Pair
	log           log.Logger
}

// Config is the configuration for the Feed.
type Config struct {
	// Pairs is a list supported pairs in the format "QUOTE/BASE".
	Pairs []string

	// PriceProvider is a price provider which is used to fetch prices.
	PriceProvider provider.Provider

	// Signer is a wallet used to sign prices.
	Signer wallet.Key

	// Transport is an implementation of transport used to send prices to
	// the network.
	Transport transport.TransportService

	// Interval describes how often we should send prices to the network.
	Interval *timeutil.Ticker

	// Logger is a current logger interface used by the Feed.
	Logger log.Logger
}

// New creates a new instance of the Feed.
func New(cfg Config) (*Feed, error) {
	if cfg.PriceProvider == nil {
		return nil, errors.New("price provider must not be nil")
	}
	if cfg.Signer == nil {
		return nil, errors.New("signer must not be nil")
	}
	if cfg.Transport == nil {
		return nil, errors.New("transport must not be nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	pairs, err := provider.NewPairs(cfg.Pairs...)
	if err != nil {
		return nil, err
	}
	g := &Feed{
		waitCh:        make(chan error),
		priceProvider: cfg.PriceProvider,
		signer:        cfg.Signer,
		transport:     cfg.Transport,
		interval:      cfg.Interval,
		pairs:         pairs,
		log:           cfg.Logger.WithField("tag", LoggerTag),
	}
	return g, nil
}

// Start implements the supervisor.Service interface.
func (g *Feed) Start(ctx context.Context) error {
	if g.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	g.log.Debug("Starting")
	g.ctx = ctx
	g.interval.Start(g.ctx)
	go g.broadcasterRoutine()
	go g.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (g *Feed) Wait() <-chan error {
	return g.waitCh
}

// broadcast sends price for single pair to the network. This method uses
// current price from the Provider, so it must be updated beforehand.
func (g *Feed) broadcast(pair provider.Pair) error {
	var err error

	// Create price.
	tick, err := g.priceProvider.Price(pair)
	if err != nil {
		return err
	}
	if tick.Error != "" {
		return errors.New(tick.Error)
	}
	price := &median.Price{Wat: pair.Base + pair.Quote, Age: tick.Time}
	price.SetFloat64Price(tick.Price)

	// Sign price.
	err = price.Sign(g.signer)
	if err != nil {
		return err
	}

	// Broadcast price to P2P network.
	msg, err := toPriceMessage(price, tick)
	if err != nil {
		return err
	}
	if err = g.transport.Broadcast(messages.PriceV0MessageName, msg.AsV0()); err != nil {
		return err
	}
	if err = g.transport.Broadcast(messages.PriceV1MessageName, msg.AsV1()); err != nil {
		return err
	}
	if err = g.transport.Broadcast(messages.DataPointV1MessageName, messages.Price2DataPoint(msg)); err != nil {
		return err
	}
	return err
}

func (g *Feed) broadcasterRoutine() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case <-g.interval.TickCh():
			// Send prices to the network.
			for _, pair := range g.pairs {
				if err := g.broadcast(pair); err != nil {
					g.log.
						WithField("assetPair", pair).
						WithError(err).
						Warn("Unable to broadcast price")
					continue
				}
				g.log.
					WithField("assetPair", pair).
					Info("Price broadcast")
			}
		}
	}
}

func (g *Feed) contextCancelHandler() {
	defer func() { close(g.waitCh) }()
	defer g.log.Debug("Stopped")
	<-g.ctx.Done()
}

func toPriceMessage(price *median.Price, provider *provider.Price) (*messages.Price, error) {
	trace, err := marshal.Marshall(marshal.JSON, provider)
	if err != nil {
		return nil, err
	}
	return &messages.Price{
		Price: price,
		Trace: trace,
	}, nil
}
