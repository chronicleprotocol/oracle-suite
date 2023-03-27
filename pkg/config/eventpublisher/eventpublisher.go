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
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/defiweb/go-eth/types"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/replayer"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportevm"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher/teleportstarknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	starknetClient "github.com/chronicleprotocol/oracle-suite/pkg/starknet"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

type Dependencies struct {
	KeyRegistry    ethereumConfig.KeyRegistry
	ClientRegistry ethereumConfig.ClientRegistry
	Transport      transport.Transport
	Logger         log.Logger
}

type ConfigEventPublisher struct {
	// EthereumKey is the name of the key to use for signing events.
	EthereumKey string `hcl:"ethereum_key"`

	// TeleportEVM is a list of Teleport listeners for EVM-compatible chains.
	TeleportEVM []teleportEVMListener `hcl:"teleport_evm,block"`

	// TeleportStarknet is a list of Teleport listeners for Starknet.
	TeleportStarknet []teleportStarknetListener `hcl:"teleport_starknet,block"`

	// Configured service:
	eventPublisher *publisher.EventPublisher
}

type teleportEVMListener struct {
	// EthereumClient is the name of the Ethereum client to use for
	// listening to events.
	EthereumClient string `hcl:"ethereum_client"`

	// Interval specifies how often, in seconds, the event listener should
	// check for new events.
	Interval int64 `hcl:"interval"`

	// PrefetchPeriod specifies how far, in seconds, the event listener should
	// check for new  events during the initial synchronization.
	PrefetchPeriod int64 `hcl:"prefetch_period"`

	// BlockConfirmations is the number of blocks to wait before
	// considering a block final.
	BlockConfirmations int64 `hcl:"block_confirmations"`

	// BlockLimit is the maximum range of blocks to fetch in a single
	// filter log request.
	BlockLimit int `hcl:"block_limit"`

	// ReplayAfter specifies after which time, in seconds, the event listener
	// should replay events. It is used to guarantee that events are eventually
	// delivered to subscribers even if they are not online at the time the event
	// was published.
	ReplayAfter []int64 `hcl:"replay_after"`

	// ContractAddrs is a list of teleport contract addresses to listen
	// to.
	ContractAddrs []string `hcl:"contract_addrs"`
}

type teleportStarknetListener struct {
	// Sequencer is the name of the Starknet sequencer to use for listening
	// to events.
	Sequencer string `hcl:"sequencer"`

	// Interval specifies how often, in seconds, the event listener should
	// check for new events.
	Interval int64 `hcl:"interval"`

	// PrefetchPeriod specifies how far, in seconds, the event listener should
	// check for new  events during the initial synchronization.
	PrefetchPeriod int64 `hcl:"prefetch_period"`

	// ReplayAfter specifies after which time, in seconds, the event listener
	// should replay events. It is used to guarantee that events are eventually
	// delivered to subscribers even if they are not online at the time the event
	// was published.
	ReplayAfter []int64 `hcl:"replay_after"`

	// ContractAddrs is a list of teleport contract addresses to listen
	// to.
	ContractAddrs []string `hcl:"contract_addrs"`
}

