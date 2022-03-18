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
	transportConfig "github.com/chronicleprotocol/oracle-suite/internal/config/transport"
	"github.com/chronicleprotocol/oracle-suite/internal/supervisor"
)

type Config struct {
	Transport transportConfig.Transport `json:"transport"`
}

func PrepareSupervisor(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	log := opts.Logger()
	tra, err := opts.Config.Transport.ConfigureP2PBoostrap(transportConfig.BootstrapDependencies{
		Logger: log,
	})
	if err != nil {
		return nil, err
	}
	sup := supervisor.New(ctx)
	sup.Watch(tra)
	return sup, nil
}
