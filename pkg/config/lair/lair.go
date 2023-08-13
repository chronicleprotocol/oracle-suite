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

package lair

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	eventAPIConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/eventapi"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/api"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportevm"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportstarknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	pkgTransport "github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// Config is the configuration for Lair.
type Config struct {
	EventAPI  eventAPIConfig.Config  `hcl:"lair,block"`
	Ethereum  *ethereumConfig.Config `hcl:"ethereum,block,optional"`
	Transport transportConfig.Config `hcl:"transport,block"`
	Logger    *loggerConfig.Config   `hcl:"logger,block,optional"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	Transport  pkgTransport.Service
	EventStore *store.EventStore
	EventAPI   *api.EventAPI
	Logger     log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.Transport, s.EventStore, s.EventAPI, sysmon.New(time.Minute, s.Logger))
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
func (c *Config) Services(baseLogger log.Logger) (*Services, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "lair",
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
	transport, err := c.Transport.Transport(transportConfig.Dependencies{
		Clients: clients,
		Keys:    keys,
		Logger:  logger,
		Messages: map[string]pkgTransport.Message{
			messages.EventV1MessageName: (*messages.Event)(nil),
		},
	})
	if err != nil {
		return nil, err
	}
	storage, err := c.EventAPI.Storage()
	if err != nil {
		return nil, err
	}
	eventStore, err := store.New(store.Config{
		EventTypes: []string{teleportevm.TeleportEventType, teleportstarknet.TeleportEventType},
		Storage:    storage,
		Transport:  transport,
		Logger:     logger,
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Event Store service: %v", err),
			Subject:  c.EventAPI.Range.Ptr(),
		}
	}
	eventAPI, err := c.EventAPI.EventAPI(eventAPIConfig.Dependencies{
		EventStore: eventStore,
		Transport:  transport,
		Logger:     logger,
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Event API service: %v", err),
			Subject:  c.EventAPI.Range.Ptr(),
		}
	}
	return &Services{
		Transport:  transport,
		EventStore: eventStore,
		EventAPI:   eventAPI,
		Logger:     logger,
	}, nil
}
