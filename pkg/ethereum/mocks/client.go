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

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type Client struct {
	mock.Mock
}

func (c *Client) Call(ctx context.Context, call ethereum.Call) ([]byte, error) {
	args := c.Called(ctx, call)
	return args.Get(0).([]byte), args.Error(1)
}

func (c *Client) MultiCall(ctx context.Context, calls []ethereum.Call) ([][]byte, error) {
	args := c.Called(ctx, calls)
	return args.Get(0).([][]byte), args.Error(1)
}

func (c *Client) Storage(ctx context.Context, address ethereum.Address, key ethereum.Hash) ([]byte, error) {
	args := c.Called(ctx, address, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (c *Client) SendTransaction(ctx context.Context, transaction *ethereum.Transaction) (*ethereum.Hash, error) {
	args := c.Called(ctx, transaction)
	return args.Get(0).(*ethereum.Hash), args.Error(1)
}
