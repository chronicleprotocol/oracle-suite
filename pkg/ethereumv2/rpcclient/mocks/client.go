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

	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereumv2/types"
)

type Client struct {
	mock.Mock
}

func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	args := c.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (c *Client) BlockByNumber(ctx context.Context, number types.BlockNumber) (*types.BlockTxHashes, error) {
	args := c.Called(ctx, number)
	return args.Get(0).(*types.BlockTxHashes), args.Error(1)
}

func (c *Client) FullBlockByNumber(ctx context.Context, number types.BlockNumber) (*types.BlockTxObjects, error) {
	args := c.Called(ctx, number)
	return args.Get(0).(*types.BlockTxObjects), args.Error(1)
}

func (c *Client) GetTransactionCount(ctx context.Context, acc types.Address, block types.BlockNumber) (uint64, error) {
	args := c.Called(ctx, acc, block)
	return args.Get(0).(uint64), args.Error(1)
}

func (c *Client) SendRawTransaction(ctx context.Context, data types.Bytes) (*types.Hash, error) {
	args := c.Called(ctx, data)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (c *Client) GetStorageAt(ctx context.Context, acc types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error) {
	args := c.Called(ctx, acc, key, block)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (c *Client) FilterLogs(ctx context.Context, q types.FilterLogsQuery) ([]types.Log, error) {
	args := c.Called(ctx, q)
	return args.Get(0).([]types.Log), args.Error(1)
}
