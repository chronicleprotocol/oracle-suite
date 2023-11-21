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
)

type mockRPC struct {
	rpc.Client

	blockNumberFn     func(ctx context.Context) (*big.Int, error)
	getStorageAtFn    func(ctx context.Context, account types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error)
	callFn            func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error)
	sendTransactionFn func(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error)
}

func newMockRPC(t *testing.T) *mockRPC {
	m := &mockRPC{}
	m.reset(t)
	return m
}

func (m *mockRPC) reset(t *testing.T) {
	m.blockNumberFn = func(ctx context.Context) (*big.Int, error) {
		assert.FailNow(t, "unexpected call to BlockNumber")
		return nil, nil
	}
	m.getStorageAtFn = func(ctx context.Context, account types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error) {
		assert.FailNow(t, "unexpected call to GetStorageAt")
		return nil, nil
	}
	m.callFn = func(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
		assert.FailNow(t, "unexpected call to Call")
		return nil, nil, nil
	}
	m.sendTransactionFn = func(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error) {
		assert.FailNow(t, "unexpected call to SendTransaction")
		return nil, nil, nil
	}
}

func (m *mockRPC) BlockNumber(ctx context.Context) (*big.Int, error) {
	return m.blockNumberFn(ctx)
}

func (m *mockRPC) GetStorageAt(ctx context.Context, account types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error) {
	return m.getStorageAtFn(ctx, account, key, block)
}

func (m *mockRPC) Call(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, *types.Call, error) {
	return m.callFn(ctx, call, blockNumber)
}

func (m *mockRPC) SendTransaction(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error) {
	return m.sendTransactionFn(ctx, tx)
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
