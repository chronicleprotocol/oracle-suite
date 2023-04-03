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
	Ethereum  *ethereumConfig.Config `hcl:"ethereum,block"`
	Transport transportConfig.Config `hcl:"transport,block"`
	Logger    *loggerConfig.Config   `hcl:"logger,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	Transport  pkgTransport.Transport
	EventStore *store.EventStore
	EventAPI   *api.EventAPI
	Logger     log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
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
