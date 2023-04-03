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

package feed

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type Config struct {
	// EthereumKey is the name of the Ethereum key to use for signing prices.
	EthereumKey string `hcl:"ethereum_key"`

	// Interval is the interval at which to publish prices in seconds.
	Interval uint32 `hcl:"interval"`

	// Pairs is the list of pairs to publish prices for.
	Pairs []provider.Pair `hcl:"pairs"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured service:
	feeder *feeder.Feeder
}

type Dependencies struct {
	Keys      ethereumConfig.KeyRegistry
	Gofer     provider.Provider
	Transport transport.Transport
	Logger    log.Logger
}

func (c *Config) Feed(d Dependencies) (*feeder.Feeder, error) {
	if c.feeder != nil {
		return c.feeder, nil
	}
	if c.Interval == 0 {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Summary:  "Validation error",
			Detail:   "Interval cannot be zero",
			Severity: hcl.DiagError,
			Subject:  c.Content.Attributes["interval"].Range.Ptr(),
		}}
	}
	ethereumKey, ok := d.Keys[c.EthereumKey]
	if !ok {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Ethereum key %q is not configured", c.EthereumKey),
			Subject:  c.Content.Attributes["ethereum_key"].Range.Ptr(),
		}
	}
	pairs := make([]string, len(c.Pairs))
	for i, p := range c.Pairs {
		pairs[i] = p.String()
	}
	cfg := feeder.Config{
		PriceProvider: d.Gofer,
		Signer:        ethereumKey,
		Transport:     d.Transport,
		Logger:        d.Logger,
		Interval:      timeutil.NewTicker(time.Second * time.Duration(c.Interval)),
		Pairs:         pairs,
	}
	feed, err := feeder.New(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Feed service: %v", err),
			Subject:  c.Range.Ptr(),
		}
	}
	c.feeder = feed
	return feed, nil
}
