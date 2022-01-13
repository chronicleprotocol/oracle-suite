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
	"fmt"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/internal/gofer/marshal"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "GHOST"

type ErrUnableToFindAsset struct {
	AssetName string
}

func (e ErrUnableToFindAsset) Error() string {
	return fmt.Sprintf("unable to find the %s in Gofer price models", e.AssetName)
}

type Ghost struct {
	ctx    context.Context
	waitCh chan error

	gofer      gofer.Gofer
	signer     ethereum.Signer
	transport  transport.Transport
	interval   time.Duration
	pairs      []string
	goferPairs map[gofer.Pair]string
	log        log.Logger
}

type Config struct {
	// Gofer is an instance of the gofer.Gofer which will be used to fetch
	// prices.
	Gofer gofer.Gofer
	// Signer is an instance of the ethereum.Signer which will be used to
	// sign prices.
	Signer ethereum.Signer
	// Transport is a implementation of transport used to send prices to
	// relayers.
	Transport transport.Transport
	// Interval describes how often we should send prices to the network.
	Interval time.Duration
	// Logger is a current logger interface used by the Ghost. The Logger
	// helps to monitor asynchronous processes.
	Logger log.Logger
	// Pairs is a list supported pairs.
	Pairs []string
}

func NewGhost(ctx context.Context, cfg Config) (*Ghost, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	g := &Ghost{
		ctx:        ctx,
		waitCh:     make(chan error),
		gofer:      cfg.Gofer,
		signer:     cfg.Signer,
		transport:  cfg.Transport,
		interval:   cfg.Interval,
		pairs:      cfg.Pairs,
		goferPairs: make(map[gofer.Pair]string),
		log:        cfg.Logger.WithField("tag", LoggerTag),
	}
	return g, nil
}

func (g *Ghost) Start() error {
	g.log.Infof("Starting")

	// Unfortunately, the Gofer stores pairs in the AAA/BBB format but Ghost
	// (and oracle contract) stores them in AAABBB format. Because of this we
	// need to make this wired mapping:
	for _, pair := range g.pairs {
		goferPairs, err := g.gofer.Pairs()
		if err != nil {
			return err
		}
		found := false
		for _, goferPair := range goferPairs {
			if goferPair.Base+goferPair.Quote == pair {
				g.goferPairs[goferPair] = pair
				found = true
				break
			}
		}
		if !found {
			return ErrUnableToFindAsset{AssetName: pair}
		}
	}

	err := g.broadcasterLoop()
	if err != nil {
		return err
	}

	go g.contextCancelHandler()
	return nil
}

func (g *Ghost) Wait() chan error {
	return g.waitCh
}

// broadcast sends price for single pair to the network. This method uses
// current price from the Gofer so it must be updated beforehand.
func (g *Ghost) broadcast(goferPair gofer.Pair) error {
	var err error

	pair := g.goferPairs[goferPair]
	tick, err := g.gofer.Price(goferPair)
	if err != nil {
		return err
	}
	if tick.Error != "" {
		return errors.New(tick.Error)
	}

	// Create price:
	price := &oracle.Price{Wat: pair, Age: tick.Time}
	price.SetFloat64Price(tick.Price)

	// Sign price:
	err = price.Sign(g.signer)
	if err != nil {
		return err
	}

	// Broadcast price to P2P network:
	message, err := createPriceMessage(price, tick)
	if err != nil {
		return err
	}
	err = g.transport.Broadcast(messages.PriceMessageName, message)
	if err != nil {
		return err
	}

	return err
}

// broadcasterLoop creates a asynchronous loop which fetches prices from exchanges and then
// sends them to the network at a specified interval.
func (g *Ghost) broadcasterLoop() error {
	if g.interval == 0 {
		return nil
	}

	ticker := time.NewTicker(g.interval)
	wg := sync.WaitGroup{}
	go func() {
		for {
			select {
			case <-g.waitCh:
				ticker.Stop()
				return
			case <-ticker.C:
				// TODO: fetch all prices before broadcast is called

				// Send prices to the network:
				//
				// Signing may be slow, especially with high KDF so this is why
				// we're using goroutines here.
				wg.Add(1)
				go func() {
					for assetPair := range g.goferPairs {
						err := g.broadcast(assetPair)
						if err != nil {
							g.log.
								WithFields(log.Fields{"assetPair": assetPair}).
								WithError(err).
								Warn("Unable to broadcast price")
						} else {
							g.log.
								WithFields(log.Fields{"assetPair": assetPair}).
								Info("Price broadcast")
						}
					}
					wg.Done()
				}()
			}
			wg.Wait()
		}
	}()

	return nil
}

func (g *Ghost) contextCancelHandler() {
	defer func() { g.waitCh <- nil }()
	defer g.log.Info("Stopped")
	<-g.ctx.Done()
}

func createPriceMessage(op *oracle.Price, gp *gofer.Price) (*messages.Price, error) {
	trace, err := marshal.Marshall(marshal.JSON, gp)
	if err != nil {
		return nil, err
	}

	return &messages.Price{
		Price: op,
		Trace: trace,
	}, nil
}
