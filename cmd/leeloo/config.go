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

	"github.com/chronicleprotocol/oracle-suite/internal/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/internal/config/ethereum"
	leelooConfig "github.com/chronicleprotocol/oracle-suite/internal/config/eventpublisher"
	feedsConfig "github.com/chronicleprotocol/oracle-suite/internal/config/feeds"
	transportConfig "github.com/chronicleprotocol/oracle-suite/internal/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Config struct {
	Leeloo    leelooConfig.EventPublisher `json:"leeloo"`
	Ethereum  ethereumConfig.Ethereum     `json:"ethereum"`
	Transport transportConfig.Transport   `json:"transport"`
	Feeds     feedsConfig.Feeds           `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies) (transport.Transport, *publisher.EventPublisher, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, err
	}
	fed, err := c.Feeds.Addresses()
	if err != nil {
		return nil, nil, err
	}
	tra, err := c.Transport.Configure(transportConfig.Dependencies{
		Context: d.Context,
		Signer:  sig,
		Feeds:   fed,
		Logger:  d.Logger,
	},
		map[string]transport.Message{messages.EventMessageName: (*messages.Event)(nil)},
	)
	if err != nil {
		return nil, nil, err
	}
	lel, err := c.Leeloo.Configure(leelooConfig.Dependencies{
		Context:   d.Context,
		Signer:    sig,
		Transport: tra,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, nil, err
	}
	return tra, lel, nil
}

type Service struct {
	ctxCancel context.CancelFunc
	Transport transport.Transport
	Leeloo    *publisher.EventPublisher
}

func PrepareService(ctx context.Context, opts *options) (*Service, error) {
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
		return nil, err
	}

	// Services:
	tra, lel, err := opts.Config.Configure(Dependencies{
		Context: ctx,
		Logger:  opts.Logger(),
	})
	if err != nil {
		return nil, err
	}

	return &Service{
		ctxCancel: ctxCancel,
		Transport: tra,
		Leeloo:    lel,
	}, nil
}

func (s *Service) Start() error {
	var err error
	if err = s.Leeloo.Start(); err != nil {
		return err
	}
	if err = s.Transport.Start(); err != nil {
		return err
	}
	return nil
}

func (s *Service) CancelAndWait() {
	s.ctxCancel()
	<-s.Leeloo.Wait()
	<-s.Transport.Wait()
}
