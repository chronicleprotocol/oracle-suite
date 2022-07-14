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

package store

import (
	"context"
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type MemoryStorage struct {
	mu sync.RWMutex
	ps map[FeederPrice]*messages.Price
}

// NewMemoryStorage creates a new store instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{ps: make(map[FeederPrice]*messages.Price)}
}

// Add implements the store.Storage interface.
func (p *MemoryStorage) Add(_ context.Context, from ethereum.Address, price *messages.Price) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	fp := FeederPrice{AssetPair: price.Price.Wat, Feeder: from}
	if prev, ok := p.ps[fp]; ok && prev.Price.Age.After(price.Price.Age) {
		return nil
	}
	p.ps[fp] = price
	return nil
}

// GetAll implements the store.Storage interface.
func (p *MemoryStorage) GetAll(_ context.Context) (map[FeederPrice]*messages.Price, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	r := map[FeederPrice]*messages.Price{}
	for k, v := range p.ps {
		r[k] = v
	}
	return r, nil
}

// GetByAssetPair implements the store.Storage interface.
func (p *MemoryStorage) GetByAssetPair(_ context.Context, pair string) ([]*messages.Price, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var ps []*messages.Price
	for k, v := range p.ps {
		if k.AssetPair != pair {
			continue
		}
		ps = append(ps, v)
	}
	return ps, nil
}

// GetByFeeder implements the store.Storage interface.
func (p *MemoryStorage) GetByFeeder(_ context.Context, pair string, feeder ethereum.Address) (*messages.Price, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	fp := FeederPrice{
		AssetPair: pair,
		Feeder:    feeder,
	}
	if m, ok := p.ps[fp]; ok {
		return m, nil
	}
	return nil, nil
}

var _ Storage = (*MemoryStorage)(nil)
