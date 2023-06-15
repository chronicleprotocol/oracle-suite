package store

import (
	"context"
	"sync"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
)

// MemoryStorage is an in-memory implementation of Storage.
type MemoryStorage struct {
	mu sync.RWMutex
	ds map[dataPointKey]datapoint.Point
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		ds: make(map[dataPointKey]datapoint.Point),
	}
}

// Add implements the Storage interface.
func (m *MemoryStorage) Add(_ context.Context, from types.Address, model string, point datapoint.Point) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	prev, ok := m.ds[dataPointKey{feeder: from, model: model}]
	if ok && prev.Time.After(point.Time) {
		return nil // ignore older points
	}
	m.ds[dataPointKey{feeder: from, model: model}] = point
	return nil
}

// LatestFrom implements the Storage interface.
func (m *MemoryStorage) LatestFrom(_ context.Context, from types.Address, model string) (datapoint.Point, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.ds[dataPointKey{feeder: from, model: model}]
	return p, ok, nil
}

// Latest implements the Storage interface.
func (m *MemoryStorage) Latest(_ context.Context, model string) (map[types.Address]datapoint.Point, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ps := make(map[types.Address]datapoint.Point)
	for k, v := range m.ds {
		if k.model == model {
			ps[k.feeder] = v
		}
	}
	return ps, nil
}

type dataPointKey struct {
	feeder types.Address
	model  string
}
