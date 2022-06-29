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

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	goferConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/gofer"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/marshal"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
)

type Config struct {
	Ethereum ethereumConfig.Ethereum `json:"ethereum"`
	Gofer    goferConfig.Gofer       `json:"gofer"`
	Logger   loggerConfig.Logger     `json:"logger"`
}

func PrepareClientServices(
	ctx context.Context,
	opts *options,
) (*supervisor.Supervisor, gofer.Gofer, marshal.Marshaller, error) {

	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(`config error: %w`, err)
	}
	log, err := opts.Config.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "gofer",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	cli, err := opts.Config.Ethereum.ConfigureEthereumClient(nil, log)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	gof, err := opts.Config.Gofer.ConfigureGofer(cli, log, opts.NoRPC)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(`gofer config error: %w`, err)
	}
	mar, err := marshal.NewMarshal(opts.Format.format)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(`invalid format option: %w`, err)
	}
	sup := supervisor.New(ctx, log)
	if g, ok := gof.(gofer.StartableGofer); ok {
		sup.Watch(g)
	}
	return sup, gof, mar, nil
}

func PrepareAgentServices(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf(`config error: %w`, err)
	}
	log, err := opts.Config.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "gofer",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, fmt.Errorf(`logger config error: %w`, err)
	}
	cli, err := opts.Config.Ethereum.ConfigureEthereumClient(nil, log)
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	gof, err := opts.Config.Gofer.ConfigureAsyncGofer(cli, log)
	if err != nil {
		return nil, fmt.Errorf(`gofer config error: %w`, err)
	}
	age, err := opts.Config.Gofer.ConfigureRPCAgent(cli, gof, log)
	if err != nil {
		return nil, fmt.Errorf(`gofer config error: %w`, err)
	}
	sup := supervisor.New(ctx, log)
	sup.Watch(gof, age, sysmon.New(time.Minute, log))
	return sup, nil
}
