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

package relayer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const LoggerTag = "RELAYER"

// Relayer is a service that relays prices to the Medianizer contracts.
// TODO(mdobak): Rename to Relay.
type Relayer struct {
	mu     sync.Mutex
	ctx    context.Context
	waitCh chan error

	store   *store.PriceStore
	ticker  *timeutil.Ticker
	pairs   map[string]*Pair
	log     log.Logger
	recover crypto.Recoverer
}

// Config is the configuration for Relayer.
type Config struct {
	// PriceStore is the price store which will be used to get the latest
	// prices.
	PriceStore *store.PriceStore

	// PokeTicker invokes the Relayer routine that relays prices to the Medianizer
	// contracts.
	PokeTicker *timeutil.Ticker

	// Pairs is the list supported pairs by Relayer with their configuration.
	Pairs []*Pair

	// Logger is a current logger interface used by the Relayer.
	Logger log.Logger

	// Recoverer provides a method to recover the public key from a signature.
	// The default is crypto.ECRecoverer.
	Recoverer crypto.Recoverer
}

type Pair struct {
	// AssetPair is the name of asset pair, e.g. ETHUSD.
	AssetPair string

	// Spread is the minimum calcSpread between the Oracle price and new
	// price required to send update.
	Spread float64

	// Expiration is the minimum time difference between the last Oracle
	// update on the Medianizer contract and current time required to send
	// update.
	Expiration time.Duration

	// Median is the instance of the oracle.Median which is the interface for
	// the Medianizer contract.
	Median median.Median

	// FeedAddresses is the list of addresses which are allowed to send
	// updates to the Medianizer contract.
	FeedAddresses []types.Address

	// FeedAddressesUpdateTicker invokes the FeedAddresses update routine
	// when ticked.
	//
	// TODO(mdobak): Instead of updating the list periodically, we should
	//               listen for events from the Medianizer contract.
	FeedAddressesUpdateTicker *timeutil.Ticker
}

func New(cfg Config) (*Relayer, error) {
	if cfg.PriceStore == nil {
		return nil, errors.New("price store must not be nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	if cfg.Recoverer == nil {
		cfg.Recoverer = crypto.ECRecoverer
	}
	r := &Relayer{
		waitCh:  make(chan error),
		store:   cfg.PriceStore,
		ticker:  cfg.PokeTicker,
		pairs:   make(map[string]*Pair, len(cfg.Pairs)),
		log:     cfg.Logger.WithField("tag", LoggerTag),
		recover: cfg.Recoverer,
	}
	for _, p := range cfg.Pairs {
		r.pairs[p.AssetPair] = p
	}
	return r, nil
}

// Start implements the service.Service interface.
func (s *Relayer) Start(ctx context.Context) error {
	if s.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	s.log.Debug("Starting")
	s.ctx = ctx
	for _, p := range s.pairs {
		if err := s.syncFeedAddresses(p); err != nil {
			return err
		}
		p.FeedAddressesUpdateTicker.Start(ctx)
		go s.syncFeedAddressesRoutine(p)
	}
	s.ticker.Start(s.ctx)
	go s.relayerRoutine()
	go s.contextCancelHandler()
	return nil
}

// Wait implements the service.Service interface.
func (s *Relayer) Wait() <-chan error {
	return s.waitCh
}

// relay tries to update an Oracle contract for given pair.
// In returns a transaction hash if the update was successful.
// If update is not required, it returns nil.
func (s *Relayer) relay(assetPair string) (*types.Hash, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pair, ok := s.pairs[assetPair]
	if !ok {
		return nil, fmt.Errorf("unknown asset pair: %s", assetPair)
	}
	prices, err := s.store.GetByAssetPair(s.ctx, assetPair)
	if err != nil {
		return nil, err
	}
	oracleQuorum, err := pair.Median.Bar(s.ctx)
	if err != nil {
		return nil, err
	}
	oracleTime, err := pair.Median.Age(s.ctx)
	if err != nil {
		return nil, err
	}
	oraclePrice, err := pair.Median.Val(s.ctx)
	if err != nil {
		return nil, err
	}

	// Clear expired prices.
	clearOlderThan(&prices, oracleTime)

	// Remove prices from addresses outside the FeedAddresses list.
	filterAddresses(&prices, pair.FeedAddresses, s.recover)

	// Use only a minimum prices required to achieve a quorum.
	// Using a different number of prices that specified in the bar field cause
	// the transaction to fail.
	truncate(&prices, oracleQuorum)

	// Check if price on the Medianizer contract needs to be updated.
	// The price needs to be updated if:
	// - Price is older than the interval specified in the OracleExpiration
	//   field.
	// - Price differs from the current price by more than is specified in the
	//   OracleSpread field.
	spread := calcSpread(&prices, oraclePrice)
	isExpired := oracleTime.Add(pair.Expiration).Before(time.Now())
	isStale := spread >= pair.Spread

	// Print logs.
	s.log.
		WithFields(log.Fields{
			"assetPair":        assetPair,
			"bar":              oracleQuorum,
			"age":              oracleTime.String(),
			"val":              oraclePrice.String(),
			"expired":          isExpired,
			"stale":            isStale,
			"oracleExpiration": pair.Expiration.String(),
			"oracleSpread":     pair.Spread,
			"timeToExpiration": time.Since(oracleTime).String(),
			"currentSpread":    spread,
		}).
		Debug("Trying to update Oracle")
	for _, price := range prices {
		s.log.
			WithFields(price.Price.Fields(s.recover)).
			Debug("Feed Price for Oracle update")
	}

	// If price is stale or expired, send update.
	if isExpired || isStale {
		// Check if there are enough prices to achieve a quorum.
		if int64(len(prices)) != oracleQuorum {
			return nil, fmt.Errorf("not enough prices to achieve quorum: %d/%d", len(prices), oracleQuorum)
		}

		// Send *actual* transaction.
		return pair.Median.Poke(s.ctx, toOraclePrices(&prices), true)
	}

	// There is no need to update the price.
	return nil, nil
}

func (s *Relayer) syncFeedAddresses(p *Pair) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get list of addresses from the contract.
	addresses, err := p.Median.Feeds(s.ctx)
	if err != nil {
		return err
	}

	// Update the list.
	p.FeedAddresses = []types.Address{}
	for _, addr := range addresses {
		if addr.IsZero() {
			continue
		}
		s.log.
			WithField("symbol", p.AssetPair).
			WithField("feed", addr.String()).
			WithField("oracle", p.Median.Address().String()).
			Info("Oracle Feed")
		p.FeedAddresses = append(p.FeedAddresses, addr)
	}
	return nil
}

