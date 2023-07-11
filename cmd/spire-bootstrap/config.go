//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

	"github.com/hashicorp/hcl/v2"

	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
)

type Config struct {
	Transport transportConfig.Config `hcl:"transport,block"`
	Logger    *loggerConfig.Config   `hcl:"logger,block"`

	Remain hcl.Body `hcl:",remain"` // To ignore unknown blocks.
}

func PrepareSupervisor(_ context.Context, opts *options) (*pkgSupervisor.Supervisor, error) {
	if err := opts.LoadConfigFiles(&opts.Config); err != nil {
		return nil, fmt.Errorf(`config error: %w`, err)
	}
	logger, err := opts.Config.Logger.Logger(loggerConfig.Dependencies{
		BaseLogger: opts.Logger(),
	})
	if err != nil {
		return nil, fmt.Errorf(`ethereum config error: %w`, err)
	}
	transport, err := opts.Config.Transport.LibP2PBootstrap(transportConfig.BootstrapDependencies{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf(`transport config error: %w`, err)
	}
	if _, ok := transport.(*libp2p.P2P); !ok {
		return nil, errors.New("spire-bootstrap works only with the libp2p transport")
	}
	supervisor := pkgSupervisor.New(logger)
	supervisor.Watch(transport, sysmon.New(time.Minute, logger))
	if l, ok := logger.(pkgSupervisor.Service); ok {
		supervisor.Watch(l)
	}
	return supervisor, nil
}
