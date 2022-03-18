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

package rpc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const AgentLoggerTag = "GOFER_AGENT"

type AgentConfig struct {
	// Gofer instance which will be used by the agent. If this instance
	// implements the gofer.StartableGofer interface, the Start and Stop
	// methods are called whenever corresponding Agent's Start and
	// Stop are called.
	Gofer gofer.Gofer
	// Network is used for the rpc.Listener function.
	Network string
	// Address is used for the rpc.Listener function.
	Address string
	Logger  log.Logger
}

// Agent creates and manages an RPC server for remote Gofer calls.
type Agent struct {
	ctx    context.Context
	waitCh chan error

	api      *API
	rpc      *rpc.Server
	listener net.Listener
	network  string
	address  string
	log      log.Logger
}

// NewAgent returns a new Agent instance.
func NewAgent(cfg AgentConfig) (*Agent, error) {
	server := &Agent{
		waitCh: make(chan error),
		api: &API{
			gofer: cfg.Gofer,
			log:   cfg.Logger.WithField("tag", AgentLoggerTag),
		},
		rpc:     rpc.NewServer(),
		network: cfg.Network,
		address: cfg.Address,
		log:     cfg.Logger.WithField("tag", AgentLoggerTag),
	}

	err := server.rpc.Register(server.api)
	if err != nil {
		return nil, err
	}
	server.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	return server, nil
}

// Start starts the RPC server.
func (s *Agent) Start(ctx context.Context) error {
	s.log.Infof("Starting")
	var err error

	if ctx == nil {
		return errors.New("context must not be nil")
	}
	s.ctx = ctx

	// Start RPC server:
	s.listener, err = net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	go func() {
		err := http.Serve(s.listener, nil)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.WithError(err).Error("RPC server crashed")
		}
	}()

	go s.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (s *Agent) Wait() chan error {
	return s.waitCh
}

func (s *Agent) contextCancelHandler() {
	defer s.log.Info("Stopped")
	<-s.ctx.Done()
	s.waitCh <- s.listener.Close()
}
