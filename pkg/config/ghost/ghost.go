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

package ghost

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	feedConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/feed"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	priceproviderConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/priceprovider"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/feed"

	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	pkgTransport "github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// Config is the configuration for Lair.
type Config struct {
	Ghost     feedConfig.Config          `hcl:"ghost,block"`
	Gofer     priceproviderConfig.Config `hcl:"gofer,block"`
	Ethereum  ethereumConfig.Config      `hcl:"ethereum,block"`
	Transport transportConfig.Config     `hcl:"transport,block"`
	Logger    *loggerConfig.Config       `hcl:"logger,block,optional"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	Feed      *feed.Feed
	Transport pkgTransport.TransportService
	Logger    log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.Transport, s.Feed, sysmon.New(time.Minute, s.Logger))
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *Services) Wait() <-chan error {
	return s.supervisor.Wait()
}

// Services returns the services configured for Lair.
func (c *Config) Services(baseLogger log.Logger, noRPC bool) (*Services, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "ghost",
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
	// Include random keys in feed list to enable broadcasting.
	for _, v := range c.Ethereum.RandKeys {
		c.Transport.LibP2P.Feeds = append(c.Transport.LibP2P.Feeds, keys[v].Address())
	}
	transport, err := c.Transport.Transport(transportConfig.Dependencies{
		Keys:    keys,
		Clients: clients,
		Messages: map[string]pkgTransport.Message{
			messages.PriceV0MessageName:     (*messages.Price)(nil),
			messages.PriceV1MessageName:     (*messages.Price)(nil),
			messages.DataPointV1MessageName: (*messages.DataPoint)(nil),
		},
		Logger: logger,
	})
	if err != nil {
		return nil, err
	}
	gofer, err := c.Gofer.PriceProvider(priceproviderConfig.Dependencies{
		Clients: clients,
		Logger:  logger,
	}, noRPC)
	if err != nil {
		return nil, err
	}
	ghost, err := c.Ghost.Feed(feedConfig.Dependencies{
		KeysRegistry:  keys,
		PriceProvider: gofer,
		Transport:     transport,
		Logger:        logger,
	})
	if err != nil {
		return nil, err
	}
	return &Services{
		Feed:      ghost,
		Transport: transport,
		Logger:    logger,
	}, nil
}
