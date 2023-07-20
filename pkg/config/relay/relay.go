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

package relay

import (
	"fmt"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median/eth"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/relayer"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

type Dependencies struct {
	Clients    ethereumConfig.ClientRegistry
	PriceStore *store.PriceStore
	Logger     log.Logger
}

type PriceStoreDependencies struct {
	Transport transport.Service
	Logger    log.Logger
}

type Config struct {
	// Interval is a time interval in seconds between checking if the price
	// needs to be updated.
	Interval uint32 `hcl:"interval"`

	// Median is a list of Median contracts to watch.
	Median []configMedian `hcl:"median,block"`

	// Pairs is a list of pairs to store in the price store.
	Feeds []types.Address `hcl:"feeds"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured services:
	relayer    *relayer.Relayer
	priceStore *store.PriceStore
}

type configMedian struct {
	// EthereumClient is a name of an Ethereum client to use.
	EthereumClient string `hcl:"ethereum_client"`

	// ContractAddr is an address of a Median contract.
	ContractAddr types.Address `hcl:"contract_addr"`

	// Pair is a pair name in the format "BASEQUOTE" (without slash).
	Pair string `hcl:"pair"`

	// Spread is a spread in percent points above which the price is considered
	// stale.
	Spread float64 `hcl:"spread"`

	// Expiration is a time in seconds after which the price is considered
	// stale.
	Expiration uint32 `hcl:"expiration"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

func (c *Config) Relay(d Dependencies) (*relayer.Relayer, error) {
	if c.relayer != nil {
		return c.relayer, nil
	}
	if c.Interval == 0 {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "Interval must be greater than 0",
			Subject:  c.Content.Attributes["interval"].Range.Ptr(),
		}}
	}
	cfg := relayer.Config{
		PokeTicker: timeutil.NewTicker(time.Second * time.Duration(c.Interval)),
		PriceStore: d.PriceStore,
		Logger:     d.Logger,
	}
	for _, pair := range c.Median {
		if pair.Expiration == 0 {
			return nil, hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   "Expiration must be greater than 0",
				Subject:  pair.Content.Attributes["expiration"].Range.Ptr(),
			}}
		}
		rpcClient := d.Clients[pair.EthereumClient]
		if rpcClient == nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Ethereum client %q is not configured", pair.EthereumClient),
				Subject:  pair.Content.Attributes["ethereum_client"].Range.Ptr(),
			}
		}
		cfg.Pairs = append(cfg.Pairs, &relayer.Pair{
			AssetPair:                 pair.Pair,
			Spread:                    pair.Spread,
			Expiration:                time.Second * time.Duration(pair.Expiration),
			Median:                    eth.NewMedian(rpcClient, pair.ContractAddr),
			FeedAddressesUpdateTicker: timeutil.NewTicker(time.Minute * 60),
		})
	}
	rel, err := relayer.New(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Relay error",
			Detail:   fmt.Sprintf("Failed to create the Relay service: %v", err),
			Subject:  &c.Range,
		}
	}
	c.relayer = rel
	return rel, nil
}

func (c *Config) PriceStore(d PriceStoreDependencies) (*store.PriceStore, error) {
	if c.priceStore != nil {
		return c.priceStore, nil
	}
	var pairs []string
	for _, pair := range c.Median {
		pairs = append(pairs, pair.Pair)
	}
	cfg := store.Config{
		Storage:   store.NewMemoryStorage(),
		Transport: d.Transport,
		Pairs:     pairs,
		Feeds:     c.Feeds,
		Logger:    d.Logger,
	}
	priceStore, err := store.New(cfg)
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
