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

package contract

import (
	"context"
	"testing"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRPC struct {
	rpc.Client
	mock.Mock
}

func (m *mockRPC) GetStorageAt(ctx context.Context, account types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error) {
	args := m.Called(ctx, account, key, block)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (m *mockRPC) Call(ctx context.Context, call types.Call, blockNumber types.BlockNumber) ([]byte, error) {
	args := m.Called(ctx, call, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockRPC) SendTransaction(ctx context.Context, tx types.Transaction) (*types.Hash, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func TestSimulateTransaction(t *testing.T) {
	ctx := context.Background()

	// Mocked transaction for the test
	tx := types.Transaction{
		Call: types.Call{
			To:    types.MustAddressFromHexPtr("0x1122344556677889900112233445566778899002"),
			Input: []byte("mockedInput"),
		},
	}

	// Fake data for revert and panic
	revertData := hexutil.MustHexToBytes(
		"0x" +
			"08c379a0" +
			"0000000000000000000000000000000000000000000000000000000000000020" +
			"0000000000000000000000000000000000000000000000000000000000000005" +
			"7265766572740000000000000000000000000000000000000000000000000000",
	)
	panicData := hexutil.MustHexToBytes(
		"0x" +
			"4e487b71" +
			"7265766572740000000000000000000000000000000000000000000000000000",
	)

	t.Run("successful transaction", func(t *testing.T) {
		mockClient := new(mockRPC)
		mockClient.On(
			"Call",
			ctx,
			tx.Call,
			types.LatestBlockNumber,
		).Return(
			[]byte{},
			nil,
		)

		err := simulateTransaction(ctx, mockClient, tx)
		require.NoError(t, err)
	})

	t.Run("reverted transaction", func(t *testing.T) {
		mockClient := new(mockRPC)
		mockClient.On(
			"Call",
			ctx,
			tx.Call,
			types.LatestBlockNumber,
		).Return(
			revertData,
			nil,
		)

		err := simulateTransaction(ctx, mockClient, tx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transaction reverted")
	})

	t.Run("panicked transaction", func(t *testing.T) {
		mockClient := new(mockRPC)
		mockClient.On(
			"Call",
			ctx,
			tx.Call,
			types.LatestBlockNumber,
		).Return(
			panicData,
			nil,
		)

		err := simulateTransaction(ctx, mockClient, tx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transaction panicked")
	})
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
			assert.Equal(t, tt.expected, bytesToString(tt.input))
		})
	}
}
