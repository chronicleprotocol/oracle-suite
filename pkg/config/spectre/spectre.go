package spectre

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	relayConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/relay"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/relayer"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	pkgTransport "github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// Config is the configuration for Spectre.
type Config struct {
	Spectre   relayConfig.Config     `hcl:"spectre,block"`
	Transport transportConfig.Config `hcl:"transport,block"`
	Ethereum  ethereumConfig.Config  `hcl:"ethereum,block"`
	Logger    *loggerConfig.Config   `hcl:"logger,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	Relay      *relayer.Relayer
	PriceStore *store.PriceStore
	Transport  pkgTransport.Transport
	Logger     log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.Transport, s.PriceStore, s.Relay, sysmon.New(time.Minute, s.Logger))
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *Services) Wait() <-chan error {
	return s.supervisor.Wait()
}

// Services returns the services configured for Spectre.
func (c *Config) Services(baseLogger log.Logger) (*Services, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "spectre",
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
		Keys:    keys,
		Clients: clients,
		Messages: map[string]pkgTransport.Message{
			messages.PriceV0MessageName: (*messages.Price)(nil),
			messages.PriceV1MessageName: (*messages.Price)(nil),
		},
		Logger: logger,
	})
	if err != nil {
		return nil, err
	}
	priceStore, err := c.Spectre.PriceStore(relayConfig.PriceStoreDependencies{
		Transport: transport,
		Logger:    logger,
	})
	if err != nil {
		return nil, err
	}
	relay, err := c.Spectre.Relay(relayConfig.Dependencies{
		Clients:    clients,
		PriceStore: priceStore,
		Logger:     logger,
	})
	if err != nil {
		return nil, err
	}
	return &Services{
		Relay:      relay,
		PriceStore: priceStore,
		Transport:  transport,
		Logger:     logger,
	}, nil
}
