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

package eventpublisher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportevm"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportstarknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	starknetClient "github.com/chronicleprotocol/oracle-suite/pkg/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var eventPublisherFactory = func(cfg publisher.Config) (*publisher.EventPublisher, error) {
	return publisher.New(cfg)
}

type EventPublisher struct {
	Listeners listeners `yaml:"listeners"`
}

type listeners struct {
	TeleportEVM      []teleportEVMListener      `yaml:"teleportEVM"`
	TeleportStarknet []teleportStarknetListener `yaml:"teleportStarknet"`
}

type teleportEVMListener struct {
	Ethereum    ethereumConfig.Ethereum `yaml:"ethereum"`
	Interval    int64                   `yaml:"interval"`
	BlocksDelta []int                   `yaml:"blocksDelta"`
	BlocksLimit int                     `yaml:"blocksLimit"`
	Addresses   []common.Address        `yaml:"addresses"`
}

type teleportStarknetListener struct {
	Sequencer   string                 `yaml:"sequencer"`
	Interval    int64                  `yaml:"interval"`
	BlocksDelta []int                  `yaml:"blocksDelta"`
	BlocksLimit int                    `yaml:"blocksLimit"`
	Addresses   []*starknetClient.Felt `yaml:"addresses"`
}

type Dependencies struct {
	Signer    ethereum.Signer
	Transport transport.Transport
	Logger    log.Logger
}

func (c *EventPublisher) Configure(d Dependencies) (*publisher.EventPublisher, error) {
	if d.Signer == nil {
		return nil, fmt.Errorf("eventpublisher config: signer cannot be nil")
	}
	if d.Transport == nil {
		return nil, fmt.Errorf("eventpublisher config: transport cannot be nil")
	}
	if d.Logger == nil {
		return nil, fmt.Errorf("eventpublisher config: logger cannot be nil")
	}
	var lis []publisher.EventProvider
	if err := c.configureTeleportEVM(&lis, d.Logger); err != nil {
		return nil, fmt.Errorf("eventpublisher config: %w", err)
	}
	if err := c.configureTeleportStarknet(&lis, d.Logger); err != nil {
		return nil, fmt.Errorf("eventpublisher config: %w", err)
	}
	sig := []publisher.Signer{teleportevm.NewSigner(d.Signer, []string{
		teleportevm.TeleportEventType,
		teleportstarknet.TeleportEventType,
	})}
	cfg := publisher.Config{
		Listeners: lis,
		Signers:   sig,
		Transport: d.Transport,
		Logger:    d.Logger,
	}
	ep, err := eventPublisherFactory(cfg)
	if err != nil {
		return nil, fmt.Errorf("eventpublisher config: %w", err)
	}
	return ep, nil
}

func (c *EventPublisher) configureTeleportEVM(lis *[]publisher.EventProvider, logger log.Logger) error {
	clis := ethClients{}
	for _, w := range c.Listeners.TeleportEVM {
		cli, err := clis.configure(w.Ethereum, logger)
		if err != nil {
			return err
		}
		interval := w.Interval
		if interval < 1 {
			interval = 1
		}
		if len(w.BlocksDelta) < 1 {
			return fmt.Errorf("blocksDelta must contains at least one element")
		}
		if w.BlocksLimit <= 0 {
			return fmt.Errorf("blocksLimit must greather than 0")
		}
		*lis = append(*lis, teleportevm.New(teleportevm.TeleportEventProviderConfig{
			Client:      cli,
			Addresses:   w.Addresses,
			Interval:    time.Second * time.Duration(interval),
			BlocksDelta: w.BlocksDelta,
			BlocksLimit: w.BlocksLimit,
			Logger:      logger,
		}))
	}
	return nil
}

func (c *EventPublisher) configureTeleportStarknet(lis *[]publisher.EventProvider, logger log.Logger) error {
	for _, w := range c.Listeners.TeleportStarknet {
		interval := w.Interval
		if interval < 1 {
			interval = 1
		}
		if _, err := url.Parse(w.Sequencer); err != nil {
			return fmt.Errorf("sequencer address is not valid url: %w", err)
		}
		if len(w.BlocksDelta) < 1 {
			return fmt.Errorf("blocksDelta must contains at least one element")
		}
		if w.BlocksLimit <= 0 {
			return fmt.Errorf("blocksLimit must greather than 0")
		}
		*lis = append(*lis, teleportstarknet.New(teleportstarknet.TeleportEventProviderConfig{
			Sequencer:   starknetClient.NewSequencer(w.Sequencer, http.Client{}),
			Addresses:   w.Addresses,
			Interval:    time.Second * time.Duration(interval),
			BlocksDelta: w.BlocksDelta,
			BlocksLimit: w.BlocksLimit,
			Logger:      logger,
		}))
	}
	return nil
}

type ethClients map[string]geth.EthClient

// configure returns an Ethereum client for given configuration.
// It will return the same instance of the client for the same
// configuration.
func (m ethClients) configure(ethereum ethereumConfig.Ethereum, logger log.Logger) (geth.EthClient, error) {
	key, err := json.Marshal(ethereum)
	if err != nil {
		return nil, err
	}
	if c, ok := m[string(key)]; ok {
		return c, nil
	}
	c, err := ethereum.ConfigureRPCClient(logger)
	if err != nil {
		return nil, err
	}
	m[string(key)] = c
	return c, nil
}
