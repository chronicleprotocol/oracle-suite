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

package spectre

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	relayConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/relay"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/relay"

	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// Config is the configuration for Spectre.
type Config struct {
	Spectre   relayConfig.Config     `hcl:"spectre,block"`
	Transport transportConfig.Config `hcl:"transport,block"`
	Ethereum  ethereumConfig.Config  `hcl:"ethereum,block"`
	Logger    *loggerConfig.Config   `hcl:"logger,block,optional"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	Relay      *relay.Relay
	PriceStore *store.Store
	MuSigStore *relay.MuSigStore
	Transport  transport.Service
	Logger     log.Logger

	supervisor *supervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = supervisor.New(s.Logger)
	s.supervisor.Watch(
		s.Transport,
		s.PriceStore,
		s.MuSigStore,
		s.Relay,
	)
	if l, ok := s.Logger.(supervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *Services) Wait() <-chan error {
	return s.supervisor.Wait()
}

// Services returns the services configured for Spectre.
func (c *Config) Services(baseLogger log.Logger, appName string, appVersion string) (supervisor.Service, error) {
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
	messageMap, err := transport.AllMessagesMap.SelectByTopic(
		messages.DataPointV1MessageName,
		messages.MuSigStartV1MessageName,
		messages.MuSigTerminateV1MessageName,
		messages.MuSigCommitmentV1MessageName,
		messages.MuSigPartialSignatureV1MessageName,
		messages.MuSigSignatureV1MessageName,
		messages.MuSigOptimisticSignatureV1MessageName,
	)
	if err != nil {
		return nil, err
	}
	transportSrv, err := c.Transport.Transport(transportConfig.Dependencies{
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
	srvs, err := c.Spectre.Relay(relayConfig.Dependencies{
		Clients:   clients,
		Transport: transportSrv,
		Logger:    logger,
	})
	if err != nil {
		return nil, err
	}
	return &Services{
		Relay:      srvs.Relay,
		PriceStore: srvs.PriceStore,
		MuSigStore: srvs.MuSigStore,
		Transport:  transportSrv,
		Logger:     logger,
	}, nil
}
