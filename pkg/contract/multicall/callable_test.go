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

package multicall

import (
	"context"
	"testing"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/mock"
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

func TestAggregatedCallables_Address(t *testing.T) {
	call1 := mock.NewCaller(t).MockAllowAllCalls()
	call2 := mock.NewCaller(t).MockAllowAllCalls()
	call1.AddressFn = func() types.Address { return types.MustAddressFromHex("0x1122334455667788990011223344556677889900") }
	call2.AddressFn = func() types.Address { return types.MustAddressFromHex("0x2233445566778899001122334455667788990011") }
	tests := []struct {
		name   string
		client rpc.RPC
		calls  []contract.Callable
		want   types.Address
	}{
		{
			name:   "no calls",
			client: &mockClient{},
			calls:  []contract.Callable{},
			want:   types.ZeroAddress,
		},
		{
			name:   "single call",
			client: &mockClient{},
			calls:  []contract.Callable{call1},
			want:   call1.Address(),
		},
		{
			name:   "two calls",
			client: &mockClient{},
			calls:  []contract.Callable{call1, call2},
			want:   multicallAddress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AggregateCallables(tt.client, tt.calls...)
			assert.Equal(t, tt.want, a.Address())
		})
	}
}

func TestAggregatedCallables_CallData(t *testing.T) {
	call1 := mock.NewCaller(t).MockAllowAllCalls()
	call2 := mock.NewCaller(t).MockAllowAllCalls()
	call3 := mock.NewCaller(t).MockAllowAllCalls()
	call1.AddressFn = func() types.Address { return types.MustAddressFromHex("0x1122334455667788990011223344556677889900") }
	call2.AddressFn = func() types.Address { return types.MustAddressFromHex("0x2233445566778899001122334455667788990011") }
	call3.AddressFn = func() types.Address { return types.ZeroAddress }
	call1.CallDataFn = func() ([]byte, error) { return []byte{0x01, 0x02, 0x03}, nil }
	call2.CallDataFn = func() ([]byte, error) { return []byte{0x04, 0x05, 0x06}, nil }
	call3.CallDataFn = func() ([]byte, error) { return []byte{0x07, 0x08, 0x09}, nil }
	tests := []struct {
		name   string
		client rpc.RPC
		calls  []contract.Callable
		want   []byte
	}{
		{
			name:   "no calls",
			client: &mockClient{},
			calls:  []contract.Callable{},
			want:   nil,
		},
		{
			name:   "single call",
			client: &mockClient{},
			calls:  []contract.Callable{call1},
			want:   []byte{0x01, 0x02, 0x03},
		},
		{
			name:   "two calls",
			client: &mockClient{},
			calls:  []contract.Callable{call1, call2},
			want: hexutil.MustHexToBytes(
				"0x82ad56cb" +
					"0000000000000000000000000000000000000000000000000000000000000020" +
					"0000000000000000000000000000000000000000000000000000000000000002" +
					"0000000000000000000000000000000000000000000000000000000000000040" +
					"00000000000000000000000000000000000000000000000000000000000000e0" +
					"0000000000000000000000001122334455667788990011223344556677889900" +
					"0000000000000000000000000000000000000000000000000000000000000000" +
					"0000000000000000000000000000000000000000000000000000000000000060" +
					"0000000000000000000000000000000000000000000000000000000000000003" +
					"0102030000000000000000000000000000000000000000000000000000000000" +
					"0000000000000000000000002233445566778899001122334455667788990011" +
					"0000000000000000000000000000000000000000000000000000000000000000" +
					"0000000000000000000000000000000000000000000000000000000000000060" +
					"0000000000000000000000000000000000000000000000000000000000000003" +
					"0405060000000000000000000000000000000000000000000000000000000000",
			),
		},
		{
			name:   "skip zero address call",
			client: &mockClient{},
			calls:  []contract.Callable{call1, call3},
			want:   []byte{0x01, 0x02, 0x03},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AggregateCallables(tt.client, tt.calls...)
			got, err := a.CallData()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAggregatedCallables_DecodeTo(t *testing.T) {
	call1 := mock.NewCaller(t).MockAllowAllCalls()
	call2 := mock.NewCaller(t).MockAllowAllCalls()
	call3 := mock.NewCaller(t).MockAllowAllCalls()
	call1.AddressFn = func() types.Address { return types.MustAddressFromHex("0x1122334455667788990011223344556677889900") }
	call2.AddressFn = func() types.Address { return types.MustAddressFromHex("0x2233445566778899001122334455667788990011") }
	call3.AddressFn = func() types.Address { return types.ZeroAddress }
	call1.DecodeToFn = func(bytes []byte, res any) error {
		*res.(*int) = 1
		return nil
	}
	call2.DecodeToFn = func(bytes []byte, res any) error {
		*res.(*int) = 2
		return nil
	}
	call3.DecodeToFn = func(bytes []byte, res any) error {
		*res.(*int) = 3
		return nil
	}
	tests := []struct {
		name     string
		client   rpc.RPC
		calls    []contract.Callable
		callData []byte
		want     []*int
	}{
		{
			name:     "no calls",
			client:   &mockClient{},
			calls:    []contract.Callable{},
			callData: nil,
			want:     []*int{},
		},
		{
			name:     "single call",
			client:   &mockClient{},
			calls:    []contract.Callable{call1},
			callData: hexutil.MustHexToBytes("0000000000000000000000000000000000000000000000000000000000000001"),
			want:     []*int{ptrInt(1)},
		},
		{
			name:   "two calls",
			client: &mockClient{},
			calls:  []contract.Callable{call1, call2},
			callData: hexutil.MustHexToBytes("" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000002" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"00000000000000000000000000000000000000000000000000000000000000c0" +
				"0000000000000000000000000000000000000000000000000000000000000001" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000001" +
				"0000000000000000000000000000000000000000000000000000000000000001" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000002"),
			want: []*int{ptrInt(1), ptrInt(2)},
		},
		{
			name:     "skip zero address call (two calls)",
			client:   &mockClient{},
			calls:    []contract.Callable{call1, call3},
			callData: hexutil.MustHexToBytes("0000000000000000000000000000000000000000000000000000000000000001"),
			want:     []*int{ptrInt(1), ptrInt(3)},
		},
		{
			name:   "skip zero address call (three calls)",
			client: &mockClient{},
			calls:  []contract.Callable{call1, call3, call2},
			callData: hexutil.MustHexToBytes("" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000002" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"00000000000000000000000000000000000000000000000000000000000000c0" +
				"0000000000000000000000000000000000000000000000000000000000000001" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000001" +
				"0000000000000000000000000000000000000000000000000000000000000001" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000002"),
			want: []*int{ptrInt(1), ptrInt(3), ptrInt(2)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AggregateCallables(tt.client, tt.calls...)
			got := make([]*int, len(tt.calls))
			err := a.DecodeTo(tt.callData, got)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func ptrInt(i int) *int {
	return &i
}
