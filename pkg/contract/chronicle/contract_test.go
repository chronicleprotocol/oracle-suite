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

package chronicle

import (
	"context"
	"math/big"
	"testing"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRPC struct {
	rpc.Client
	mock.Mock
}

func (m *mockRPC) BlockNumber(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *mockRPC) GetStorageAt(ctx context.Context, account types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error) {
	args := m.Called(ctx, account, key, block)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (m *mockRPC) Call(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
	args := m.Called(ctx, call, blockNumber)
	return args.Get(0).([]byte), args.Get(1).(*types.Call), args.Error(2)
}

func (m *mockRPC) SendTransaction(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(*types.Hash), args.Get(1).(*types.Transaction), args.Error(2)
}

func TestBytesToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "byte slice with null byte",
			input:    []byte("hello\x00world"),
			expected: "hello",
		},
		{
			name:     "byte slice without null byte",
			input:    []byte("hello"),
			expected: "hello",
		},
		{
			name:     "empty byte slice",
			input:    []byte(""),
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, bytes32ToString(tt.input))
		})
	}
}

func toEIP191(msg []byte) types.Hash {
	return crypto.Keccak256(crypto.AddMessagePrefix(msg))
}
