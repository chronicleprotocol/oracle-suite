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

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/marshal"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	goferConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/gofer"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
)

type Config struct {
	Ethereum ethereumConfig.ConfigEthereum `hcl:"ethereum,block"`
	Gofer    goferConfig.ConfigGofer       `hcl:"gofer,block"`
	Logger   *loggerConfig.ConfigLogger    `hcl:"logger,block"`

	Remain hcl.Body `hcl:",remain"` // To ignore unknown blocks.
}

func PrepareClientServices(
	ctx context.Context,
	opts *options,
) (
	*supervisor.Supervisor,
	provider.Provider,
	marshal.Marshaller,
	provider.PriceHook,
	error,
) {

	err := config.LoadFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf(`config error: %w`, err)
	}
	logger, err := opts.Config.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "leeloo",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf(`logger config error: %w`, err)
	}
	clients, err := opts.Config.Ethereum.ClientRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	gofer, err := opts.Config.Gofer.ConfigureGofer(goferConfig.Dependencies{
		Clients: clients,
		Logger:  logger,
	}, opts.NoRPC)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf(`gofer config error: %w`, err)
	}
	hook, err := opts.Config.Gofer.ConfigurePriceHook(goferConfig.HookDependencies{
		Context: ctx,
		Clients: clients,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf(`price hook config error: %w`, err)
	}
	marshaler, err := marshal.NewMarshal(opts.Format.format)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf(`invalid format option: %w`, err)
	}
	sup := supervisor.New(logger)
	if g, ok := gofer.(supervisor.Service); ok {
		sup.Watch(g)
	}
	if l, ok := logger.(supervisor.Service); ok {
		sup.Watch(l)
	}
	return sup, gofer, marshaler, hook, nil
}

func PrepareAgentServices(_ context.Context, opts *options) (*supervisor.Supervisor, error) {
	err := config.LoadFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf(`config error: %w`, err)
	}
	logger, err := opts.Config.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "leeloo",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, fmt.Errorf(`logger config error: %w`, err)
	}
	clients, err := opts.Config.Ethereum.ClientRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	gofer, err := opts.Config.Gofer.ConfigureAsyncGofer(goferConfig.AsyncDependencies{
		Clients: clients,
		Logger:  nil,
	})
	if err != nil {
		return nil, fmt.Errorf(`gofer config error: %w`, err)
	}
	agent, err := opts.Config.Gofer.ConfigureRPCAgent(goferConfig.AgentDependencies{
		Provider: gofer,
		Logger:   logger,
	})
	if err != nil {
		return nil, fmt.Errorf(`gofer config error: %w`, err)
	}
	sup := supervisor.New(logger)
	sup.Watch(gofer.(supervisor.Service), agent, sysmon.New(time.Minute, logger))
	if l, ok := logger.(supervisor.Service); ok {
		sup.Watch(l)
	}
	return sup, nil
}
