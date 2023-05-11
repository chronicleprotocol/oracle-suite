package graph

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type mockOrigin struct {
	fetchDataPoints func(_ context.Context, queries []any) (map[any]data.Point, error)
}

func (f *mockOrigin) FetchDataPoints(ctx context.Context, types []any) (map[any]data.Point, error) {
	return f.fetchDataPoints(ctx, types)
}

type numericValue struct {
	x *bn.FloatNumber
}

func (n numericValue) Number() *bn.FloatNumber {
	return n.x
}

func (n numericValue) Print() string {
	return n.x.String()
}

type stringValue string

func (s stringValue) Print() string {
	return string(s)
}

type mockNode struct {
	mock.Mock
}

func (m *mockNode) AddNodes(nodes ...Node) error {
	args := m.Called(nodes)
	return args.Error(0)
}

func (m *mockNode) Nodes() []Node {
	args := m.Called()
	return args.Get(0).([]Node)
}

func (m *mockNode) DataPoint() data.Point {
	args := m.Called()
	return args.Get(0).(data.Point)
}

func (m *mockNode) Meta() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}
