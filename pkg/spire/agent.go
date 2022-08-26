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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/httpserver"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

const AgentLoggerTag = "SPIRE_AGENT"

// defaultHTTPTimeout is the default timeout for the HTTP server.
const defaultHTTPTimeout = 3 * time.Second

type Agent struct {
	ctx context.Context

	srv *httpserver.HTTPServer
	log log.Logger
}

type AgentConfig struct {
	PriceStore *store.PriceStore
	Transport  transport.Transport
	Signer     ethereum.Signer
	Address    string
	Logger     log.Logger
}

func NewAgent(cfg AgentConfig) (*Agent, error) {
	logger := cfg.Logger.WithField("tag", AgentLoggerTag)
	rpcSrv := rpc.NewServer()
	err := rpcSrv.Register(&API{
		priceStore: cfg.PriceStore,
		transport:  cfg.Transport,
		signer:     cfg.Signer,
		log:        logger,
	})
	if err != nil {
		return nil, err
	}
	return &Agent{
		srv: httpserver.New(&http.Server{
			Addr:              cfg.Address,
			Handler:           rpcSrv,
			IdleTimeout:       defaultHTTPTimeout,
			ReadTimeout:       defaultHTTPTimeout,
			WriteTimeout:      defaultHTTPTimeout,
			ReadHeaderTimeout: defaultHTTPTimeout,
		}),
		log: logger,
	}, nil
}

func (s *Agent) Start(ctx context.Context) error {
	if s.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	s.log.Infof("Starting")
	s.ctx = ctx
	err := s.srv.Start(ctx)
	if err != nil {
		return fmt.Errorf("unable to start the HTTP server: %w", err)
	}
	go s.contextCancelHandler()
	return nil
}

// Wait waits until agent's context is cancelled.
func (s *Agent) Wait() chan error {
	return s.srv.Wait()
}

func (s *Agent) contextCancelHandler() {
	defer s.log.Info("Stopped")
	<-s.ctx.Done()
}
