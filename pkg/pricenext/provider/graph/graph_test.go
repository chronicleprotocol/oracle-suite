package graph

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

type mockOrigin struct {
	fetchTicks func(_ context.Context, pairs []provider.Pair) []provider.Tick
}

func (f *mockOrigin) FetchTicks(ctx context.Context, pairs []provider.Pair) []provider.Tick {
	return f.fetchTicks(ctx, pairs)
}

type mockNode struct {
	mock.Mock
}

func (m *mockNode) AddBranch(branch ...Node) error {
	args := m.Called(branch)
	return args.Error(0)
}

func (m *mockNode) Branches() []Node {
	args := m.Called()
	return args.Get(0).([]Node)
}

func (m *mockNode) Pair() provider.Pair {
	args := m.Called()
	return args.Get(0).(provider.Pair)
}

func (m *mockNode) Tick() provider.Tick {
	args := m.Called()
	return args.Get(0).(provider.Tick)
}

func (m *mockNode) Meta() provider.Meta {
	args := m.Called()
	return args.Get(0).(provider.Meta)
}
