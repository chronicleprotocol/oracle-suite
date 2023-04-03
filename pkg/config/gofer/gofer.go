package gofer

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	priceProviderConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/priceprovider"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/marshal"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/rpc"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"

	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

// Config is the configuration for Gofer.
type Config struct {
	Gofer    priceProviderConfig.Config `hcl:"gofer,block"`
	Ethereum *ethereumConfig.Config     `hcl:"ethereum,block"`
	Logger   *loggerConfig.Config       `hcl:"logger,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

// ClientServices returns the services that are configured from the Config struct.
type ClientServices struct {
	PriceProvider provider.Provider
	PriceHook     provider.PriceHook
	Marshaller    marshal.Marshaller
	Logger        log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// AgentServices returns the services that are configured from the Config struct.
type AgentServices struct {
	PriceProvider provider.Provider
	Agent         *rpc.Agent
	Logger        log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *ClientServices) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	if p, ok := s.PriceProvider.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(p)
	}
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *ClientServices) Wait() <-chan error {
	return s.supervisor.Wait()
}

// Start implements the supervisor.Service interface.
func (s *AgentServices) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.Agent, sysmon.New(time.Minute, s.Logger))
	if p, ok := s.PriceProvider.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(p)
	}
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *AgentServices) Wait() <-chan error {
	return s.supervisor.Wait()
}

// ClientServices returns the services configured for Gofer.
func (c *Config) ClientServices(
	ctx context.Context,
	baseLogger log.Logger,
	noRPC bool,
	format marshal.FormatType,
) (*ClientServices, error) {

	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "gofer",
		BaseLogger: baseLogger,
	})
	if err != nil {
		return nil, err
	}
	clients, err := c.Ethereum.ClientRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, err
	}
	priceProvider, err := c.Gofer.PriceProvider(priceProviderConfig.Dependencies{
		Clients: clients,
		Logger:  logger,
	}, noRPC)
	if err != nil {
		return nil, err
	}
	hook, err := c.Gofer.PriceHook(priceProviderConfig.HookDependencies{
		Context: ctx,
		Clients: clients,
	})
	if err != nil {
		return nil, err
	}
	marshaler, err := marshal.NewMarshal(format)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create Gofer price marshaler: %v", err),
			Subject:  c.Gofer.Range.Ptr(),
		}
	}
	return &ClientServices{
		PriceProvider: priceProvider,
		PriceHook:     hook,
		Marshaller:    marshaler,
		Logger:        logger,
	}, nil
}

// AgentServices returns the services configured for Gofer Agent.
func (c *Config) AgentServices(baseLogger log.Logger) (*AgentServices, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "gofer",
		BaseLogger: baseLogger,
	})
	if err != nil {
		return nil, err
	}
	clients, err := c.Ethereum.ClientRegistry(ethereumConfig.Dependencies{Logger: baseLogger})
	if err != nil {
		return nil, err
	}
	priceProvider, err := c.Gofer.AsyncPriceProvider(priceProviderConfig.AsyncDependencies{
		Clients: clients,
		Logger:  baseLogger,
	})
	if err != nil {
		return nil, err
	}
	agent, err := c.Gofer.RPCAgent(priceProviderConfig.AgentDependencies{
		Provider: priceProvider,
		Logger:   logger,
	})
	if err != nil {
		return nil, err
	}
	return &AgentServices{
		PriceProvider: priceProvider,
		Agent:         agent,
		Logger:        logger,
	}, nil
}
