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

	"github.com/chronicleprotocol/oracle-suite/pkg/confighcl"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/confighcl/ethereum"
	feedsConfig "github.com/chronicleprotocol/oracle-suite/pkg/confighcl/feeds"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/confighcl/logger"
	spireConfig "github.com/chronicleprotocol/oracle-suite/pkg/confighcl/spire"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/confighcl/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type ConfigHCL struct {
	Transport transportConfig.Transport `hcl:"transport,block"`
	Ethereum  ethereumConfig.Ethereum   `hcl:"ethereum,block"`
	Spire     spireConfig.Spire         `hcl:"spire,block"`
	Feeds     feedsConfig.Feeds         `hcl:"feeds,block"`
	Logger    *loggerConfig.Logger      `hcl:"logger,block"`
}

func PrepareAgentServicesHCL(_ context.Context, opts *options) (*supervisor.Supervisor, error) {
	err := confighcl.LoadFile(&opts.ConfigHCL, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf(`config error: %w`, err)
	}
	log, err := opts.ConfigHCL.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "spire",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, fmt.Errorf(`logger config error: %w`, err)
	}
	sig, err := opts.ConfigHCL.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	fed, err := opts.ConfigHCL.Feeds.ConfigureAddresses()
	if err != nil {
		return nil, fmt.Errorf(`feeds config error: %w`, err)
	}
	tra, err := opts.ConfigHCL.Transport.Configure(transportConfig.Dependencies{
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
	dat, err := opts.ConfigHCL.Spire.ConfigurePriceStore(spireConfig.PriceStoreDependencies{
		Signer:    sig,
		Transport: tra,
		Feeds:     fed,
		Logger:    log,
	})
	if err != nil {
		return nil, fmt.Errorf(`spire config error: %w`, err)
	}
	age, err := opts.ConfigHCL.Spire.ConfigureAgent(spireConfig.AgentDependencies{
		Signer:     sig,
		Transport:  tra,
		PriceStore: dat,
		Feeds:      fed,
		Logger:     log,
	})
	if err != nil {
		return nil, fmt.Errorf(`spire config error: %w`, err)
	}
	sup := supervisor.New(log)
	sup.Watch(tra, dat, age, sysmon.New(time.Minute, log))
	if l, ok := log.(supervisor.Service); ok {
		sup.Watch(l)
	}
	return sup, nil
}

func PrepareClientServicesHCL(_ context.Context, opts *options) (*supervisor.Supervisor, *spire.Client, error) {
	err := confighcl.LoadFile(&opts.ConfigHCL, opts.ConfigFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf(`config error: %w`, err)
	}
	log, err := opts.ConfigHCL.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "spire",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	sig, err := opts.ConfigHCL.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	cli, err := opts.ConfigHCL.Spire.ConfigureClient(spireConfig.ClientDependencies{
		Signer: sig,
	})
	if err != nil {
		return nil, nil, fmt.Errorf(`spire config error: %w`, err)
	}
	sup := supervisor.New(log)
	sup.Watch(cli)
	if l, ok := log.(supervisor.Service); ok {
		sup.Watch(l)
	}
	return sup, cli, nil
}

func PrepareStreamServicesHCL(_ context.Context, opts *options) (*supervisor.Supervisor, transport.Transport, error) {
	err := confighcl.LoadFile(&opts.ConfigHCL, opts.ConfigFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf(`config error: %w`, err)
	}
	log, err := opts.ConfigHCL.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "spire",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf(`logger config error: %w`, err)
	}
	fed, err := opts.ConfigHCL.Feeds.ConfigureAddresses()
	if err != nil {
		return nil, nil, fmt.Errorf(`feeds config error: %w`, err)
	}
	sig, err := opts.ConfigHCL.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	tra, err := opts.ConfigHCL.Transport.Configure(transportConfig.Dependencies{
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
		return nil, nil, fmt.Errorf(`transport config error: %w`, err)
	}
	sup := supervisor.New(log)
	sup.Watch(tra)
	if l, ok := log.(supervisor.Service); ok {
		sup.Watch(l)
	}
	return sup, tra, nil
}
