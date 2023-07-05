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

package store

import (
	"context"
	"errors"
	"math/big"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "PRICE_STORE"

var ErrInvalidSignature = errors.New("received price has an invalid signature")
var ErrInvalidPrice = errors.New("received price is invalid")
var ErrUnknownPair = errors.New("received pair is not configured")
var ErrUnknownFeed = errors.New("messages from feed are not allowed")

// PriceStore contains a list of prices.
type PriceStore struct {
	ctx       context.Context
	storage   Storage
	transport transport.Transport
	pairs     []string
	feeds     []string
	log       log.Logger
	recover   crypto.Recoverer
	waitCh    chan error
}

// Config is the configuration for Storage.
type Config struct {
	// Storage is the storage implementation.
	Storage Storage

	// Transport is an implementation of transport used to fetch prices from
	// feeds.
	Transport transport.Transport

	// Pairs is the list of asset pairs which are supported by the store.
	Pairs []string

	// Feeds is the list of feeds which are supported by the store.
	Feeds []types.Address

	// Logger is a current logger interface used by the PriceStore.
	// The Logger is required to monitor asynchronous processes.
	Logger log.Logger

	// Recoverer provides a method to recover the public key from a signature.
	// The default is crypto.ECREcoverer.
	Recoverer crypto.Recoverer
}

// Storage provides an interface to the price storage.
type Storage interface {
	// Add adds a price to the store. The method is thread-safe.
	Add(ctx context.Context, from types.Address, msg *messages.Price) error

	// GetAll returns all prices. The method is thread-safe.
	GetAll(ctx context.Context) (map[FeedPrice]*messages.Price, error)

	// GetByAssetPair returns all prices for given asset pair. The method is
	// thread-safe.
	GetByAssetPair(ctx context.Context, pair string) ([]*messages.Price, error)

	// GetByFeed returns the latest price for given asset pair sent by given
	// feed. The method is thread-safe.
	GetByFeed(ctx context.Context, pair string, feed types.Address) (*messages.Price, error)
}

type FeedPrice struct {
	AssetPair string
	Feed      types.Address
}

// New creates a new store instance.
func New(cfg Config) (*PriceStore, error) {
	if cfg.Storage == nil {
		return nil, errors.New("storage must not be nil")
	}
	if cfg.Transport == nil {
		return nil, errors.New("transport must not be nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	if cfg.Recoverer == nil {
		cfg.Recoverer = crypto.ECRecoverer
	}
	feeds := make([]string, len(cfg.Feeds))
	for i, feed := range cfg.Feeds {
		feeds[i] = feed.String()
	}
	return &PriceStore{
		storage:   cfg.Storage,
		transport: cfg.Transport,
		pairs:     cfg.Pairs,
		feeds:     feeds,
		log:       cfg.Logger.WithField("tag", LoggerTag),
		recover:   cfg.Recoverer,
		waitCh:    make(chan error),
	}, nil
}

// Start implements the supervisor.Service interface.
func (p *PriceStore) Start(ctx context.Context) error {
	if p.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	p.log.Debug("Starting")
	p.ctx = ctx
	go p.priceCollectorRoutine()
	go p.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (p *PriceStore) Wait() <-chan error {
	return p.waitCh
}

// Add adds a new price to the list. If a price from same feed already
// exists, the newer one will be used.
func (p *PriceStore) Add(ctx context.Context, from types.Address, msg *messages.Price) error {
	return p.storage.Add(ctx, from, msg)
}

// GetAll returns all prices.
func (p *PriceStore) GetAll(ctx context.Context) (map[FeedPrice]*messages.Price, error) {
	return p.storage.GetAll(ctx)
}

// GetByAssetPair returns all prices for given asset pair.
func (p *PriceStore) GetByAssetPair(ctx context.Context, pair string) ([]*messages.Price, error) {
	return p.storage.GetByAssetPair(ctx, pair)
}

// GetByFeed returns the latest price for given asset pair sent by given feed.
func (p *PriceStore) GetByFeed(ctx context.Context, pair string, feed types.Address) (*messages.Price, error) {
	return p.storage.GetByFeed(ctx, pair, feed)
}

func (p *PriceStore) collectPrice(price *messages.Price) error {
	from, err := price.Price.From(p.recover)
	if err != nil {
		return ErrInvalidSignature
	}
	if !p.isFeedSupported(from.String()) {
		return ErrUnknownFeed
	}
	if !p.isPairSupported(price.Price.Wat) {
		return ErrUnknownPair
	}
	if price.Price.Val.Cmp(big.NewInt(0)) <= 0 {
		return ErrInvalidPrice
	}
	return p.Add(p.ctx, *from, price)
}

func (p *PriceStore) isPairSupported(pair string) bool {
	for _, a := range p.pairs {
		if a == pair {
			return true
		}
	}
	return false
}
func (p *PriceStore) isFeedSupported(feed string) bool {
	for _, a := range p.feeds {
		if a == feed {
			return true
		}
	}
	return false
}

func (p *PriceStore) priceCollectorRoutine() {
	priceV0Ch := p.transport.Messages(messages.PriceV0MessageName)
	priceV1Ch := p.transport.Messages(messages.PriceV1MessageName)
	for {
		select {
		case <-p.ctx.Done():
			return
		case msg := <-priceV0Ch:
			p.handlePriceMessage(msg)
		case msg := <-priceV1Ch:
			p.handlePriceMessage(msg)
		}
	}
}

func (p *PriceStore) handlePriceMessage(msg transport.ReceivedMessage) {
	if msg.Error != nil {
		p.log.
			WithError(msg.Error).
			Error("Unable to read prices from the transport layer")
		return
	}
	price, ok := msg.Message.(*messages.Price)
	if !ok {
		p.log.
			WithFields(msg.Fields()).
			Error("Unexpected value returned from the transport layer")
		return
	}
	err := p.collectPrice(price)
	if err != nil {
		p.log.
			WithError(err).
			WithFields(price.Price.Fields(p.recover)).
			WithField("version", price.Version).
			WithFields(msg.Fields()).
			Warn("Price rejected")
	} else {
		p.log.
			WithFields(price.Price.Fields(p.recover)).
			WithField("version", price.Version).
			WithFields(msg.Fields()).
			Info("Price collected")
	}
}

// contextCancelHandler handles context cancellation.
func (p *PriceStore) contextCancelHandler() {
	defer func() { close(p.waitCh) }()
	defer p.log.Debug("Stopped")
	<-p.ctx.Done()
}
