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

package mocks

import (
	"context"
	"sync"

	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/starknet"
)

type Sequencer struct {
	mu sync.Mutex
	mock.Mock
}

func (c *Sequencer) GetPendingBlock(ctx context.Context) (*starknet.Block, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	args := c.Called(ctx)
	return args.Get(0).(*starknet.Block), args.Error(1)
}

func (c *Sequencer) GetLatestBlock(ctx context.Context) (*starknet.Block, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	args := c.Called(ctx)
	return args.Get(0).(*starknet.Block), args.Error(1)
}

func (c *Sequencer) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*starknet.Block, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	args := c.Called(ctx, blockNumber)
	return args.Get(0).(*starknet.Block), args.Error(1)
}

func (c *Sequencer) Calls() []mock.Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Mock.Calls
}
