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

package spire

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/datastore"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

const AgentLoggerTag = "SPIRE_AGENT"

type Agent struct {
	ctx    context.Context
	doneCh chan struct{}

	api      *API
	rpc      *rpc.Server
	listener net.Listener
	network  string
	address  string
	log      log.Logger
}

type AgentConfig struct {
	Datastore datastore.Datastore
	Transport transport.Transport
	Signer    ethereum.Signer
	Network   string
	Address   string
	Logger    log.Logger
}

func NewAgent(ctx context.Context, cfg AgentConfig) (*Agent, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	s := &Agent{
		ctx:    ctx,
		doneCh: make(chan struct{}),
		api: &API{
			datastore: cfg.Datastore,
			transport: cfg.Transport,
			signer:    cfg.Signer,
			log:       cfg.Logger.WithField("tag", AgentLoggerTag),
		},
		rpc:     rpc.NewServer(),
		network: cfg.Network,
		address: cfg.Address,
		log:     cfg.Logger.WithField("tag", AgentLoggerTag),
	}
	err := s.rpc.Register(s.api)
	if err != nil {
		return nil, err
	}
	s.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	return s, nil
}

func (s *Agent) Start() error {
	s.log.Infof("Starting")
	var err error

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

// Wait waits until agent's context is cancelled.
func (s *Agent) Wait() {
	<-s.doneCh
}

func (s *Agent) contextCancelHandler() {
	defer func() { close(s.doneCh) }()
	defer s.log.Info("Stopped")
	<-s.ctx.Done()

	err := s.listener.Close()
	if err != nil {
		s.log.WithError(err).Error("Unable to close RPC listener")
	}
}
