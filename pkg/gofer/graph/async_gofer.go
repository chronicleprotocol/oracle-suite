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

package graph

import (
	"context"
	"errors"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/nodes"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const LoggerTag = "ASYNC_GOFER"

// AsyncGofer implements the gofer.Gofer interface. It works just like Gofer
// but allows updating prices asynchronously.
type AsyncGofer struct {
	*Gofer
	ctx    context.Context
	waitCh chan error
	feeder *feeder.Feeder
	nodes  []nodes.Node
	log    log.Logger
}

// NewAsyncGofer returns a new AsyncGofer instance.
func NewAsyncGofer(
	graph map[gofer.Pair]nodes.Aggregator,
	feeder *feeder.Feeder,
	nodes []nodes.Node,
	logger log.Logger,
) (*AsyncGofer, error) {

	return &AsyncGofer{
		Gofer:  NewGofer(graph, nil),
		waitCh: make(chan error),
		feeder: feeder,
		nodes:  nodes,
		log:    logger.WithField("tag", LoggerTag),
	}, nil
}

// Start starts asynchronous price updater.
func (a *AsyncGofer) Start(ctx context.Context) error {
	a.log.Infof("Starting")

	if ctx == nil {
		return errors.New("context must not be nil")
	}
	a.ctx = ctx

	gcdTTL := getGCDTTL(a.nodes)
	if gcdTTL < time.Second {
		gcdTTL = time.Second
	}
	a.log.WithField("interval", gcdTTL.String()).Infof("Update interval (GCD of all TTLs)")

	feed := func() {
		// We have to add gcdTTL to the current time because we want
		// to find all nodes that will expire before the next tick.
		t := time.Now().Add(gcdTTL)
		warns := a.feeder.Feed(a.nodes, t)
		if len(warns.List) > 0 {
			a.log.WithError(warns.ToError()).Warn("Unable to feed some nodes")
		}
	}

	ticker := time.NewTicker(gcdTTL)
	go func() {
		feed()
		for {
			select {
			case <-a.ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				feed()
			}
		}
	}()

	go a.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (a *AsyncGofer) Wait() chan error {
	return a.waitCh
}

func (a *AsyncGofer) contextCancelHandler() {
	defer func() { close(a.waitCh) }()
	defer a.log.Info("Stopped")
	<-a.ctx.Done()
}

// getGCDTTL returns the greatest common divisor of nodes minTTLs.
func getGCDTTL(ns []nodes.Node) time.Duration {
	ttl := time.Duration(0)
	nodes.Walk(func(n nodes.Node) {
		if f, ok := n.(feeder.Feedable); ok {
			if ttl == 0 {
				ttl = f.MinTTL()
			}
			a := ttl
			b := f.MinTTL()
			for b != 0 {
				t := b
				b = a % b
				a = t
			}
			ttl = a
		}
	}, ns...)
	return ttl
}
