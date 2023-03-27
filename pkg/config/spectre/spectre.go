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

package spectre

import (
	"fmt"
	"time"

	"github.com/defiweb/go-eth/types"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	medianGeth "github.com/chronicleprotocol/oracle-suite/pkg/price/median/geth"
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
	Transport transport.Transport
	Logger    log.Logger
}

type ConfigSpectre struct {
	// Interval is a time interval in seconds between checking if the price
	// needs to be updated.
	Interval int64 `hcl:"interval"`

	// Median is a list of Median contracts to watch.
	Median []configMedian `hcl:"median,block"`

	// Configured services:
	relayer    *relayer.Relayer
	priceStore *store.PriceStore
}

type configMedian struct {
	// EthereumClient is a name of an Ethereum client to use.
	EthereumClient string `hcl:"ethereum_client"`

	// ContractAddr is an address of a Median contract.
	ContractAddr string `hcl:"contract_addr"`

	// Pair is a pair name in the format "base/quote".
	Pair string `hcl:"pair"`

	// Spread is a spread in percent points above which the price is considered
	// stale.
	Spread float64 `hcl:"spread"`

	// Expiration is a time in seconds after which the price is considered
	// stale.
	Expiration int64 `hcl:"expiration"`
}

func (c *ConfigSpectre) Relayer(d Dependencies) (*relayer.Relayer, error) {
	if c.relayer != nil {
		return c.relayer, nil
	}
	cfg := relayer.Config{
		PokeTicker: timeutil.NewTicker(time.Second * time.Duration(c.Interval)),
		PriceStore: d.PriceStore,
		Logger:     d.Logger,
	}
	for _, pair := range c.Median {
		contractAddr, err := types.AddressFromHex(pair.ContractAddr)
		if err != nil {
			return nil, fmt.Errorf("spectre config: invalid contract address %s", pair.ContractAddr)
		}
		rpcClient := d.Clients[pair.EthereumClient]
		if rpcClient == nil {
			return nil, fmt.Errorf("spectre config: ethereum client %s not found", pair.EthereumClient)
		}
		ethClient := geth.NewClient(rpcClient) //nolint:staticcheck // deprecated ethereum.Client
		cfg.Pairs = append(cfg.Pairs, &relayer.Pair{
			AssetPair:                   pair.Pair,
			Spread:                      pair.Spread,
			Expiration:                  time.Second * time.Duration(pair.Expiration),
			Median:                      medianGeth.NewMedian(ethClient, contractAddr),
			FeederAddressesUpdateTicker: timeutil.NewTicker(time.Minute * 60),
		})
	}
	rel, err := relayer.New(cfg)
	if err != nil {
		return nil, err
	}
	c.relayer = rel
	return rel, nil
}

func (c *ConfigSpectre) PriceStore(d PriceStoreDependencies) (*store.PriceStore, error) {
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
		Logger:    d.Logger,
	}
	priceStore, err := store.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("spectre config: failed to create price store: %w", err)
	}
	c.priceStore = priceStore
	return priceStore, nil
}
