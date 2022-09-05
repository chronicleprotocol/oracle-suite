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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ghost"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var ghostFactory = func(cfg ghost.Config) (*ghost.Ghost, error) {
	return ghost.New(cfg)
}

type Ghost struct {
	Interval int      `yaml:"interval"`
	Pairs    []string `yaml:"pairs"`
}

type Dependencies struct {
	Gofer     provider.Provider
	Signer    ethereum.Signer
	Transport transport.Transport
	Logger    log.Logger
}

func (c *Ghost) Configure(d Dependencies) (*ghost.Ghost, error) {
	cfg := ghost.Config{
		PriceProvider: d.Gofer,
		Signer:        d.Signer,
		Transport:     d.Transport,
		Logger:        d.Logger,
		Interval:      time.Second * time.Duration(c.Interval),
		Pairs:         c.Pairs,
	}
	return ghostFactory(cfg)
}
