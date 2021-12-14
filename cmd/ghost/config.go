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
	"errors"
	"fmt"

	"github.com/makerdao/oracle-suite/internal/config"
	ethereumConfig "github.com/makerdao/oracle-suite/internal/config/ethereum"
	feedsConfig "github.com/makerdao/oracle-suite/internal/config/feeds"
	ghostConfig "github.com/makerdao/oracle-suite/internal/config/ghost"
	goferConfig "github.com/makerdao/oracle-suite/internal/config/gofer"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/ghost"
	"github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"

	"github.com/makerdao/oracle-suite/pkg/log"
)

type Config struct {
	Gofer     goferConfig.Gofer         `json:"gofer"`
	Ethereum  ethereumConfig.Ethereum   `json:"ethereum"`
	Transport transportConfig.Transport `json:"transport"`
	Ghost     ghostConfig.Ghost         `json:"ghost"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies, noGoferRPC bool) (transport.Transport, gofer.Gofer, *ghost.Ghost, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, nil, err
	}
	cli, err := c.Ethereum.ConfigureEthereumClient(nil) // signer may be empty here
	if err != nil {
		return nil, nil, nil, err
	}
	gof, err := c.Gofer.ConfigureGofer(d.Context, cli, d.Logger, noGoferRPC)
	if err != nil {
		return nil, nil, nil, err
	}

	if sig.Address() == ethereum.EmptyAddress {
		return nil, nil, nil, errors.New("ethereum account must be configured")
	}
	fed, err := c.Feeds.Addresses()
	if err != nil {
		return nil, nil, nil, err
	}
	tra, err := c.Transport.Configure(transportConfig.Dependencies{
		Context: d.Context,
		Signer:  sig,
		Feeds:   fed,
		Logger:  d.Logger,
	},
		map[string]transport.Message{messages.PriceMessageName: (*messages.Price)(nil)},
	)
	if err != nil {
		return nil, nil, nil, err
	}
	gho, err := c.Ghost.Configure(ghostConfig.Dependencies{
		Context:   d.Context,
		Gofer:     gof,
		Signer:    sig,
		Transport: tra,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return tra, gof, gho, nil
}

type Services struct {
	ctxCancel context.CancelFunc
	Transport transport.Transport
	Gofer     gofer.Gofer
	Ghost     *ghost.Ghost
}

func PrepareServices(ctx context.Context, opts *options) (*Services, error) {
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
	tra, gof, gho, err := opts.Config.Configure(Dependencies{
		Context: ctx,
		Logger:  opts.Logger(),
	}, opts.GoferNoRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to load Ghost configuration: %w", err)
	}

	return &Services{
		ctxCancel: ctxCancel,
		Transport: tra,
		Gofer:     gof,
		Ghost:     gho,
	}, nil
}

func (s *Services) Start() error {
	var err error
	if g, ok := s.Gofer.(gofer.StartableGofer); ok {
		if err = g.Start(); err != nil {
			return err
		}
	}
	if err = s.Transport.Start(); err != nil {
		return err
	}
	if err = s.Ghost.Start(); err != nil {
		return err
	}
	return nil
}

func (s *Services) CancelAndWait() {
	s.ctxCancel()
	s.Transport.Wait()
	s.Ghost.Wait()
	if g, ok := s.Gofer.(gofer.StartableGofer); ok {
		g.Wait()
	}
}
