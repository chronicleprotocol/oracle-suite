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
	"time"

	"github.com/ethereum/go-ethereum/common"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/internal/config/ethereum"
	starknetClient "github.com/chronicleprotocol/oracle-suite/internal/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher"
	publisherEthereum "github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
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
	Wormhole         []wormholeListener         `json:"wormhole"`
	WormholeStarknet []wormholeStarknetListener `json:"wormholeStarknet"`
}

type wormholeListener struct {
	Ethereum     ethereumConfig.Ethereum `json:"ethereum"`
	Interval     int64                   `json:"interval"`
	BlocksBehind []int                   `json:"blocksBehind"`
	MaxBlocks    int                     `json:"maxBlocks"`
	Addresses    []common.Address        `json:"addresses"`
}

type wormholeStarknetListener struct {
	RPC          string                 `json:"rpc"`
	Interval     int64                  `json:"interval"`
	BlocksBehind []int                  `json:"blocksBehind"`
	MaxBlocks    int                    `json:"maxBlocks"`
	Addresses    []*starknetClient.Felt `json:"addresses"`
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
	if err := c.configureWormholeListeners(&lis, d.Logger); err != nil {
		return nil, fmt.Errorf("eventpublisher config: %w", err)
	}
	c.configureWormholeStarknetListeners(&lis, d.Logger)
	sig := []publisher.Signer{publisherEthereum.NewSigner(d.Signer, []string{publisherEthereum.WormholeEventType})}
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

func (c *EventPublisher) configureWormholeListeners(lis *[]publisher.Listener, logger log.Logger) error {
	clis := ethClients{}
	for _, w := range c.Listeners.Wormhole {
		cli, err := clis.configure(w.Ethereum, logger)
		if err != nil {
			return err
		}
		interval := w.Interval
		if interval < 1 {
			interval = 1
		}
		if len(w.BlocksBehind) < 1 {
			return fmt.Errorf("blocksBehind must contains at least one element")
		}
		if w.MaxBlocks <= 0 {
			return fmt.Errorf("maxBlocks must greather than 0")
		}
		for _, blocksBehind := range w.BlocksBehind {
			*lis = append(*lis, publisherEthereum.NewWormholeListener(publisherEthereum.WormholeListenerConfig{
				Client:       cli,
				Addresses:    w.Addresses,
				Interval:     time.Second * time.Duration(interval),
				BlocksBehind: blocksBehind,
				MaxBlocks:    w.MaxBlocks,
				Logger:       logger,
			}))
		}
	}
	return nil
}

func (c *EventPublisher) configureWormholeStarknetListeners(lis *[]publisher.Listener, logger log.Logger) {
	for _, w := range c.Listeners.WormholeStarknet {
		*lis = append(*lis, starknet.NewWormholeListener(starknet.WormholeListenerConfig{
			Sequencer:    starknetClient.NewSequencer(w.RPC, http.Client{}),
			Addresses:    w.Addresses,
			Interval:     time.Second * time.Duration(w.Interval),
			BlocksBehind: w.BlocksBehind,
			MaxBlocks:    w.MaxBlocks,
			Logger:       logger,
		}))
	}
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
