package store

import (
	"context"
	"sync"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
)

type MemoryStorage struct {
	mu sync.RWMutex
	ds map[feederDataPoint]datapoint.Point
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		ds: make(map[feederDataPoint]datapoint.Point),
	}
}

func (m *MemoryStorage) Add(_ context.Context, from types.Address, model string, point datapoint.Point) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ds[feederDataPoint{feeder: from, model: model}] = point
	return nil
}

func (m *MemoryStorage) LatestFrom(_ context.Context, from types.Address, model string) (datapoint.Point, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.ds[feederDataPoint{feeder: from, model: model}]
	return p, ok, nil
}

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

type feederDataPoint struct {
	feeder types.Address
	model  string
}
