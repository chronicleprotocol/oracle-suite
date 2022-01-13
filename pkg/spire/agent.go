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
	"fmt"
	"net/http"
	"net/rpc"

	"github.com/chronicleprotocol/oracle-suite/internal/httpserver"
	"github.com/chronicleprotocol/oracle-suite/pkg/datastore"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

const AgentLoggerTag = "SPIRE_AGENT"

type Agent struct {
	ctx    context.Context
	waitCh chan error

	srv *httpserver.HTTPServer
	log log.Logger
}

type AgentConfig struct {
	Datastore datastore.Datastore
	Transport transport.Transport
	Signer    ethereum.Signer
	Address   string
	Logger    log.Logger
}

func NewAgent(ctx context.Context, cfg AgentConfig) (*Agent, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	logger := cfg.Logger.WithField("tag", AgentLoggerTag)
	rpcSrv := rpc.NewServer()
	err := rpcSrv.Register(&API{
		datastore: cfg.Datastore,
		transport: cfg.Transport,
		signer:    cfg.Signer,
		log:       logger,
	})
	if err != nil {
		return nil, err
	}
	return &Agent{
		ctx:    ctx,
		waitCh: make(chan error),
		srv:    httpserver.New(ctx, &http.Server{Addr: cfg.Address, Handler: rpcSrv}),
		log:    logger,
	}, nil
}

func (s *Agent) Start() error {
	s.log.Infof("Starting")
	err := s.srv.Start()
	if err != nil {
		return fmt.Errorf("unable to start the HTTP server: %w", err)
	}
	go s.contextCancelHandler()

	return nil
}

// Wait waits until agent's context is cancelled.
func (s *Agent) Wait() chan error {
	return s.waitCh
}

func (s *Agent) contextCancelHandler() {
	defer s.log.Info("Stopped")
	<-s.ctx.Done()
	s.waitCh <- <-s.srv.Wait()
}
