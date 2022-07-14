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

package spectre

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
)

const LoggerTag = "SPECTRE"

type errNotEnoughPricesForQuorum struct {
	AssetPair string
}

func (e errNotEnoughPricesForQuorum) Error() string {
	return fmt.Sprintf(
		"unable to update the Oracle for %s pair, there is not enough prices to achieve a quorum",
		e.AssetPair,
	)
}

type errUnknownAsset struct {
	AssetPair string
}

func (e errUnknownAsset) Error() string {
	return fmt.Sprintf("pair %s does not exists", e.AssetPair)
}

type errNoPrices struct {
	AssetPair string
}

func (e errNoPrices) Error() string {
	return fmt.Sprintf("there is no prices in the priceStore for %s pair", e.AssetPair)
}

type Spectre struct {
	ctx    context.Context
	mu     sync.Mutex
	waitCh chan error

	signer     ethereum.Signer
	priceStore *store.PriceStore
	interval   time.Duration
	log        log.Logger
	pairs      map[string]*Pair
}

// Config is the configuration for Spectre.
type Config struct {
	Signer ethereum.Signer
	// PriceStore provides prices for Spectre.
	PriceStore *store.PriceStore
	// Interval describes how often we should try to update Oracles.
	Interval time.Duration
	// Pairs is the list supported pairs by Spectre with their configuration.
	Pairs []*Pair
	// Logger is a current logger interface used by the Spectre. The Logger is
	// required to monitor asynchronous processes.
	Logger log.Logger
}

type Pair struct {
	// AssetPair is the name of asset pair, e.g. ETHUSD.
	AssetPair string
	// OracleSpread is the minimum spread between the Oracle price and new price
	// required to send update.
	OracleSpread float64
	// OracleExpiration is the minimum time difference between the Oracle time
	// and current time required to send an update.
	OracleExpiration time.Duration
	// PriceExpiration is the maximum amount of time before price received
	// from the feeder will be considered as expired.
	PriceExpiration time.Duration
	// Median is the instance of the oracle.Median which is the interface for
	// the Oracle contract.
	Median oracle.Median
}

func NewSpectre(cfg Config) (*Spectre, error) {
	if cfg.Signer == nil {
		return nil, errors.New("signer must not be nil")
	}
	if cfg.PriceStore == nil {
		return nil, errors.New("price store must not be nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	r := &Spectre{
		waitCh:     make(chan error),
		signer:     cfg.Signer,
		priceStore: cfg.PriceStore,
		interval:   cfg.Interval,
		pairs:      make(map[string]*Pair),
		log:        cfg.Logger.WithField("tag", LoggerTag),
	}
	for _, p := range cfg.Pairs {
		r.pairs[p.AssetPair] = p
	}
	return r, nil
}

func (s *Spectre) Start(ctx context.Context) error {
	if s.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	s.log.Info("Starting")
	s.ctx = ctx
	go s.contextCancelHandler()
	s.relayerLoop()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (s *Spectre) Wait() chan error {
	return s.waitCh
}

// relay tries to update an Oracle contract for given pair. It'll return
// transaction hash or nil if there is no need to update Oracle.
func (s *Spectre) relay(assetPair string) (*ethereum.Hash, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pair, ok := s.pairs[assetPair]
	if !ok {
		return nil, errUnknownAsset{AssetPair: assetPair}
	}

	pricesSlice, err := s.priceStore.GetByAssetPair(context.Background(), assetPair)
	if err != nil {
		return nil, err
	}

	pricesList := newPricesList(pricesSlice)
	if pricesList == nil || pricesList.len() == 0 {
		return nil, errNoPrices{AssetPair: assetPair}
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

	// Clear expired prices:
	pricesList.clearOlderThan(time.Now().Add(-1 * pair.PriceExpiration))
	pricesList.clearOlderThan(oracleTime)

	// Use only a minimum prices required to achieve a quorum:
	pricesList.truncate(oracleQuorum)

	spread := pricesList.spread(oraclePrice)
	isExpired := oracleTime.Add(pair.OracleExpiration).Before(time.Now())
	isStale := spread >= pair.OracleSpread

	// Print logs:
	s.log.
		WithFields(log.Fields{
			"assetPair":        assetPair,
			"bar":              oracleQuorum,
			"age":              oracleTime.String(),
			"val":              oraclePrice.String(),
			"expired":          isExpired,
			"stale":            isStale,
			"oracleExpiration": pair.OracleExpiration.String(),
			"oracleSpread":     pair.OracleSpread,
			"timeToExpiration": time.Since(oracleTime).String(),
			"currentSpread":    spread,
		}).
		Debug("Trying to update Oracle")
	for _, price := range pricesList.oraclePrices() {
		s.log.
			WithFields(price.Fields(s.signer)).
			Debug("Feed")
	}

	if isExpired || isStale {
		// Check if there are enough prices to achieve a quorum:
		if int64(pricesList.len()) != oracleQuorum {
			return nil, errNotEnoughPricesForQuorum{AssetPair: assetPair}
		}

		// Send *actual* transaction to the Ethereum network:
		tx, err := pair.Median.Poke(s.ctx, pricesList.oraclePrices(), true)
		return tx, err
	}

	// There is no need to update Oracle:
	return nil, nil
}

// relayerLoop creates a asynchronous loop which tries to send an update
// to an Oracle contract at a specified interval.
func (s *Spectre) relayerLoop() {
	if s.interval == 0 {
		return
	}

	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				for assetPair := range s.pairs {
					tx, err := s.relay(assetPair)

					// Print log in case of an error:
					if err != nil {
						s.log.
							WithFields(log.Fields{"assetPair": assetPair}).
							WithError(err).
							Warn("Unable to update Oracle")
					}
					// Print log if there was no need to update prices:
					if err == nil && tx == nil {
						s.log.
							WithFields(log.Fields{"assetPair": assetPair}).
							Info("Oracle price is still valid")
					}
					// Print log if Oracle update transaction was sent:
					if tx != nil {
						s.log.
							WithFields(log.Fields{"assetPair": assetPair, "tx": tx.String()}).
							Info("Oracle updated")
					}
				}
			}
		}
	}()
}

func (s *Spectre) contextCancelHandler() {
	defer func() { close(s.waitCh) }()
	defer s.log.Info("Stopped")
	<-s.ctx.Done()
}
