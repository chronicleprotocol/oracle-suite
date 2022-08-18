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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	feedsConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/feeds"
	ghostConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ghost"
	goferConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/gofer"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Config struct {
	Gofer     goferConfig.Gofer         `json:"gofer"`
	Ethereum  ethereumConfig.Ethereum   `json:"ethereum"`
	Transport transportConfig.Transport `json:"transport"`
	Ghost     ghostConfig.Ghost         `json:"ghost"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
	Logger    loggerConfig.Logger       `json:"logger"`
}

func PrepareServices(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf(`config error: %w`, err)
	}
	log, err := opts.Config.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "ghost",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	sig, err := opts.Config.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	cli, err := opts.Config.Ethereum.ConfigureEthereumClient(nil, log) // signer may be empty here
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	gof, err := opts.Config.Gofer.ConfigureGofer(cli, log, opts.GoferNoRPC)
	if err != nil {
		return nil, fmt.Errorf(`gofer config error: %w`, err)
	}

	if sig.Address() == ethereum.EmptyAddress {
		return nil, errors.New("ethereum account must be configured")
	}
	fed, err := opts.Config.Feeds.Addresses()
	if err != nil {
		return nil, fmt.Errorf(`feeds config error: %w`, err)
	}
	tra, err := opts.Config.Transport.Configure(transportConfig.Dependencies{
		Signer: sig,
		Feeds:  fed,
		Logger: log,
	},
		map[string]transport.Message{
			messages.PriceV0MessageName: (*messages.Price)(nil),
			messages.PriceV1MessageName: (*messages.Price)(nil),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(`transport config error: %w`, err)
	}
	gho, err := opts.Config.Ghost.Configure(ghostConfig.Dependencies{
		Gofer:     gof,
		Signer:    sig,
		Transport: tra,
		Logger:    log,
	})
	if err != nil {
		return nil, fmt.Errorf(`ghost config error: %w`, err)
	}
	sup := supervisor.New(log)
	sup.Watch(tra, gho, sysmon.New(time.Minute, log))
	if g, ok := gof.(supervisor.Service); ok {
		sup.Watch(g)
	}
	if l, ok := log.(supervisor.Service); ok {
		sup.Watch(l)
	}
	return sup, nil
}
