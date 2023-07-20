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

package eventpublisher

import (
	"fmt"
	"net/http"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
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
	Keys      ethereumConfig.KeyRegistry
	Clients   ethereumConfig.ClientRegistry
	Transport transport.Service
	Logger    log.Logger
}

type Config struct {
	// EthereumKey is the name of the key to use for signing events.
	EthereumKey string `hcl:"ethereum_key"`

	// TeleportEVM is a list of Teleport listeners for EVM-compatible chains.
	TeleportEVM []teleportEVMListener `hcl:"teleport_evm,block"`

	// TeleportStarknet is a list of Teleport listeners for Starknet.
	TeleportStarknet []teleportStarknetListener `hcl:"teleport_starknet,block"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured service:
	eventPublisher *publisher.EventPublisher
}

type teleportEVMListener struct {
	// EthereumClient is the name of the Ethereum client to use for
	// listening to events.
	EthereumClient string `hcl:"ethereum_client"`

	// Interval specifies how often, in seconds, the event listener should
	// check for new events.
	Interval uint32 `hcl:"interval"`

	// PrefetchPeriod specifies how far, in seconds, the event listener should
	// check for new  events during the initial synchronization.
	PrefetchPeriod uint64 `hcl:"prefetch_period"`

	// BlockConfirmations is the number of blocks to wait before
	// considering a block final.
	BlockConfirmations uint64 `hcl:"block_confirmations"`

	// BlockLimit is the maximum range of blocks to fetch in a single
	// filter log request.
	BlockLimit uint64 `hcl:"block_limit"`

	// ReplayAfter specifies after which time, in seconds, the event listener
	// should replay events. It is used to guarantee that events are eventually
	// delivered to subscribers even if they are not online at the time the event
	// was published.
	ReplayAfter []uint64 `hcl:"replay_after"`

	// ContractAddrs is a list of teleport contract addresses to listen
	// to.
	ContractAddrs []types.Address `hcl:"contract_addrs"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type teleportStarknetListener struct {
	// Sequencer is the name of the Starknet sequencer to use for listening
	// to events.
	Sequencer config.URL `hcl:"sequencer"`

	// Interval specifies how often, in seconds, the event listener should
	// check for new events.
	Interval uint32 `hcl:"interval"`

	// PrefetchPeriod specifies how far, in seconds, the event listener should
	// check for new  events during the initial synchronization.
	PrefetchPeriod uint32 `hcl:"prefetch_period"`

	// ReplayAfter specifies after which time, in seconds, the event listener
	// should replay events. It is used to guarantee that events are eventually
	// delivered to subscribers even if they are not online at the time the event
	// was published.
	ReplayAfter []uint32 `hcl:"replay_after"`

	// ContractAddrs is a list of teleport contract addresses to listen
	// to.
	ContractAddrs []*starknetClient.Felt `hcl:"contract_addrs"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

func (c *Config) EventPublisher(d Dependencies) (*publisher.EventPublisher, error) {
	if c.eventPublisher != nil {
		return c.eventPublisher, nil
	}
	var eventProviders []publisher.EventProvider
	if err := c.teleportEVM(&eventProviders, d); err != nil {
		return nil, err
	}
	if err := c.teleportStarknet(&eventProviders, d); err != nil {
		return nil, err
	}
	key, ok := d.Keys[c.EthereumKey]
	if !ok {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Ethereum key %q is not configured", c.EthereumKey),
			Subject:  c.Content.Attributes["ethereum_key"].Range.Ptr(),
		}
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
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Event Publisher service: %v", err),
			Subject:  c.Range.Ptr(),
		}
	}
	c.eventPublisher = eventPublisher
	return eventPublisher, nil
}

func (c *Config) teleportEVM(eps *[]publisher.EventProvider, d Dependencies) error {
	var err error
	for _, cfg := range c.TeleportEVM {
		if cfg.Interval == 0 {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Summary:  "Validation error",
				Detail:   "Interval cannot be zero",
				Severity: hcl.DiagError,
				Subject:  cfg.Content.Attributes["interval"].Range.Ptr(),
			}}
		}
		if len(cfg.ContractAddrs) == 0 {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Summary:  "Validation error",
				Detail:   "Contract addresses cannot be empty",
				Severity: hcl.DiagError,
				Subject:  cfg.Content.Attributes["contract_addrs"].Range.Ptr(),
			}}
		}
		client, ok := d.Clients[cfg.EthereumClient]
		if !ok {
			return &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Ethereum client %q is not configured", cfg.EthereumClient),
				Subject:  cfg.Content.Attributes["ethereum_client"].Range.Ptr(),
			}
		}
		replayAfter := make([]time.Duration, len(cfg.ReplayAfter))
		for i, r := range cfg.ReplayAfter {
			replayAfter[i] = time.Second * time.Duration(r)
		}
		var eventProvider publisher.EventProvider
		eventProvider, err = teleportevm.New(teleportevm.Config{
			Client:             geth.NewClient(client), //nolint:staticcheck // deprecated ethereum.Client
			Addresses:          cfg.ContractAddrs,
			Interval:           time.Second * time.Duration(cfg.Interval),
			PrefetchPeriod:     time.Second * time.Duration(cfg.PrefetchPeriod),
			BlockLimit:         cfg.BlockLimit,
			BlockConfirmations: cfg.BlockConfirmations,
			Logger:             d.Logger,
		})
		if err != nil {
			return &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create the EVM Event Provider for teleport: %v", err),
				Subject:  cfg.Range.Ptr(),
			}
		}
		if len(cfg.ReplayAfter) > 0 {
			eventProvider, err = replayer.New(replayer.Config{
				EventProvider: eventProvider,
				Interval:      time.Minute,
				ReplayAfter:   replayAfter,
			})
		}
		if err != nil {
			return &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create the EVM Event Provider for teleport: %v", err),
				Subject:  cfg.Range.Ptr(),
			}
		}
		*eps = append(*eps, eventProvider)
	}
	return nil
}

func (c *Config) teleportStarknet(eps *[]publisher.EventProvider, d Dependencies) error {
	var err error
	for _, cfg := range c.TeleportStarknet {
		if cfg.Interval == 0 {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Summary:  "Validation error",
				Detail:   "Interval cannot be zero",
				Severity: hcl.DiagError,
				Subject:  cfg.Content.Attributes["interval"].Range.Ptr(),
			}}
		}
		if len(cfg.ContractAddrs) == 0 {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Summary:  "Validation error",
				Detail:   "Contract addresses cannot be empty",
				Severity: hcl.DiagError,
				Subject:  cfg.Content.Attributes["contract_addrs"].Range.Ptr(),
			}}
		}
		replayAfter := make([]time.Duration, len(cfg.ReplayAfter))
		for i, r := range cfg.ReplayAfter {
			replayAfter[i] = time.Duration(r)
		}
		var eventProvider publisher.EventProvider
		eventProvider, err = teleportstarknet.New(teleportstarknet.Config{
			Sequencer:      starknetClient.NewSequencer(cfg.Sequencer.String(), http.Client{}),
			Addresses:      cfg.ContractAddrs,
			Interval:       time.Second * time.Duration(cfg.Interval),
			PrefetchPeriod: time.Second * time.Duration(cfg.PrefetchPeriod),
			Logger:         d.Logger,
		})
		if err != nil {
			return &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create the Starknet Event Provider for teleport: %v", err),
				Subject:  cfg.Range.Ptr(),
			}
		}
		if len(cfg.ReplayAfter) > 0 {
			eventProvider, err = replayer.New(replayer.Config{
				EventProvider: eventProvider,
				Interval:      time.Minute,
				ReplayAfter:   replayAfter,
			})
			if err != nil {
				return &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Runtime error",
					Detail:   fmt.Sprintf("Failed to create the Starknet Event Provider for teleport: %v", err),
					Subject:  cfg.Range.Ptr(),
				}
			}
		}
		*eps = append(*eps, eventProvider)
	}
	return nil
}
