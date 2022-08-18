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
	eventAPIConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/eventapi"
	feedsConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/feeds"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportevm"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportstarknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Config struct {
	Lair      eventAPIConfig.EventAPI   `json:"lair"`
	Transport transportConfig.Transport `json:"transport"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
	Logger    loggerConfig.Logger       `json:"logger"`
}

func PrepareServices(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf(`config error: %w`, err)
	}
	log, err := opts.Config.Logger.Configure(loggerConfig.Dependencies{
		AppName:    "lair",
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	fed, err := opts.Config.Feeds.Addresses()
	if err != nil {
		return nil, fmt.Errorf(`feeds config error: %w`, err)
	}
	tra, err := opts.Config.Transport.Configure(transportConfig.Dependencies{
		Signer: geth.NewSigner(nil),
		Feeds:  fed,
		Logger: log,
	},
		map[string]transport.Message{messages.EventV1MessageName: (*messages.Event)(nil)},
	)
	if err != nil {
		return nil, fmt.Errorf(`transport config error: %w`, err)
	}
	sto, err := opts.Config.Lair.ConfigureStorage()
	if err != nil {
		return nil, fmt.Errorf(`lair config error: %w`, err)
	}
	evs, err := store.New(store.Config{
		EventTypes: []string{teleportevm.TeleportEventType, teleportstarknet.TeleportEventType},
		Storage:    sto,
		Transport:  tra,
		Logger:     log,
	})
	if err != nil {
		return nil, fmt.Errorf(`lair config error: %w`, err)
	}
	api, err := opts.Config.Lair.Configure(eventAPIConfig.Dependencies{
		EventStore: evs,
		Transport:  tra,
		Logger:     log,
	})
	if err != nil {
		return nil, fmt.Errorf(`lair config error: %w`, err)
	}
	sup := supervisor.New(log)
	sup.Watch(tra, evs, api, sysmon.New(time.Minute, log))
	if l, ok := log.(supervisor.Service); ok {
		sup.Watch(l)
	}
	return sup, nil
}
