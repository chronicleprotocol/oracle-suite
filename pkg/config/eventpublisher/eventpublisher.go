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
	publisherEthereum "github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/ethereum"
	publisherStarknet "github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	starknetClient "github.com/chronicleprotocol/oracle-suite/pkg/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var eventPublisherFactory = func(cfg publisher.Config) (*publisher.EventPublisher, error) {
	return publisher.New(cfg)
}

type EventPublisher struct {
	Listeners listeners `json:"listeners"`
}

type listeners struct {
	TeleportEVM      []teleportEVMListener      `json:"teleportEVM"`
	TeleportStarknet []teleportStarknetListener `json:"teleportStarknet"`
}

type teleportEVMListener struct {
	Ethereum    ethereumConfig.Ethereum `json:"ethereum"`
	Interval    int64                   `json:"interval"`
	BlocksDelta []int                   `json:"blocksDelta"`
	BlocksLimit int                     `json:"blocksLimit"`
	Addresses   []common.Address        `json:"addresses"`
}

type teleportStarknetListener struct {
	Sequencer   string                 `json:"sequencer"`
	Interval    int64                  `json:"interval"`
	BlocksDelta []int                  `json:"blocksDelta"`
	BlocksLimit int                    `json:"blocksLimit"`
	Addresses   []*starknetClient.Felt `json:"addresses"`
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
	var lis []publisher.Listener
	if err := c.configureTeleportEVMListeners(&lis, d.Logger); err != nil {
		return nil, fmt.Errorf("eventpublisher config: %w", err)
	}
	if err := c.configureTeleportStarknetListeners(&lis, d.Logger); err != nil {
		return nil, fmt.Errorf("eventpublisher config: %w", err)
	}
	sig := []publisher.Signer{publisherEthereum.NewSigner(d.Signer, []string{
		publisherEthereum.TeleportEventType,
		publisherStarknet.TeleportEventType,
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

func (c *EventPublisher) configureTeleportEVMListeners(lis *[]publisher.Listener, logger log.Logger) error {
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
		*lis = append(*lis, publisherEthereum.NewTeleportListener(publisherEthereum.TeleportListenerConfig{
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

func (c *EventPublisher) configureTeleportStarknetListeners(lis *[]publisher.Listener, logger log.Logger) error {
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
		*lis = append(*lis, publisherStarknet.NewTeleportListener(publisherStarknet.TeleportListenerConfig{
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
