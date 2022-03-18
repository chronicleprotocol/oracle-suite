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
func NewAsyncGofer(g map[gofer.Pair]nodes.Aggregator, f *feeder.Feeder) (*AsyncGofer, error) {
	return &AsyncGofer{
		Gofer:  NewGofer(g, nil),
		feeder: f,
		waitCh: make(chan error),
	}, nil
}

// Start starts asynchronous price updater.
func (a *AsyncGofer) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	a.ctx = ctx
	go a.contextCancelHandler()
	return a.feeder.Start(ctx)
}

// Wait waits until the context is canceled or until an error occurs.
func (a *AsyncGofer) Wait() chan error {
	return a.waitCh
}

func (a *AsyncGofer) contextCancelHandler() {
	defer func() { close(a.waitCh) }()
	<-a.ctx.Done()
}
