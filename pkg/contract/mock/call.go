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

package mock

import (
	"context"
	"reflect"
	"testing"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
)

func NewCaller(t *testing.T) *Caller {
	mc := &Caller{}
	mc.MockReset(t)
	return mc
}

func NewTypedCaller[T any](t *testing.T) *TypedCaller[T] {
	mc := &TypedCaller[T]{}
	mc.MockReset(t)
	return mc
}

type Caller struct {
	AddressFn         func() types.Address
	CallDataFn        func() ([]byte, error)
	ClientFn          func() rpc.RPC
	CallFn            func(ctx context.Context, number types.BlockNumber, res any) error
	GasFn             func(ctx context.Context, number types.BlockNumber) (uint64, error)
	SendTransactionFn func(ctx context.Context) (*types.Hash, *types.Transaction, error)
	DecodeToFn        func(bytes []byte, a any) error
}

func (m *Caller) MockAllowAllCalls() *Caller {
	m.AddressFn = func() types.Address { return types.Address{} }
	m.CallDataFn = func() ([]byte, error) { return nil, nil }
	m.ClientFn = func() rpc.RPC { return nil }
	m.CallFn = func(ctx context.Context, number types.BlockNumber, res any) error { return nil }
	m.GasFn = func(ctx context.Context, number types.BlockNumber) (uint64, error) { return 0, nil }
	m.SendTransactionFn = func(ctx context.Context) (*types.Hash, *types.Transaction, error) { return nil, nil, nil }
	m.DecodeToFn = func(bytes []byte, a any) error { return nil }
	return m
}

func (m *Caller) MockResult(result any, err error) *Caller {
	m.AddressFn = func() types.Address {
		return types.Address{}
	}
	m.CallFn = func(ctx context.Context, number types.BlockNumber, res any) error {
		tv := reflect.ValueOf(res)
		rv := reflect.ValueOf(result)
		for tv.Kind() == reflect.Ptr || tv.Kind() == reflect.Interface {
			if tv.Kind() == rv.Kind() {
				break
			}
			tv = tv.Elem()
		}
		tv.Set(rv)
		return err
	}
	m.GasFn = func(ctx context.Context, number types.BlockNumber) (uint64, error) {
		return 0, nil
	}
	m.DecodeToFn = func(bytes []byte, target any) error {
		tv := reflect.ValueOf(target)
		rv := reflect.ValueOf(result)
		for tv.Kind() == reflect.Ptr || tv.Kind() == reflect.Interface {
			if tv.Kind() == rv.Kind() {
				break
			}
			tv = tv.Elem()
		}
		tv.Set(rv)
		return err
	}
	return m
}

func (m *Caller) MockReset(t *testing.T) *Caller {
	m.AddressFn = func() types.Address {
		assert.FailNow(t, "unexpected call to Address")
		return types.Address{}
	}
	m.CallDataFn = func() ([]byte, error) {
		assert.FailNow(t, "unexpected call to CallData")
		return nil, nil
	}
	m.ClientFn = func() rpc.RPC {
		assert.FailNow(t, "unexpected call to Client")
		return nil
	}
	m.CallFn = func(ctx context.Context, number types.BlockNumber, res any) error {
		assert.FailNow(t, "unexpected call to Call")
		return nil
	}
	m.GasFn = func(ctx context.Context, number types.BlockNumber) (uint64, error) {
		assert.FailNow(t, "unexpected call to Gas")
		return 0, nil
	}
	m.SendTransactionFn = func(ctx context.Context) (*types.Hash, *types.Transaction, error) {
		assert.FailNow(t, "unexpected call to SendTransaction")
		return nil, nil, nil
	}
	m.DecodeToFn = func(bytes []byte, a any) error {
		assert.FailNow(t, "unexpected call to DecodeTo")
		return nil
	}
	return m
}

func (m *Caller) Address() types.Address {
	return m.AddressFn()
}

func (m *Caller) CallData() ([]byte, error) {
	return m.CallDataFn()
}

func (m *Caller) Client() rpc.RPC {
	return m.ClientFn()
}

func (m *Caller) Call(ctx context.Context, number types.BlockNumber, res any) error {
	return m.CallFn(ctx, number, res)
}

func (m *Caller) Gas(ctx context.Context, number types.BlockNumber) (uint64, error) {
	return m.GasFn(ctx, number)
}

func (m *Caller) SendTransaction(ctx context.Context) (*types.Hash, *types.Transaction, error) {
	return m.SendTransactionFn(ctx)
}

func (m *Caller) DecodeTo(bytes []byte, a any) error {
	return m.DecodeToFn(bytes, a)
}

type TypedCaller[T any] struct {
	Caller

	CallFn   func(ctx context.Context, number types.BlockNumber) (T, error)
	DecodeFn func(bytes []byte) (T, error)
}

func (m *TypedCaller[T]) MockResult(result T, err error) *TypedCaller[T] {
	m.Caller.MockResult(result, err)
	m.CallFn = func(ctx context.Context, number types.BlockNumber) (T, error) {
		return result, err
	}
	m.DecodeFn = func(bytes []byte) (T, error) {
		return result, err
	}
	return m
}

func (m *TypedCaller[T]) MockReset(t *testing.T) *TypedCaller[T] {
	m.Caller.MockReset(t)
	m.CallFn = func(ctx context.Context, number types.BlockNumber) (T, error) {
		var zero T
		assert.FailNow(t, "unexpected call to Call")
		return zero, nil
	}
	m.DecodeFn = func(bytes []byte) (T, error) {
		var zero T
		assert.FailNow(t, "unexpected call to Decode")
		return zero, nil
	}
	return m
}

func (m *TypedCaller[T]) Call(ctx context.Context, number types.BlockNumber) (T, error) {
	return m.CallFn(ctx, number)
}

func (m *TypedCaller[T]) Decode(bytes []byte) (T, error) {
	return m.DecodeFn(bytes)
}
