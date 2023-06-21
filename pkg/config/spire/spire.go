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

package spire

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	pkgTransport "github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// Config is the configuration for Spire.
type Config struct {
	Spire     ConfigSpire            `hcl:"spire,block"`
	Transport transportConfig.Config `hcl:"transport,block"`
	Ethereum  ethereumConfig.Config  `hcl:"ethereum,block"`
	Logger    *loggerConfig.Config   `hcl:"logger,block,optional"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

type ConfigSpire struct {
	// RPCListenAddr is an address to listen for RPC requests.
	RPCListenAddr string `hcl:"rpc_listen_addr"`

	// RPCAgentAddr is an address of the agent to connect to.
	RPCAgentAddr string `hcl:"rpc_agent_addr"`

	// Pairs is a list of pairs to store in the price store.
	Pairs []string `hcl:"pairs"`

	// EthereumKey is a name of an Ethereum key to use for signing
	// prices.
	EthereumKey string `hcl:"ethereum_key,optional"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured services:
	agent      *spire.Agent
	client     *spire.Client
	priceStore *store.PriceStore
}

// ClientServices returns the services that are configured from the Config struct.
type ClientServices struct {
	SpireClient *spire.Client
	Logger      log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// AgentServices returns the services that are configured from the Config struct.
type AgentServices struct {
	SpireAgent *spire.Agent
	Transport  pkgTransport.Transport
	PriceStore *store.PriceStore
	Logger     log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// StreamServices returns the services that are configured from the Config struct.
type StreamServices struct {
	Transport pkgTransport.Transport
	Logger    log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *ClientServices) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.SpireClient, sysmon.New(time.Minute, s.Logger))
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
	s.supervisor.Watch(s.Transport, s.PriceStore, s.SpireAgent, sysmon.New(time.Minute, s.Logger))
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *AgentServices) Wait() <-chan error {
	return s.supervisor.Wait()
}

// Start implements the supervisor.Service interface.
func (s *StreamServices) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.Transport, sysmon.New(time.Minute, s.Logger))
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *StreamServices) Wait() <-chan error {
	return s.supervisor.Wait()
}

// ClientServices returns the services configured for Spire.
func (c *Config) ClientServices(baseLogger log.Logger) (*ClientServices, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "spire",
		BaseLogger: baseLogger,
	})
	if err != nil {
		return nil, err
	}
	keys, err := c.Ethereum.KeyRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, err
	}
	client, err := c.Spire.ConfigureClient(keys)
	if err != nil {
		return nil, err
	}
	return &ClientServices{
		SpireClient: client,
		Logger:      logger,
	}, nil
}

// AgentServices returns the services configured for Spire.
func (c *Config) AgentServices(baseLogger log.Logger) (*AgentServices, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "spire",
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
	priceStore, err := c.Spire.PriceStore(logger, transport)
	if err != nil {
		return nil, err
	}
	spireAgent, err := c.Spire.ConfigureAgent(logger, transport, priceStore)
	if err != nil {
		return nil, err
	}
	return &AgentServices{
		SpireAgent: spireAgent,
		Transport:  transport,
		PriceStore: priceStore,
		Logger:     logger,
	}, nil
}

// StreamServices returns the services configured for Spire.
func (c *Config) StreamServices(baseLogger log.Logger) (*StreamServices, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    "spire",
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
	return &StreamServices{
		Transport: transport,
		Logger:    logger,
	}, nil
}

func (c *ConfigSpire) ConfigureAgent(
	logger log.Logger,
	transport pkgTransport.Transport,
	priceStore *store.PriceStore,
) (
	*spire.Agent, error,
) {

	if c.agent != nil {
		return c.agent, nil
	}
	agent, err := spire.NewAgent(spire.AgentConfig{
		PriceStore: priceStore,
		Transport:  transport,
		Address:    c.RPCListenAddr,
		Logger:     logger,
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Spire agent: %v", err),
			Subject:  &c.Range,
		}
	}
	c.agent = agent
	return agent, nil
}

func (c *ConfigSpire) ConfigureClient(kr ethereumConfig.KeyRegistry) (*spire.Client, error) {
	if c.client != nil {
		return c.client, nil
	}
	signer := kr[c.EthereumKey] // Signer may be nil.
	client, err := spire.NewClient(spire.ClientConfig{
		Signer:  signer,
		Address: c.RPCAgentAddr,
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Spire client: %v", err),
			Subject:  &c.Range,
		}
	}
	c.client = client
	return client, nil
}

func (c *ConfigSpire) PriceStore(l log.Logger, t pkgTransport.Transport) (*store.PriceStore, error) {
	if c.priceStore != nil {
		return c.priceStore, nil
	}
	priceStore, err := store.New(store.Config{
		Storage:   store.NewMemoryStorage(),
		Transport: t,
		Pairs:     c.Pairs,
		Logger:    l,
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Price Store service: %v", err),
			Subject:  &c.Range,
		}
	}
	c.priceStore = priceStore
	return priceStore, nil
}
