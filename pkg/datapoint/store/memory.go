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

package store

import (
	"context"
	"sync"

	"github.com/defiweb/go-eth/types"
)

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu sync.RWMutex
	ds map[dataPointKey]StoredDataPoint
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		ds: make(map[dataPointKey]StoredDataPoint),
	}
}

// Add implements the Storage interface.
func (m *MemoryStorage) Add(_ context.Context, point StoredDataPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	prev, ok := m.ds[dataPointKey{feed: point.From, model: point.Model}]
	if ok && prev.DataPoint.Time.After(point.DataPoint.Time) {
		return nil // ignore older points
	}
	m.ds[dataPointKey{feed: point.From, model: point.Model}] = point
	return nil
}

// LatestFrom implements the Storage interface.
func (m *MemoryStorage) LatestFrom(_ context.Context, from types.Address, model string) (StoredDataPoint, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.ds[dataPointKey{feed: from, model: model}]
	return p, ok, nil
}

// Latest implements the Storage interface.
func (m *MemoryStorage) Latest(_ context.Context, model string) (map[types.Address]StoredDataPoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ps := make(map[types.Address]StoredDataPoint)
	for k, v := range m.ds {
		if k.model == model {
			ps[k.feed] = v
		}
	}
	return ps, nil
}

type dataPointKey struct {
	feed  types.Address
	model string
}