func (c *ConfigEventPublisher) EventPublisher(d Dependencies) (*publisher.EventPublisher, error) {
	if c.eventPublisher != nil {
		return c.eventPublisher, nil
	}
	if d.Transport == nil {
		return nil, fmt.Errorf("eventpublisher config: transport cannot be nil")
	}
	if d.Logger == nil {
		return nil, fmt.Errorf("eventpublisher config: logger cannot be nil")
	}
	var eventProviders []publisher.EventProvider
	if err := c.teleportEVM(&eventProviders, d); err != nil {
		return nil, fmt.Errorf("eventpublisher config: teleport EVM: %w", err)
	}
	if err := c.teleportStarknet(&eventProviders, d); err != nil {
		return nil, fmt.Errorf("eventpublisher config: teleport Starknet: %w", err)
	}
	key, ok := d.KeyRegistry[c.EthereumKey]
	if !ok {
		return nil, fmt.Errorf("spire config: ethereum key %q not found", c.EthereumKey)
	}
	signer := []publisher.EventSigner{teleportevm.NewSigner(key, []string{
		teleportevm.TeleportEventType,
		teleportstarknet.TeleportEventType,
	})}
	eventPublisher, err := publisher.New(publisher.Config{
		Providers: eventProviders,
		Signers:   signer,
		Transport: d.Transport,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("eventpublisher config: %w", err)
	}
	c.eventPublisher = eventPublisher
	return eventPublisher, nil
}

func (c *ConfigEventPublisher) teleportEVM(eps *[]publisher.EventProvider, d Dependencies) error {
	var err error
	for _, cfg := range c.TeleportEVM {
		client, ok := d.ClientRegistry[cfg.EthereumClient]
		if !ok {
			return fmt.Errorf("ethereum client %q not found", cfg.EthereumClient)
		}
		interval := cfg.Interval
		if interval <= 0 {
			return fmt.Errorf("interval must be greater than 0")
		}
		if cfg.PrefetchPeriod < 0 {
			return fmt.Errorf("prefetch period must be greater or equal to 0")
		}
		if cfg.BlockConfirmations < 0 {
			return fmt.Errorf("block confirmations must be greater or equal to 0")
		}
		if cfg.BlockLimit <= 0 {
			return fmt.Errorf("block limit must be greater than 0")
		}
		if len(cfg.ContractAddrs) == 0 {
			return fmt.Errorf("contract addresses cannot be empty")
		}
		contractAddrs := make([]types.Address, len(cfg.ContractAddrs))
		for i, addr := range cfg.ContractAddrs {
			contractAddrs[i], err = types.AddressFromHex(addr)
			if err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}
		}
		replayAfter := make([]time.Duration, len(cfg.ReplayAfter))
		for i, r := range cfg.ReplayAfter {
			replayAfter[i] = time.Duration(r) * time.Second
		}
		var eventProvider publisher.EventProvider
		eventProvider, err = teleportevm.New(teleportevm.Config{
			Client:             geth.NewClient(client), //nolint:staticcheck // deprecated ethereum.Client
			Addresses:          contractAddrs,
			Interval:           time.Second * time.Duration(interval),
			PrefetchPeriod:     time.Duration(cfg.PrefetchPeriod) * time.Second,
			BlockLimit:         uint64(cfg.BlockLimit),
			BlockConfirmations: uint64(cfg.BlockConfirmations),
			Logger:             d.Logger,
		})
		if err != nil {
			return err
		}
		if len(cfg.ReplayAfter) > 0 {
			eventProvider, err = replayer.New(replayer.Config{
				EventProvider: eventProvider,
				Interval:      time.Minute,
				ReplayAfter:   replayAfter,
			})
		}
		if err != nil {
			return err
		}
		*eps = append(*eps, eventProvider)
	}
	return nil
}

func (c *ConfigEventPublisher) teleportStarknet(eps *[]publisher.EventProvider, d Dependencies) error {
	var err error
	for _, cfg := range c.TeleportStarknet {
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
		contractAddrs := make([]*starknetClient.Felt, len(cfg.ContractAddrs))
		for i, addr := range cfg.ContractAddrs {
			contractAddrs[i] = starknetClient.HexToFelt(addr)
		}
		var eventProvider publisher.EventProvider
		eventProvider, err = teleportstarknet.New(teleportstarknet.Config{
			Sequencer:      starknetClient.NewSequencer(cfg.Sequencer, http.Client{}),
			Addresses:      contractAddrs,
			Interval:       time.Second * time.Duration(interval),
			PrefetchPeriod: time.Duration(cfg.PrefetchPeriod) * time.Second,
			Logger:         d.Logger,
		})
		if err != nil {
			return err
		}
		if len(cfg.ReplayAfter) > 0 {
			eventProvider, err = replayer.New(replayer.Config{
				EventProvider: eventProvider,
				Interval:      time.Minute,
				ReplayAfter:   replayAfter,
			})
			if err != nil {
				return err
			}
		}
		*eps = append(*eps, eventProvider)
	}
	return nil
}
