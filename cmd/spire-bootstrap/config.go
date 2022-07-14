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
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
)

type Config struct {
	Transport transportConfig.Transport `json:"transport"`
	Logger    loggerConfig.Logger       `json:"logger"`
}

func PrepareSupervisor(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
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
	tra, err := opts.Config.Transport.ConfigureP2PBoostrap(transportConfig.BootstrapDependencies{
		Logger: log,
	})
	if err != nil {
		return nil, fmt.Errorf(`transport config error: %w`, err)
	}
	if _, ok := tra.(*libp2p.P2P); !ok {
		return nil, errors.New("spire-bootstrap works only with the libp2p transport")
	}
	sup := supervisor.New(ctx, log)
	sup.Watch(tra, sysmon.New(time.Minute, log))
	return sup, nil
}
