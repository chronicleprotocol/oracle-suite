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

	"github.com/chronicleprotocol/oracle-suite/internal/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/internal/config/ethereum"
	feedsConfig "github.com/chronicleprotocol/oracle-suite/internal/config/feeds"
	ghostConfig "github.com/chronicleprotocol/oracle-suite/internal/config/ghost"
	goferConfig "github.com/chronicleprotocol/oracle-suite/internal/config/gofer"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/internal/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/internal/config/transport"
	"github.com/chronicleprotocol/oracle-suite/internal/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
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
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	sig, err := opts.Config.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	cli, err := opts.Config.Ethereum.ConfigureEthereumClient(nil) // signer may be empty here
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
		map[string]transport.Message{messages.PriceMessageName: (*messages.Price)(nil)},
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
	sup := supervisor.New(ctx)
	sup.Watch(tra, gho)
	if g, ok := gof.(gofer.StartableGofer); ok {
		sup.Watch(g)
	}
	return sup, nil
}
