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

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/nodes"
)

// AsyncGofer implements the gofer.Gofer interface. It works just like Graph
// but allows to update prices asynchronously.
type AsyncGofer struct {
	*Gofer
	ctx    context.Context
	waitCh chan error
	feeder *feeder.Feeder
}

// NewAsyncGofer returns a new AsyncGofer instance.
func NewAsyncGofer(ctx context.Context, g map[gofer.Pair]nodes.Aggregator, f *feeder.Feeder) (*AsyncGofer, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	return &AsyncGofer{
		Gofer:  NewGofer(g, nil),
		ctx:    ctx,
		feeder: f,
		waitCh: make(chan error),
	}, nil
}

// Start starts asynchronous price updater.
func (a *AsyncGofer) Start() error {
	go a.contextCancelHandler()
	ns, _ := a.findNodes()
	return a.feeder.Start(ns...)
}

// Wait waits until the context is canceled or until an error occurs.
func (a *AsyncGofer) Wait() chan error {
	return a.waitCh
}

func (a *AsyncGofer) contextCancelHandler() {
	defer func() { a.waitCh <- nil }()
	<-a.ctx.Done()
}
