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

package ghost

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/marshal"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "GHOST"

type Ghost struct {
	ctx    context.Context
	waitCh chan error

	priceProvider provider.Provider
	signer        ethereum.Signer
	transport     transport.Transport
	interval      time.Duration
	pairs         []provider.Pair
	log           log.Logger
}

// Config is the configuration for the Ghost.
type Config struct {
	// Pairs is a list supported pairs.
	Pairs []string
	// PriceProvider is an instance of the provider.Provider.
	PriceProvider provider.Provider
	// Signer is an instance of the ethereum.Signer which will be used to
	// sign prices.
	Signer ethereum.Signer
	// Transport is an implementation of transport used to send prices to
	// relayers.
	Transport transport.Transport
	// Interval describes how often we should send prices to the network.
	Interval time.Duration
	// Logger is a current logger interface used by the Ghost. The Logger
	// helps to monitor asynchronous processes.
	Logger log.Logger
}

func New(cfg Config) (*Ghost, error) {
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
	g := &Ghost{
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

func (g *Ghost) Start(ctx context.Context) error {
	if g.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	g.log.Infof("Starting")
	g.ctx = ctx
	go g.broadcasterRoutine()
	go g.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (g *Ghost) Wait() chan error {
	return g.waitCh
}

// broadcast sends price for single pair to the network. This method uses
// current price from the Provider, so it must be updated beforehand.
func (g *Ghost) broadcast(pair provider.Pair) error {
	var err error

	tick, err := g.priceProvider.Price(pair)
	if err != nil {
		return err
	}
	if tick.Error != "" {
		return errors.New(tick.Error)
	}

	// Create price:
	price := &oracle.Price{Wat: pair.Base + pair.Quote, Age: tick.Time}
	price.SetFloat64Price(tick.Price)

	// Sign price:
	err = price.Sign(g.signer)
	if err != nil {
		return err
	}

	// Broadcast price to P2P network:
	msg, err := createPriceMessage(price, tick)
	if err != nil {
		return err
	}
	if err := g.transport.Broadcast(messages.PriceV0MessageName, msg.AsV0()); err != nil {
		return err
	}
	if err := g.transport.Broadcast(messages.PriceV1MessageName, msg.AsV1()); err != nil {
		return err
	}
	return err
}

// broadcasterRoutine creates an asynchronous loop which fetches prices from exchanges and then
// sends them to the network at a specified interval.
func (g *Ghost) broadcasterRoutine() {
	if g.interval == 0 {
		return
	}
	var wg sync.WaitGroup
	ticker := time.NewTicker(g.interval)
	for {
		select {
		case <-g.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			// Send prices to the network:
			// Signing may be slow, especially with high KDF so this is why
			// we are using goroutines here.
			wg.Add(1)
			go func() {
				for _, pair := range g.pairs {
					err := g.broadcast(pair)
					if err != nil {
						g.log.
							WithFields(log.Fields{"assetPair": pair}).
							WithError(err).
							Warn("Unable to broadcast price")
					} else {
						g.log.
							WithFields(log.Fields{"assetPair": pair}).
							Info("Price broadcast")
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func (g *Ghost) contextCancelHandler() {
	defer func() { close(g.waitCh) }()
	defer g.log.Info("Stopped")
	<-g.ctx.Done()
}

func createPriceMessage(op *oracle.Price, gp *provider.Price) (*messages.Price, error) {
	trace, err := marshal.Marshall(marshal.JSON, gp)
	if err != nil {
		return nil, err
	}
	return &messages.Price{
		Price: op,
		Trace: trace,
	}, nil
}
