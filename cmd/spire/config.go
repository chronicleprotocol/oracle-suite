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
	feedsConfig "github.com/chronicleprotocol/oracle-suite/internal/config/feeds"
	spireConfig "github.com/chronicleprotocol/oracle-suite/internal/config/spire"
	transportConfig "github.com/chronicleprotocol/oracle-suite/internal/config/transport"
	"github.com/chronicleprotocol/oracle-suite/internal/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Config struct {
	Transport transportConfig.Transport `json:"transport"`
	Ethereum  ethereumConfig.Ethereum   `json:"ethereum"`
	Spire     spireConfig.Spire         `json:"spire"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
}

func PrepareAgentSupervisor(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	log := opts.Logger()
	sig, err := opts.Config.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, err
	}
	fed, err := opts.Config.Feeds.Addresses()
	if err != nil {
		return nil, err
	}
	tra, err := opts.Config.Transport.Configure(transportConfig.Dependencies{
		Signer: sig,
		Feeds:  fed,
		Logger: log,
	},
		map[string]transport.Message{messages.PriceMessageName: (*messages.Price)(nil)},
	)
	if err != nil {
		return nil, err
	}
	dat, err := opts.Config.Spire.ConfigureDatastore(spireConfig.DatastoreDependencies{
		Signer:    sig,
		Transport: tra,
		Feeds:     fed,
		Logger:    log,
	})
	if err != nil {
		return nil, err
	}
	age, err := opts.Config.Spire.ConfigureAgent(spireConfig.AgentDependencies{
		Signer:    sig,
		Transport: tra,
		Datastore: dat,
		Feeds:     fed,
		Logger:    log,
	})
	if err != nil {
		return nil, err
	}
	sup := supervisor.New(ctx)
	sup.Watch(tra, dat, age)
	return sup, nil
}

func PrepareClientSupervisor(ctx context.Context, opts *options) (*supervisor.Supervisor, *spire.Client, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, nil, err
	}
	sig, err := opts.Config.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, err
	}
	cli, err := opts.Config.Spire.ConfigureClient(spireConfig.ClientDependencies{
		Signer: sig,
	})
	if err != nil {
		return nil, nil, err
	}
	sup := supervisor.New(ctx)
	sup.Watch(cli)
	return sup, cli, nil
}
