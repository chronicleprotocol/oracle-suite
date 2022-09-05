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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	oracleGeth "github.com/chronicleprotocol/oracle-suite/pkg/price/oracle/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/spectre"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var spectreFactory = func(cfg spectre.Config) (*spectre.Spectre, error) {
	return spectre.NewSpectre(cfg)
}

//nolint
var priceStoreFactory = func(cfg store.Config) (*store.PriceStore, error) {
	return store.New(cfg)
}

type Spectre struct {
	Interval    int64                 `yaml:"interval"`
	Medianizers map[string]Medianizer `yaml:"medianizers"`
}

type Medianizer struct {
	Contract         string  `yaml:"oracle"`
	OracleSpread     float64 `yaml:"oracleSpread"`
	OracleExpiration int64   `yaml:"oracleExpiration"`
	MsgExpiration    int64   `yaml:"msgExpiration"`
}

type Dependencies struct {
	Signer         ethereum.Signer
	PriceStore     *store.PriceStore
	EthereumClient ethereum.Client
	Feeds          []ethereum.Address
	Logger         log.Logger
}

type PriceStoreDependencies struct {
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *Spectre) ConfigureSpectre(d Dependencies) (*spectre.Spectre, error) {
	cfg := spectre.Config{
		Signer:     d.Signer,
		Interval:   time.Second * time.Duration(c.Interval),
		PriceStore: d.PriceStore,
		Logger:     d.Logger,
	}
	for name, pair := range c.Medianizers {
		cfg.Pairs = append(cfg.Pairs, &spectre.Pair{
			AssetPair:        name,
			OracleSpread:     pair.OracleSpread,
			OracleExpiration: time.Second * time.Duration(pair.OracleExpiration),
			PriceExpiration:  time.Second * time.Duration(pair.MsgExpiration),
			Median:           oracleGeth.NewMedian(d.EthereumClient, ethereum.HexToAddress(pair.Contract)),
		})
	}
	return spectreFactory(cfg)
}

func (c *Spectre) ConfigurePriceStore(d PriceStoreDependencies) (*store.PriceStore, error) {
	cfg := store.Config{
		Storage:   store.NewMemoryStorage(),
		Signer:    d.Signer,
		Transport: d.Transport,
		Pairs:     maputil.Keys(c.Medianizers),
		Logger:    d.Logger,
	}

	return priceStoreFactory(cfg)
}