func (s *Relayer) relayerRoutine() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.TickCh():
			for assetPair := range s.pairs {
				tx, err := s.relay(assetPair)

				// Print log in case of an error.
				if err != nil {
					s.log.
						WithField("assetPair", assetPair).
						WithError(err).
						Warn("Unable to update Oracle")
				}

				// Print log if there was no need to update prices.
				if err == nil && tx == nil {
					s.log.
						WithField("assetPair", assetPair).
						Info("Oracle price is still valid")
				}

				// Print log if Oracle update transaction was sent.
				if tx != nil {
					s.log.
						WithFields(log.Fields{
							"assetPair": assetPair,
							"tx":        tx.String(),
						}).
						Info("Oracle updated")
				}
			}
		}
	}
}

func (s *Relayer) syncFeedAddressesRoutine(p *Pair) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-p.FeedAddressesUpdateTicker.TickCh():
			if err := s.syncFeedAddresses(p); err != nil {
				s.log.
					WithField("assetPair", p.AssetPair).
					WithError(err).
					Warn("Unable to sync feed addresses")
			}
		}
	}
}

func (s *Relayer) contextCancelHandler() {
	defer func() { close(s.waitCh) }()
	defer s.log.Debug("Stopped")
	<-s.ctx.Done()
}

// toOraclePrices returns a slice of oracle.Prices from price messages.
func toOraclePrices(p *[]*messages.Price) []*median.Price {
	var prices []*median.Price
	for _, price := range *p {
		prices = append(prices, price.Price)
	}
	return prices
}

// truncate removes random prices until the number of remaining prices is equal
// to n.
func truncate(p *[]*messages.Price, n int64) {
	if int64(len(*p)) <= n {
		return
	}
	rand.Shuffle(len(*p), func(i, j int) {
		(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
	})
	*p = (*p)[0:n]
}

// filterAddresses removes all prices from the slice that are not signed by
// addresses from the list.
func filterAddresses(p *[]*messages.Price, addrs []types.Address, r crypto.Recoverer) {
	var prices []*messages.Price
	for _, price := range *p {
		feedAddr, err := price.Price.From(r)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if bytes.Equal(feedAddr.Bytes(), addr.Bytes()) {
				prices = append(prices, price)
				break
			}
		}
	}
	*p = prices
}

// clearOlderThan deletes messages which are older than given time.
func clearOlderThan(p *[]*messages.Price, t time.Time) {
	var prices []*messages.Price
	for _, price := range *p {
		if !price.Price.Age.Before(t) {
			prices = append(prices, price)
		}
	}
	*p = prices
}

// calcMedian calculates the median price.
func calcMedian(p *[]*messages.Price) *big.Int {
	count := len(*p)
	if count == 0 {
		return big.NewInt(0)
	}
	sort.Slice(*p, func(i, j int) bool {
		return (*p)[i].Price.Val.Cmp((*p)[j].Price.Val) < 0
	})
	if count%2 == 0 {
		m := count / 2
		x1 := (*p)[m-1].Price.Val
		x2 := (*p)[m].Price.Val
		return new(big.Int).Div(new(big.Int).Add(x1, x2), big.NewInt(2))
	}
	return (*p)[(count-1)/2].Price.Val
}

// calcSpread calculates the spread between given price and a median price.
// The spread is returned as percentage points.
func calcSpread(p *[]*messages.Price, price *big.Int) float64 {
	if len(*p) == 0 || price.Cmp(big.NewInt(0)) == 0 {
		return math.Inf(1)
	}
	oldPriceF := new(big.Float).SetInt(price)
	newPriceF := new(big.Float).SetInt(calcMedian(p))
	x := new(big.Float).Sub(newPriceF, oldPriceF)
	x = new(big.Float).Quo(x, oldPriceF)
	x = new(big.Float).Mul(x, big.NewFloat(100))
	xf, _ := x.Float64()
	return math.Abs(xf)
}
