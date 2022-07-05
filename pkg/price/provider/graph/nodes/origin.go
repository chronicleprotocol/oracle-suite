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

package nodes

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

type ErrIncompatiblePair struct {
	Given    provider.Pair
	Expected provider.Pair
}

func (e ErrIncompatiblePair) Error() string {
	return fmt.Sprintf(
		"a price with different pair ignested to the OriginNode, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

type IncompatibleOriginErr struct {
	Given    string
	Expected string
}

func (e IncompatibleOriginErr) Error() string {
	return fmt.Sprintf(
		"a price from different origin ignested to the OriginNode, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

type ErrPriceTTLExpired struct {
	Price OriginPrice
	TTL   time.Duration
}

func (e ErrPriceTTLExpired) Error() string {
	return fmt.Sprintf(
		"the price TTL for the pair %s expired",
		e.Price.Pair,
	)
}

// OriginNode contains a Price fetched directly from an origin.
type OriginNode struct {
	mu sync.RWMutex

	originPair OriginPair
	price      OriginPrice
	minTTL     time.Duration
	maxTTL     time.Duration
}

func NewOriginNode(originPair OriginPair, minTTL time.Duration, maxTTL time.Duration) *OriginNode {
	return &OriginNode{
		originPair: originPair,
		minTTL:     minTTL,
		maxTTL:     maxTTL,
	}
}

// OriginPair implements the Feedable interface.
func (n *OriginNode) OriginPair() OriginPair {
	return n.originPair
}

// Ingest implements Feedable interface.
func (n *OriginNode) Ingest(price OriginPrice) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var err error
	if !price.Pair.Equal(n.originPair.Pair) {
		err = multierror.Append(err, ErrIncompatiblePair{
			Given:    price.Pair,
			Expected: n.originPair.Pair,
		})
	}

	if price.Origin != n.originPair.Origin {
		err = multierror.Append(err, IncompatibleOriginErr{
			Given:    price.Origin,
			Expected: n.originPair.Origin,
		})
	}

	if err == nil {
		n.price = price
	}

	return err
}

// MinTTL implements the Feedable interface.
func (n *OriginNode) MinTTL() time.Duration {
	return n.minTTL
}

// MaxTTL implements the Feedable interface.
func (n *OriginNode) MaxTTL() time.Duration {
	return n.maxTTL
}

// Expired implements the Feedable interface.
func (n *OriginNode) Expired() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.expired()
}

// Price implements the Feedable interface.
func (n *OriginNode) Price() OriginPrice {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.price.Error == nil {
		if n.expired() {
			n.price.Error = ErrPriceTTLExpired{
				Price: n.price,
				TTL:   n.maxTTL,
			}
		}
	}

	return n.price
}

// Children implements the Node interface.
func (n *OriginNode) Children() []Node {
	return []Node{}
}

func (n *OriginNode) expired() bool {
	return n.price.Time.Before(time.Now().Add(-1 * n.MaxTTL()))
}
