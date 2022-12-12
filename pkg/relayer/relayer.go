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

package relayer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "RELAYER"

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

type Relayer struct {
	ctx    context.Context
	mu     sync.Mutex
	waitCh chan error

	signer     ethereum.Signer
	priceStore *store.PriceStore
	interval   time.Duration
	log        log.Logger
	pairs      map[string]*Pair
}

// Config is the configuration for Relayer.
type Config struct {
	// Signer is the signer which will be used to sign the Oracle update transactions.
	Signer ethereum.Signer

	// PriceStore provides prices for Relayer.
	PriceStore *store.PriceStore

	// Interval describes how often we should try to update Oracles.
	Interval time.Duration

	// Pairs is the list supported pairs by Relayer with their configuration.
	Pairs []*Pair

	// Logger is a current logger interface used by the Relayer. The Logger is
	// required to monitor asynchronous processes.
	Logger log.Logger
}

type Pair struct {
	// AssetPair is the name of asset pair, e.g. ETHUSD.
	AssetPair string

	// OracleSpread is the minimum calcSpread between the Oracle price and new
	// price required to send update.
	OracleSpread float64

	// OracleExpiration is the minimum time difference between the last Oracle
	// update on the Medianizer contract and current time required to send
	// update.
	OracleExpiration time.Duration

	// Median is the instance of the oracle.Median which is the interface for
	// the Medianizer contract.
	Median oracle.Median
}

func New(cfg Config) (*Relayer, error) {
	if cfg.Signer == nil {
		return nil, errors.New("signer must not be nil")
	}
	if cfg.PriceStore == nil {
		return nil, errors.New("price store must not be nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	r := &Relayer{
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

func (s *Relayer) Start(ctx context.Context) error {
	if s.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	s.log.Info("Starting")
	s.ctx = ctx
	go s.relayerRoutine()
	go s.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (s *Relayer) Wait() chan error {
	return s.waitCh
}

// relay tries to update an Oracle contract for given pair. It'll return
// transaction hash or nil if there is no need to update Oracle.
func (s *Relayer) relay(assetPair string) (*ethereum.Hash, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pair, ok := s.pairs[assetPair]
	if !ok {
		return nil, errUnknownAsset{AssetPair: assetPair}
	}
	prices, err := s.priceStore.GetByAssetPair(s.ctx, assetPair)
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

	// Use only a minimum prices required to achieve a quorum.
	// Using a different number of prices that specified in the bar field cause
	// the transaction to fail.
	truncate(&prices, oracleQuorum)

	// Check if price on the Medianizer contract needs to be updated.
	// The price needs to be updated if:
	// - price is older than the OracleExpiration.
	// - price differs from the current price by more than the OracleSpread.
	spread := calcSpread(&prices, oraclePrice)
	isExpired := oracleTime.Add(pair.OracleExpiration).Before(time.Now())
	isStale := spread >= pair.OracleSpread

	// Print logs.
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
	for _, price := range messagesToPrices(&prices) {
		s.log.
			WithFields(price.Fields(s.signer)).
			Debug("Feed")
	}

	// If price is stale or expired, send update.
	if isExpired || isStale {
		// Check if there are enough prices to achieve a quorum.
		if int64(len(prices)) != oracleQuorum {
			return nil, errNotEnoughPricesForQuorum{AssetPair: assetPair}
		}

		// Send *actual* transaction to the Ethereum network.
		tx, err := pair.Median.Poke(s.ctx, messagesToPrices(&prices), true)
		return tx, err
	}

	// There is no need to update Oracle.
	return nil, nil
}

// relayerRoutine creates an asynchronous loop that tries to send an update
// to an Oracle contract at a specified interval.
func (s *Relayer) relayerRoutine() {
	if s.interval == 0 {
		return
	}
	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-s.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			for assetPair := range s.pairs {
				tx, err := s.relay(assetPair)

				// Print log in case of an error.
				if err != nil {
					s.log.
						WithFields(log.Fields{"assetPair": assetPair}).
						WithError(err).
						Warn("Unable to update Oracle")
				}
				// Print log if there was no need to update prices.
				if err == nil && tx == nil {
					s.log.
						WithFields(log.Fields{"assetPair": assetPair}).
						Info("Oracle price is still valid")
				}
				// Print log if Oracle update transaction was sent.
				if tx != nil {
					s.log.
						WithFields(log.Fields{"assetPair": assetPair, "tx": tx.String()}).
						Info("Oracle updated")
				}
			}
		}
	}
}

func (s *Relayer) contextCancelHandler() {
	defer func() { close(s.waitCh) }()
	defer s.log.Info("Stopped")
	<-s.ctx.Done()
}

// messagesToPrices returns oracle prices.
func messagesToPrices(p *[]*messages.Price) []*oracle.Price {
	var prices []*oracle.Price
	for _, price := range *p {
		prices = append(prices, price.Price)
	}
	return prices
}

// truncate removes random prices until the number of remaining prices is equal
// to n. If the number of prices is less or equal to n, it does nothing.
//
// This method is used to reduce number of arguments in transaction which will
// reduce transaction costs.
func truncate(p *[]*messages.Price, n int64) {
	if int64(len(*p)) <= n {
		return
	}
	rand.Shuffle(len(*p), func(i, j int) {
		(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
	})
	*p = (*p)[0:n]
}

// calcMedian calculates the calcMedian price for all messages in the list.
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

// calcSpread calculates the calcSpread between given price and a calcMedian price.
// The calcSpread is returned as percentage points.
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
