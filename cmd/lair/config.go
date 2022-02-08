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
	"time"

	"github.com/chronicleprotocol/oracle-suite/internal/config"
	eventAPIConfig "github.com/chronicleprotocol/oracle-suite/internal/config/eventapi"
	feedsConfig "github.com/chronicleprotocol/oracle-suite/internal/config/feeds"
	transportConfig "github.com/chronicleprotocol/oracle-suite/internal/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	eventAPI "github.com/chronicleprotocol/oracle-suite/pkg/event/api"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const week = time.Hour * 7 * 24

type Config struct {
	Lair      eventAPIConfig.EventAPI   `json:"lair"`
	Transport transportConfig.Transport `json:"transport"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies) (transport.Transport, *store.EventStore, *eventAPI.EventAPI, error) {
	fed, err := c.Feeds.Addresses()
	if err != nil {
		return nil, nil, nil, err
	}
	tra, err := c.Transport.Configure(transportConfig.Dependencies{
		Context: d.Context,
		Signer:  geth.NewSigner(nil),
		Feeds:   fed,
		Logger:  d.Logger,
	},
		map[string]transport.Message{messages.EventMessageName: (*messages.Event)(nil)},
	)
	if err != nil {
		return nil, nil, nil, err
	}
	sto, err := c.Lair.ConfigureStorage()
	if err != nil {
		return nil, nil, nil, err
	}
	evs, err := store.New(d.Context, store.Config{
		Storage:   sto,
		Transport: tra,
		Log:       d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	api, err := c.Lair.Configure(eventAPIConfig.Dependencies{
		Context:    d.Context,
		EventStore: evs,
		Transport:  tra,
		Logger:     d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return tra, evs, api, nil
}

type Service struct {
	ctxCancel  context.CancelFunc
	Transport  transport.Transport
	EventStore *store.EventStore
	Lair       *eventAPI.EventAPI
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
	tra, evs, lel, err := opts.Config.Configure(Dependencies{
		Context: ctx,
		Logger:  opts.Logger(),
	})
	if err != nil {
		return nil, err
	}

	return &Service{
		ctxCancel:  ctxCancel,
		Transport:  tra,
		EventStore: evs,
		Lair:       lel,
	}, nil
}

func (s *Service) Start() error {
	var err error
	if err = s.Transport.Start(); err != nil {
		return err
	}
	if err = s.EventStore.Start(); err != nil {
		return err
	}
	if err = s.Lair.Start(); err != nil {
		return err
	}
	return nil
}

func (s *Service) CancelAndWait() {
	s.ctxCancel()
	<-s.Lair.Wait()
	<-s.Transport.Wait()
}
