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
	"errors"
	"testing"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	rpc.RPC

	callFn            func(ctx context.Context, call types.Call, block types.BlockNumber) ([]byte, *types.Call, error)
	sendTransactionFn func(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error)
}

func (m *mockClient) Call(ctx context.Context, call types.Call, block types.BlockNumber) ([]byte, *types.Call, error) {
	return m.callFn(ctx, call, block)
}

func (m *mockClient) SendTransaction(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error) {
	return m.sendTransactionFn(ctx, tx)
}

func TestCall_Address(t *testing.T) {
	c := NewCall(CallOpts{
		Address: types.MustAddressFromHex("0x1122334455667788990011223344556677889900"),
	})
	assert.Equal(t, "0x1122334455667788990011223344556677889900", c.Address().String())
}

func TestCall_Client(t *testing.T) {
	m := &mockClient{}
	c := NewCall(CallOpts{
		Client: m,
	})
	assert.Equal(t, m, c.Client())
}

func TestCall_CallData(t *testing.T) {
	c := NewCall(CallOpts{
		Encoder: func() ([]byte, error) {
			return []byte{1, 2, 3}, nil
		},
	})
	callData, err := c.CallData()
	assert.Equal(t, []byte{1, 2, 3}, callData)
	assert.NoError(t, err)
}

func TestCall_CallData_Error(t *testing.T) {
	c := NewCall(CallOpts{
		Encoder: func() ([]byte, error) {
			return nil, assert.AnError
		},
	})
	_, err := c.CallData()
	assert.Error(t, err)
}

func TestCall_DecodeTo(t *testing.T) {
	c := NewCall(CallOpts{
		Decoder: func(bytes []byte, res any) error {
			*res.(*[]byte) = bytes
			return nil
		},
	})
	var v []byte
	err := c.DecodeTo([]byte{1, 2, 3}, &v)
	assert.Equal(t, []byte{1, 2, 3}, v)
	assert.NoError(t, err)
}

func TestCall_Call(t *testing.T) {
	t.Run("valid call", func(t *testing.T) {
		m := &mockClient{}
		m.callFn = func(ctx context.Context, call types.Call, block types.BlockNumber) ([]byte, *types.Call, error) {
			assert.Equal(t, types.LatestBlockNumber, block)
			assert.Equal(t, "0x1122334455667788990011223344556677889900", call.To.String())
			assert.Equal(t, []byte{1, 2, 3}, call.Input)
			return []byte{4, 5, 6}, &types.Call{}, nil
		}
		c := NewCall(CallOpts{
			Client:  m,
			Address: types.MustAddressFromHex("0x1122334455667788990011223344556677889900"),
			Encoder: func() ([]byte, error) {
				return []byte{1, 2, 3}, nil
			},
			Decoder: func(bytes []byte, res any) error {
				*res.(*[]byte) = bytes
				return nil
			},
		})
		var v []byte
		err := c.Call(context.Background(), types.LatestBlockNumber, &v)
		assert.Equal(t, []byte{4, 5, 6}, v)
		assert.NoError(t, err)
	})
	t.Run("call error", func(t *testing.T) {
		m := &mockClient{}
		m.callFn = func(ctx context.Context, call types.Call, block types.BlockNumber) ([]byte, *types.Call, error) {
			return nil, nil, assert.AnError
		}
		c := NewCall(CallOpts{
			Client:  m,
			Address: types.MustAddressFromHex("0x1122334455667788990011223344556677889900"),
			Encoder: func() ([]byte, error) {
				return []byte{1, 2, 3}, nil
			},
			Decoder: func(bytes []byte, res any) error {
				*res.(*[]byte) = bytes
				return nil
			},
		})
		var v []byte
		err := c.Call(context.Background(), types.LatestBlockNumber, &v)
		assert.Error(t, err)
	})
	t.Run("call error with error decoder", func(t *testing.T) {
		m := &mockClient{}
		m.callFn = func(ctx context.Context, call types.Call, block types.BlockNumber) ([]byte, *types.Call, error) {
			return nil, nil, errors.New("some error")
		}
		c := NewCall(CallOpts{
			Client:  m,
			Address: types.MustAddressFromHex("0x1122334455667788990011223344556677889900"),
			Encoder: func() ([]byte, error) {
				return []byte{1, 2, 3}, nil
			},
			ErrorDecoder: func(err error) error {
				return assert.AnError
			},
		})
		var v []byte
		err := c.Call(context.Background(), types.LatestBlockNumber, &v)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestTransactableCall_SendTransaction(t *testing.T) {
	t.Run("successful call", func(t *testing.T) {
		m := &mockClient{}
		m.sendTransactionFn = func(ctx context.Context, tx types.Transaction) (*types.Hash, *types.Transaction, error) {
			assert.Equal(t, "0x1122334455667788990011223344556677889900", tx.Call.To.String())
			assert.Equal(t, []byte{1, 2, 3}, tx.Call.Input)
			return &types.Hash{}, &types.Transaction{}, nil
		}
		c := NewTransactableCall(CallOpts{
			Client:  m,
			Address: types.MustAddressFromHex("0x1122334455667788990011223344556677889900"),
			Encoder: func() ([]byte, error) {
				return []byte{1, 2, 3}, nil
			},
		})
		_, _, err := c.SendTransaction(context.Background())
		require.NoError(t, err)
	})
}
