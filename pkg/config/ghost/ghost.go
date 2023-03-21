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

package ghost

import (
	"fmt"
	"time"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type ConfigGhost struct {
	// EthereumKey is the name of the Ethereum key to use for signing prices.
	EthereumKey string `hcl:"ethereum_key"`

	// Interval is the interval at which to publish prices.
	Interval int `hcl:"interval"`

	// Pairs is the list of pairs to publish prices for.
	Pairs []string `hcl:"pairs"`
}

type Dependencies struct {
	Keys      ethereumConfig.KeyRegistry
	Gofer     provider.Provider
	Transport transport.Transport
	Logger    log.Logger
}

func (c *ConfigGhost) Ghost(d Dependencies) (*feeder.Feeder, error) {
	ethereumKey, ok := d.Keys[c.EthereumKey]
	if !ok {
		return nil, fmt.Errorf("ghost: ethereum key %s not found", c.EthereumKey)
	}
	cfg := feeder.Config{
		PriceProvider: d.Gofer,
		Signer:        ethereumKey,
		Transport:     d.Transport,
		Logger:        d.Logger,
		Interval:      timeutil.NewTicker(time.Second * time.Duration(c.Interval)),
		Pairs:         c.Pairs,
	}
	return feeder.New(cfg)
}
