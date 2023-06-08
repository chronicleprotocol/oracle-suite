package graph

import (
	"context"
	"errors"

	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type mockOrigin struct {
	fetchDataPoints func(_ context.Context, queries []any) (map[any]datapoint.Point, error)
}

func (f *mockOrigin) FetchDataPoints(ctx context.Context, types []any) (map[any]datapoint.Point, error) {
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

func (n numericValue) MarshalBinary() (data []byte, err error) {
	return nil, errors.New("not implemented")
}

func (n numericValue) UnmarshalBinary(_ []byte) error {
	return errors.New("not implemented")
}

type stringValue string

func (s stringValue) Print() string {
	return string(s)
}

func (s stringValue) MarshalBinary() (data []byte, err error) {
	return nil, errors.New("not implemented")
}

func (s stringValue) UnmarshalBinary(_ []byte) error {
	return errors.New("not implemented")
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

func (m *mockNode) DataPoint() datapoint.Point {
	args := m.Called()
	return args.Get(0).(datapoint.Point)
}

func (m *mockNode) Meta() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}
