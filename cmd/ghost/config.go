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
	"time"

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	feedsConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/feeds"
	ghostConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ghost"
	goferConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/gofer"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	pkgTransport "github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Config struct {
	Gofer     goferConfig.ConfigGofer         `hcl:"gofer,block"`
	Ethereum  ethereumConfig.ConfigEthereum   `hcl:"ethereum,block"`
	Transport transportConfig.ConfigTransport `hcl:"transport,block"`
	Ghost     ghostConfig.ConfigGhost         `hcl:"ghost,block"`
	Feeds     feedsConfig.ConfigFeeds         `hcl:"feeds,block"`
	Logger    *loggerConfig.ConfigLogger      `hcl:"logger,block"`

	Remain hcl.Body `hcl:",remain"` // To ignore unknown blocks.
}

func PrepareServices(_ context.Context, opts *options) (*pkgSupervisor.Supervisor, error) {
	err := config.LoadFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf(`config error: %w`, err)
	}
	logger, err := opts.Config.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "leeloo",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, fmt.Errorf(`logger config error: %w`, err)
	}
	keys, err := opts.Config.Ethereum.KeyRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	clients, err := opts.Config.Ethereum.ClientRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	gofer, err := opts.Config.Gofer.Gofer(goferConfig.Dependencies{
		Clients: clients,
		Logger:  logger,
	}, opts.GoferNoRPC)
	if err != nil {
		return nil, fmt.Errorf(`gofer config error: %w`, err)
	}
	feeds, err := opts.Config.Feeds.Addresses()
	if err != nil {
		return nil, fmt.Errorf(`feeds config error: %w`, err)
	}
	transport, err := opts.Config.Transport.Transport(transportConfig.Dependencies{
		Keys:    keys,
		Clients: clients,
		Messages: map[string]pkgTransport.Message{
			messages.PriceV0MessageName: (*messages.Price)(nil),
			messages.PriceV1MessageName: (*messages.Price)(nil),
		},
		Feeds:  feeds,
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf(`transport config error: %w`, err)
	}
	ghost, err := opts.Config.Ghost.Ghost(ghostConfig.Dependencies{
		Keys:      keys,
		Gofer:     gofer,
		Transport: transport,
		Logger:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf(`ghost config error: %w`, err)
	}
	supervisor := pkgSupervisor.New(logger)
	supervisor.Watch(transport, ghost, sysmon.New(time.Minute, logger))
	if g, ok := gofer.(pkgSupervisor.Service); ok {
		supervisor.Watch(g)
	}
	if l, ok := logger.(pkgSupervisor.Service); ok {
		supervisor.Watch(l)
	}
	return supervisor, nil
}
