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

package memory

import (
	"context"
	"errors"
	"math/big"
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/datastore"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "DATASTORE"

var errInvalidSignature = errors.New("received price has an invalid signature")
var errInvalidPrice = errors.New("received price is invalid")
var errUnknownPair = errors.New("received pair is not configured")
var errUnknownFeeder = errors.New("feeder is not allowed to send prices")

// Datastore reads and stores prices from the P2P network.
type Datastore struct {
	ctx    context.Context
	mu     sync.Mutex
	waitCh chan error

	signer     ethereum.Signer
	transport  transport.Transport
	pairs      map[string]*Pair
	priceStore *PriceStore
	log        log.Logger
}

type Config struct {
	// Signer is an instance of the ethereum.Signer which will be used to
	// verify price signatures.
	Signer ethereum.Signer
	// Transport is a implementation of transport used to fetch prices from
	// feeders.
	Transport transport.Transport
	// Pairs is the list supported pairs by the datastore with their
	// configuration.
	Pairs map[string]*Pair
	// Logger is a current logger interface used by the Datastore.
	// The Logger is required to monitor asynchronous processes.
	Logger log.Logger
}

type Pair struct {
	// Feeds is the list of Ethereum addresses from which prices will be
	// accepted.
	Feeds []ethereum.Address
}

func NewDatastore(cfg Config) (*Datastore, error) {
	return &Datastore{
		waitCh:     make(chan error),
		signer:     cfg.Signer,
		transport:  cfg.Transport,
		pairs:      cfg.Pairs,
		priceStore: NewPriceStore(),
		log:        cfg.Logger.WithField("tag", LoggerTag),
	}, nil
}

// Start implements the datastore.Datastore interface.
func (c *Datastore) Start(ctx context.Context) error {
	c.log.Info("Starting")
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	c.ctx = ctx
	go c.contextCancelHandler()
	return c.collectorLoop()
}

// Wait implements the datastore.Datastore interface.
func (c *Datastore) Wait() chan error {
	return c.waitCh
}

// Prices implements the datastore.Datastore interface.
func (c *Datastore) Prices() datastore.PriceStore {
	return c.priceStore
}

// collectPrice adds a price from a feeder which may be used to update
// Oracle contract. The price will be added only if a feeder is
// allowed to send prices.
func (c *Datastore) collectPrice(msg *messages.Price) error {
	from, err := msg.Price.From(c.signer)
	if err != nil {
		return errInvalidSignature
	}
	if _, ok := c.pairs[msg.Price.Wat]; !ok {
		return errUnknownPair
	}
	if !c.isFeedAllowed(msg.Price.Wat, *from) {
		return errUnknownFeeder
	}
	if msg.Price.Val.Cmp(big.NewInt(0)) <= 0 {
		return errInvalidPrice
	}

	c.priceStore.Add(*from, msg)

	return nil
}

// collectorLoop creates a asynchronous loop which fetches prices from feeders.
func (c *Datastore) collectorLoop() error {
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		for {
			select {
			case <-c.ctx.Done():
				return
			case m := <-c.transport.Messages(messages.PriceMessageName):
				// If there was a problem while reading prices from the transport:
				if m.Error != nil {
					c.log.
						WithError(m.Error).
						Warn("Unable to read prices from the transport")
					continue
				}
				price, ok := m.Message.(*messages.Price)
				if !ok {
					c.log.Error("Unexpected value returned from transport layer")
					continue
				}

				// Try to collect received price:
				err := c.collectPrice(price)

				// Print logs:
				if err != nil {
					c.log.
						WithError(err).
						WithFields(price.Price.Fields(c.signer)).
						Warn("Received invalid price")
				} else {
					c.log.
						WithFields(price.Price.Fields(c.signer)).
						Info("Price received")
				}
			}
		}
	}()

	return nil
}

func (c *Datastore) isFeedAllowed(assetPair string, address ethereum.Address) bool {
	for _, a := range c.pairs[assetPair].Feeds {
		if a == address {
			return true
		}
	}
	return false
}

func (c *Datastore) contextCancelHandler() {
	defer func() { close(c.waitCh) }()
	defer c.log.Info("Stopped")
	<-c.ctx.Done()
}
