package mocks

import (
	"context"
	"math/big"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
)

type Median struct {
	mock.Mock
}

func (m *Median) Address() types.Address {
	args := m.Called()
	return args.Get(0).(types.Address)
}

func (m *Median) Age(ctx context.Context) (time.Time, error) {
	args := m.Called(ctx)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *Median) Bar(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *Median) Val(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *Median) Wat(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.Get(0).(string), args.Error(1)
}

func (m *Median) Feeds(ctx context.Context) ([]types.Address, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.Address), args.Error(1)
}

func (m *Median) Poke(ctx context.Context, prices []*median.Price, simulateBeforeRun bool) (*types.Hash, error) {
	args := m.Called(ctx, prices, simulateBeforeRun)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (m *Median) Lift(ctx context.Context, addresses []types.Address, simulateBeforeRun bool) (*types.Hash, error) {
	args := m.Called(ctx, addresses, simulateBeforeRun)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (m *Median) Drop(ctx context.Context, addresses []types.Address, simulateBeforeRun bool) (*types.Hash, error) {
	args := m.Called(ctx, addresses, simulateBeforeRun)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (m *Median) SetBar(ctx context.Context, bar *big.Int, simulateBeforeRun bool) (*types.Hash, error) {
	args := m.Called(ctx, bar, simulateBeforeRun)
	return args.Get(0).(*types.Hash), args.Error(1)
}
