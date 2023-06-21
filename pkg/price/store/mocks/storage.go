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

func (s *Storage) GetByFeed(ctx context.Context, pair string, feeder types.Address) (*messages.Price, error) {
	args := s.Called(ctx, pair, feeder)
	return args.Get(0).(*messages.Price), args.Error(1)
}
