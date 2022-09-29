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
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/replayer"
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
	Ethereum           ethereumConfig.Ethereum `yaml:"ethereum"`
	Interval           int64                   `yaml:"interval"`
	PrefetchPeriod     int64                   `yaml:"prefetchPeriod"`
	BlockConfirmations int64                   `yaml:"blockConfirmations"`
	BlockLimit         int                     `yaml:"blockLimit"`
	ReplayAfter        []int64                 `yaml:"replayAfter"`
	Addresses          []common.Address        `yaml:"addresses"`
}

type teleportStarknetListener struct {
	Sequencer      string                 `yaml:"sequencer"`
	Interval       int64                  `yaml:"interval"`
	PrefetchPeriod int64                  `yaml:"prefetchPeriod"`
	ReplayAfter    []int64                `yaml:"replayAfter"`
	Addresses      []*starknetClient.Felt `yaml:"addresses"`
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
	var eps []publisher.EventProvider
	if err := c.configureTeleportEVM(&eps, d.Logger); err != nil {
		return nil, fmt.Errorf("eventpublisher config: teleport EVM: %w", err)
	}
	if err := c.configureTeleportStarknet(&eps, d.Logger); err != nil {
		return nil, fmt.Errorf("eventpublisher config: teleport Starknet: %w", err)
	}
	signer := []publisher.EventSigner{teleportevm.NewSigner(d.Signer, []string{
		teleportevm.TeleportEventType,
		teleportstarknet.TeleportEventType,
	})}
	cfg := publisher.Config{
		Providers: eps,
		Signers:   signer,
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
	clients := ethClients{}
	for _, cfg := range c.Listeners.TeleportEVM {
		client, err := clients.configure(cfg.Ethereum, logger)
		if err != nil {
			return err
		}
		interval := cfg.Interval
		if interval < 1 {
			interval = 1
		}
		if cfg.BlockLimit == 0 {
			cfg.BlockLimit = 1000
		}
		replayAfter := make([]time.Duration, len(cfg.ReplayAfter))
		for i, r := range cfg.ReplayAfter {
			replayAfter[i] = time.Duration(r) * time.Second
		}
		var ep publisher.EventProvider
		ep, err = teleportevm.New(teleportevm.Config{
			Client:             client,
			Addresses:          cfg.Addresses,
			Interval:           time.Second * time.Duration(interval),
			PrefetchPeriod:     time.Duration(cfg.PrefetchPeriod) * time.Second,
			BlockLimit:         uint64(cfg.BlockLimit),
			BlockConfirmations: uint64(cfg.BlockConfirmations),
			Logger:             logger,
		})
		if err != nil {
			return err
		}
		if len(cfg.ReplayAfter) > 0 {
			ep, err = replayer.New(replayer.Config{
				EventProvider: ep,
				Interval:      time.Minute,
				ReplayAfter:   replayAfter,
			})
		}
		if err != nil {
			return err
		}
		*lis = append(*lis, ep)
	}
	return nil
}

func (c *EventPublisher) configureTeleportStarknet(lis *[]publisher.EventProvider, logger log.Logger) error {
	var err error
	for _, cfg := range c.Listeners.TeleportStarknet {
		interval := cfg.Interval
		if interval < 1 {
			interval = 1
		}
		if _, err := url.Parse(cfg.Sequencer); err != nil {
			return fmt.Errorf("sequencer url is invalid: %w", err)
		}
		replayAfter := make([]time.Duration, len(cfg.ReplayAfter))
		for i, r := range cfg.ReplayAfter {
			replayAfter[i] = time.Duration(r) * time.Second
		}
		var ep publisher.EventProvider
		ep, err = teleportstarknet.New(teleportstarknet.Config{
			Sequencer: starknetClient.NewSequencer(cfg.Sequencer, http.Client{}),
			Addresses: cfg.Addresses,
			Interval:  time.Second * time.Duration(interval),
			Logger:    logger,
		})
		if err != nil {
			return err
		}
		if len(cfg.ReplayAfter) > 0 {
			ep, err = replayer.New(replayer.Config{
				EventProvider: ep,
				Interval:      time.Minute,
				ReplayAfter:   replayAfter,
			})
			if err != nil {
				return err
			}
		}
		*lis = append(*lis, ep)
	}
	return nil
}

type ethClients map[string]ethereum.Client

// configure returns an Ethereum client for given configuration.
// It will return the same instance of the client for the same
// configuration.
func (m ethClients) configure(ethereum ethereumConfig.Ethereum, logger log.Logger) (ethereum.Client, error) {
	key, err := json.Marshal(ethereum)
	if err != nil {
		return nil, err
	}
	if c, ok := m[string(key)]; ok {
		return c, nil
	}
	c, err := ethereum.ConfigureEthereumClient(nil, logger)
	if err != nil {
		return nil, err
	}
	m[string(key)] = c
	return c, nil
}
