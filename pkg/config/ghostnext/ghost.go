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

package ghostnext

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/config"
	configGoferNext "github.com/chronicleprotocol/oracle-suite/pkg/config/dataprovider"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	feedConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/feednext"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/feed"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	pkgTransport "github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// Config is the configuration for Ghost.
type Config struct {
	Ghost     feedConfig.Config      `hcl:"ghost,block"`
	Gofer     configGoferNext.Config `hcl:"gofer,block"`
	Ethereum  ethereumConfig.Config  `hcl:"ethereum,block"`
	Transport transportConfig.Config `hcl:"transport,block"`
	Logger    *loggerConfig.Config   `hcl:"logger,block,optional"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (Config) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Contracts,
		config.Defaults,
		config.Ghost,
		config.Gofer,
		config.Ethereum,
		config.Transport,
	}
}

// Services returns the services configured for Lair.
func (c *Config) Services(baseLogger log.Logger, appName string, appVersion string) (pkgSupervisor.Service, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    appName,
		AppVersion: appVersion,
		BaseLogger: baseLogger,
	})
	if err != nil {
		return nil, err
	}
	keys, err := c.Ethereum.KeyRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, err
	}
	clients, err := c.Ethereum.ClientRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, err
	}
	messageMap, err := messages.AllMessagesMap.SelectByTopic(
		messages.DataPointV1MessageName,
	)
	if err != nil {
		return nil, err
	}
	transport, err := c.Transport.Transport(transportConfig.Dependencies{
		Keys:       keys,
		Clients:    clients,
		Messages:   messageMap,
		Logger:     logger,
		AppName:    appName,
		AppVersion: appVersion,
	})
	if err != nil {
		return nil, err
	}
	dataProvider, err := c.Gofer.ConfigureDataProvider(configGoferNext.Dependencies{
		Clients: clients,
		Logger:  logger,
	})
	if err != nil {
		return nil, err
	}
	feedService, err := c.Ghost.ConfigureFeed(feedConfig.Dependencies{
		KeysRegistry: keys,
		DataProvider: dataProvider,
		Transport:    transport,
		Logger:       logger,
	})
	if err != nil {
		return nil, err
	}

	return &Services{
		Feed:      feedService,
		Transport: transport,
		Logger:    logger,
	}, nil
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	Feed      *feed.Feed
	Transport pkgTransport.Service
	Logger    log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.Transport, s.Feed)
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *Services) Wait() <-chan error {
	return s.supervisor.Wait()
}
