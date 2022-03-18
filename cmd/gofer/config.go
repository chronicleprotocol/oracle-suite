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
	goferConfig "github.com/chronicleprotocol/oracle-suite/internal/config/gofer"
	"github.com/chronicleprotocol/oracle-suite/internal/gofer/marshal"
	"github.com/chronicleprotocol/oracle-suite/internal/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
)

type Config struct {
	Ethereum ethereumConfig.Ethereum `json:"ethereum"`
	Gofer    goferConfig.Gofer       `json:"gofer"`
}

func PrepareClientServices(
	ctx context.Context,
	opts *options,
) (*supervisor.Supervisor, gofer.Gofer, marshal.Marshaller, error) {

	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, nil, nil, err
	}
	log := opts.Logger()
	cli, err := opts.Config.Ethereum.ConfigureEthereumClient(nil)
	if err != nil {
		return nil, nil, nil, err
	}
	gof, err := opts.Config.Gofer.ConfigureGofer(cli, log, opts.NoRPC)
	if err != nil {
		return nil, nil, nil, err
	}
	mar, err := marshal.NewMarshal(opts.Format.format)
	if err != nil {
		return nil, nil, nil, err
	}
	sup := supervisor.New(ctx)
	if g, ok := gof.(gofer.StartableGofer); ok {
		sup.Watch(g)
	}
	return sup, gof, mar, nil
}

func PrepareAgentServices(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	log := opts.Logger()
	cli, err := opts.Config.Ethereum.ConfigureEthereumClient(nil)
	if err != nil {
		return nil, err
	}
	gof, err := opts.Config.Gofer.ConfigureRPCAgent(cli, log)
	if err != nil {
		return nil, err
	}
	sup := supervisor.New(ctx)
	sup.Watch(gof)
	return sup, nil
}
