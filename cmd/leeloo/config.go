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
	"github.com/chronicleprotocol/oracle-suite/internal/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Config struct {
	Leeloo    leelooConfig.EventPublisher `json:"leeloo"`
	Ethereum  ethereumConfig.Ethereum     `json:"ethereum"`
	Transport transportConfig.Transport   `json:"transport"`
	Feeds     feedsConfig.Feeds           `json:"feeds"`
}

func PrepareSupervisor(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
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
		map[string]transport.Message{messages.EventMessageName: (*messages.Event)(nil)},
	)
	if err != nil {
		return nil, err
	}
	lee, err := opts.Config.Leeloo.Configure(leelooConfig.Dependencies{
		Signer:    sig,
		Transport: tra,
		Logger:    log,
	})
	if err != nil {
		return nil, err
	}
	sup := supervisor.New(ctx)
	sup.Watch(tra, lee)
	return sup, nil
}
