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

	"github.com/chronicleprotocol/oracle-suite/pkg/datastore"
	datastoreMemory "github.com/chronicleprotocol/oracle-suite/pkg/datastore/memory"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	oracleGeth "github.com/chronicleprotocol/oracle-suite/pkg/oracle/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/spectre"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var spectreFactory = func(cfg spectre.Config) (*spectre.Spectre, error) {
	return spectre.NewSpectre(cfg)
}

var datastoreFactory = func(cfg datastoreMemory.Config) (datastore.Datastore, error) {
	return datastoreMemory.NewDatastore(cfg)
}

type Spectre struct {
	Interval    int64                 `json:"interval"`
	Medianizers map[string]Medianizer `json:"medianizers"`
}

type Medianizer struct {
	Contract         string  `json:"oracle"`
	OracleSpread     float64 `json:"oracleSpread"`
	OracleExpiration int64   `json:"oracleExpiration"`
	MsgExpiration    int64   `json:"msgExpiration"`
}

type Dependencies struct {
	Signer         ethereum.Signer
	Datastore      datastore.Datastore
	EthereumClient ethereum.Client
	Feeds          []ethereum.Address
	Logger         log.Logger
}

type DatastoreDependencies struct {
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *Spectre) ConfigureSpectre(d Dependencies) (*spectre.Spectre, error) {
	cfg := spectre.Config{
		Signer:    d.Signer,
		Interval:  time.Second * time.Duration(c.Interval),
		Datastore: d.Datastore,
		Logger:    d.Logger,
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

func (c *Spectre) ConfigureDatastore(d DatastoreDependencies) (datastore.Datastore, error) {
	cfg := datastoreMemory.Config{
		Signer:    d.Signer,
		Transport: d.Transport,
		Pairs:     make(map[string]*datastoreMemory.Pair),
		Logger:    d.Logger,
	}
	for name := range c.Medianizers {
		cfg.Pairs[name] = &datastoreMemory.Pair{Feeds: d.Feeds}
	}
	return datastoreFactory(cfg)
}
