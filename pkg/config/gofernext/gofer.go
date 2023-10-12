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

package gofer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcl/v2"

	suite "github.com/chronicleprotocol/oracle-suite"
	dataproviderConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/dataprovider"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

// Config is the configuration for Gofer.
type Config struct {
	Gofer    dataproviderConfig.Config `hcl:"gofer,block"`
	Ethereum *ethereumConfig.Config    `hcl:"ethereum,block,optional"`
	Logger   *loggerConfig.Config      `hcl:"logger,block,optional"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (Config) DefaultEmbeds() [][]byte {
	return [][]byte{
		suite.ConfigContracts,
		suite.ConfigDefaults,
		suite.ConfigGofer,
		suite.ConfigEthereum,
	}
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	DataProvider datapoint.Provider
	Logger       log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	if p, ok := s.DataProvider.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(p)
	}
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *Services) Wait() <-chan error {
	return s.supervisor.Wait()
}

// Services returns the services configured for Gofer.
func (c *Config) Services(baseLogger log.Logger, appName, appVersion string) (pkgSupervisor.Service, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    appName,
		AppVersion: appVersion,
		BaseLogger: baseLogger,
	})
	if err != nil {
		return nil, err
	}
	clients, err := c.Ethereum.ClientRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, err
	}
	priceProvider, err := c.Gofer.ConfigureDataProvider(dataproviderConfig.Dependencies{
		HTTPClient: &http.Client{},
		Clients:    clients,
		Logger:     logger,
	})
	if err != nil {
		return nil, err
	}
	return &Services{
		DataProvider: priceProvider,
		Logger:       logger,
	}, nil
}
