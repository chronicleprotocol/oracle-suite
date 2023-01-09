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

package middleware

import (
	"context"
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

type BroadcastFunc func(topic string, message transport.Message) error

type BroadcastMiddleware interface {
	Broadcast(ctx context.Context, next BroadcastFunc) BroadcastFunc
}

type BroadcastMiddlewareFunc func(ctx context.Context, next BroadcastFunc) BroadcastFunc

func (m BroadcastMiddlewareFunc) Broadcast(ctx context.Context, next BroadcastFunc) BroadcastFunc {
	return m(ctx, next)
}

// Middleware is a transport implementation that allows to add middleware to
// the broadcast function.
type Middleware struct {
	ctx context.Context
	mu  sync.RWMutex

	t transport.Transport
	m []BroadcastMiddleware
	b BroadcastFunc
}

// New creates a new Middleware instance.
func New(t transport.Transport) *Middleware {
	return &Middleware{t: t, b: t.Broadcast}
}

// Use adds a middleware to the broadcast function.
func (m *Middleware) Use(bm ...BroadcastMiddleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m = append(m.m, bm...)
	m.b = m.t.Broadcast
	for i := len(m.m) - 1; i >= 0; i-- {
		m.b = m.m[i].Broadcast(m.ctx, m.b)
	}
}

// Broadcast implements the transport.Transport interface.
func (m *Middleware) Broadcast(topic string, message transport.Message) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.b(topic, message)
}

// Messages implements the transport.Transport interface.
func (m *Middleware) Messages(topic string) <-chan transport.ReceivedMessage {
	return m.t.Messages(topic)
}

// Start implements the transport.Transport interface.
func (m *Middleware) Start(ctx context.Context) error {
	if m.ctx == nil {
		m.ctx = ctx
	}
	// There is no need to check if Middleware is already started, because
	// underlying transport should do that.
	return m.t.Start(ctx)
}

// Wait implements the transport.Transport interface.
func (m *Middleware) Wait() <-chan error {
	return m.t.Wait()
}
