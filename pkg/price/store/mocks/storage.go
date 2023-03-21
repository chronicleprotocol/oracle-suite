package mocks

import (
	"context"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Storage struct {
	mock.Mock
}

func (s *Storage) Add(ctx context.Context, from types.Address, msg *messages.Price) error {
	args := s.Called(ctx, from, msg)
	return args.Error(0)
}

func (s *Storage) GetAll(ctx context.Context) (map[store.FeederPrice]*messages.Price, error) {
	args := s.Called(ctx)
	return args.Get(0).(map[store.FeederPrice]*messages.Price), args.Error(1)
}

func (s *Storage) GetByAssetPair(ctx context.Context, pair string) ([]*messages.Price, error) {
	args := s.Called(ctx, pair)
	return args.Get(0).([]*messages.Price), args.Error(1)
}

func (s *Storage) GetByFeeder(ctx context.Context, pair string, feeder types.Address) (*messages.Price, error) {
	args := s.Called(ctx, pair, feeder)
	return args.Get(0).(*messages.Price), args.Error(1)
}
