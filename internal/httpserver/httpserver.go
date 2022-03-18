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

package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

const shutdownTimeout = 1 * time.Second

type Middleware interface {
	Handle(http.Handler) http.Handler
}

type MiddlewareFunc func(http.Handler) http.Handler

func (m MiddlewareFunc) Handle(h http.Handler) http.Handler {
	return m(h)
}

// HTTPServer allows using middlewares with http.Server and allow controlling
// server lifecycle using context.
type HTTPServer struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	serveCh   chan error
	waitCh    chan error

	ln  net.Listener
	srv *http.Server

	handler        http.Handler
	wrappedHandler http.Handler
	middlewares    []Middleware
}

// New creates a new HTTPServer instance.
func New(srv *http.Server) *HTTPServer {
	s := &HTTPServer{
		serveCh: make(chan error),
		waitCh:  make(chan error),
		srv:     srv,
	}
	s.handler = srv.Handler
	srv.Handler = http.HandlerFunc(s.ServeHTTP)
	return s
}

// Use adds a middleware. Middlewares will be called in the order in which they
// were added. This function will panic after calling ServerHTTP/Start.
func (s *HTTPServer) Use(m ...Middleware) {
	if s.wrappedHandler != nil {
		panic("cannot add a middleware after calling ServeHTTP/Start")
	}
	s.middlewares = append(s.middlewares, m...)
}

// ServeHTTP prepares middlewares stack if necessary and calls ServerHTTP
// on the wrapped server.
func (s *HTTPServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if s.wrappedHandler == nil {
		if len(s.middlewares) == 0 {
			s.wrappedHandler = s.handler
		} else {
			h := s.middlewares[len(s.middlewares)-1].Handle(s.handler)
			for i := len(s.middlewares) - 2; i >= 0; i-- {
				h = s.middlewares[i].Handle(h)
			}
			s.wrappedHandler = h
		}
	}
	s.wrappedHandler.ServeHTTP(rw, r)
}

// Start starts HTTP server.
func (s *HTTPServer) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	s.ctx, s.ctxCancel = context.WithCancel(ctx)
	addr := s.srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := (&net.ListenConfig{}).Listen(s.ctx, "tcp", addr)
	if err != nil {
		return err
	}
	s.ln = ln
	go s.shutdownHandler()
	go s.serve()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (s *HTTPServer) Wait() chan error {
	return s.waitCh
}

// Addr returns the server's network address.
func (s *HTTPServer) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *HTTPServer) serve() {
	s.serveCh <- s.srv.Serve(s.ln)
}

func (s *HTTPServer) shutdownHandler() {
	defer func() { close(s.waitCh) }()
	select {
	case <-s.ctx.Done():
		ctx, ctxCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer ctxCancel()
		s.waitCh <- s.srv.Shutdown(ctx)
	case err := <-s.serveCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.waitCh <- err
		}
	}
}
