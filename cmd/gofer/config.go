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

package main

import (
	"context"
	"fmt"

	"github.com/makerdao/oracle-suite/internal/config"
	ethereumConfig "github.com/makerdao/oracle-suite/internal/config/ethereum"
	goferConfig "github.com/makerdao/oracle-suite/internal/config/gofer"
	"github.com/makerdao/oracle-suite/internal/gofer/marshal"
	pkgGofer "github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/gofer/rpc"
	"github.com/makerdao/oracle-suite/pkg/log"
)

type Config struct {
	Ethereum ethereumConfig.Ethereum `json:"ethereum"`
	Gofer    goferConfig.Gofer       `json:"gofer"`
}

func (c *Config) Configure(ctx context.Context, logger log.Logger, noRPC bool) (pkgGofer.Gofer, error) {
	cli, err := c.Ethereum.ConfigureEthereumClient(nil)
	if err != nil {
		return nil, err
	}
	return c.Gofer.ConfigureGofer(ctx, cli, logger, noRPC)
}

func (c *Config) ConfigureRPCAgent(ctx context.Context, logger log.Logger) (*rpc.Agent, error) {
	cli, err := c.Ethereum.ConfigureEthereumClient(nil)
	if err != nil {
		return nil, err
	}
	return c.Gofer.ConfigureRPCAgent(ctx, cli, logger)
}

type GoferClientServices struct {
	ctxCancel  context.CancelFunc
	Gofer      pkgGofer.Gofer
	Marshaller marshal.Marshaller
}

func PrepareGoferClientServices(ctx context.Context, opts *options) (*GoferClientServices, error) {
	var err error
	ctx, ctxCancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			ctxCancel()
		}
	}()

	// Load config file:
	err = config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Services:
	gof, err := opts.Config.Configure(ctx, opts.Logger(), opts.NoRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to load Gofer configuration: %w", err)
	}
	mar, err := marshal.NewMarshal(opts.Format.format)
	if err != nil {
		return nil, err
	}

	return &GoferClientServices{
		ctxCancel:  ctxCancel,
		Gofer:      gof,
		Marshaller: mar,
	}, nil
}

func (s *GoferClientServices) Start() error {
	if g, ok := s.Gofer.(pkgGofer.StartableGofer); ok {
		return g.Start()
	}
	return nil
}

func (s *GoferClientServices) CancelAndWait() {
	s.ctxCancel()
	if g, ok := s.Gofer.(pkgGofer.StartableGofer); ok {
		g.Wait()
	}
}

type GoferAgentService struct {
	ctxCancel context.CancelFunc
	Agent     *rpc.Agent
}

func PrepareGoferAgentService(ctx context.Context, opts *options) (*GoferAgentService, error) {
	var err error
	ctx, ctxCancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			ctxCancel()
		}
	}()

	// Load config file:
	err = config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Services:
	age, err := opts.Config.ConfigureRPCAgent(ctx, opts.Logger())
	if err != nil {
		return nil, fmt.Errorf("failed to load Gofer configuration: %w", err)
	}

	return &GoferAgentService{
		ctxCancel: ctxCancel,
		Agent:     age,
	}, nil
}

func (s *GoferAgentService) Start() error {
	return s.Agent.Start()
}

func (s *GoferAgentService) CancelAndWait() {
	s.ctxCancel()
	s.Agent.Wait()
}
