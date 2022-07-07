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

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
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

// Start implements the store.Store interface.
func (d *Datastore) Start(ctx context.Context) error {
	if d.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	d.log.Info("Starting")
	d.ctx = ctx
	go d.contextCancelHandler()
	return d.collectorLoop()
}

// Wait implements the store.Store interface.
func (d *Datastore) Wait() chan error {
	return d.waitCh
}

// Prices implements the store.Store interface.
func (d *Datastore) Prices() store.PriceStore {
	return d.priceStore
}

// collectPrice adds a price from a feeder which may be used to update
// Oracle contract. The price will be added only if a feeder is
// allowed to send prices.
func (d *Datastore) collectPrice(msg *messages.Price) error {
	from, err := msg.Price.From(d.signer)
	if err != nil {
		return errInvalidSignature
	}
	if _, ok := d.pairs[msg.Price.Wat]; !ok {
		return errUnknownPair
	}
	if !d.isFeedAllowed(msg.Price.Wat, *from) {
		return errUnknownFeeder
	}
	if msg.Price.Val.Cmp(big.NewInt(0)) <= 0 {
		return errInvalidPrice
	}

	d.priceStore.Add(*from, msg)

	return nil
}

// collectorLoop creates a asynchronous loop which fetches prices from feeders.
func (d *Datastore) collectorLoop() error {
	go func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		for {
			select {
			case <-d.ctx.Done():
				return
			case m := <-d.transport.Messages(messages.PriceMessageName):
				// If there was a problem while reading prices from the transport:
				if m.Error != nil {
					d.log.
						WithError(m.Error).
						Warn("Unable to read prices from the transport")
					continue
				}
				price, ok := m.Message.(*messages.Price)
				if !ok {
					d.log.Error("Unexpected value returned from transport layer")
					continue
				}

				// Try to collect received price:
				err := d.collectPrice(price)

				// Print logs:
				if err != nil {
					d.log.
						WithError(err).
						WithFields(price.Price.Fields(d.signer)).
						Warn("Received invalid price")
				} else {
					d.log.
						WithFields(price.Price.Fields(d.signer)).
						Info("Price received")
				}
			}
		}
	}()

	return nil
}

func (d *Datastore) isFeedAllowed(assetPair string, address ethereum.Address) bool {
	for _, a := range d.pairs[assetPair].Feeds {
		if a == address {
			return true
		}
	}
	return false
}

func (d *Datastore) contextCancelHandler() {
	defer func() { close(d.waitCh) }()
	defer d.log.Info("Stopped")
	<-d.ctx.Done()
}
