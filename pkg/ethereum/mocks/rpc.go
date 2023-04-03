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
	"math/big"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/mock"
)

type RPC struct {
	mock.Mock
}

func (r *RPC) ChainID(ctx context.Context) (uint64, error) {
	args := r.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (r *RPC) GasPrice(ctx context.Context) (*big.Int, error) {
	args := r.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (r *RPC) Accounts(ctx context.Context) ([]types.Address, error) {
	args := r.Called(ctx)
	return args.Get(0).([]types.Address), args.Error(1)
}

func (r *RPC) BlockNumber(ctx context.Context) (*big.Int, error) {
	args := r.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (r *RPC) GetBalance(ctx context.Context, address types.Address, block types.BlockNumber) (*big.Int, error) {
	args := r.Called(ctx, address, block)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (r *RPC) GetStorageAt(ctx context.Context, account types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error) {
	args := r.Called(ctx, account, key, block)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (r *RPC) GetTransactionCount(ctx context.Context, account types.Address, block types.BlockNumber) (uint64, error) {
	args := r.Called(ctx, account, block)
	return args.Get(0).(uint64), args.Error(1)
}

func (r *RPC) GetBlockTransactionCountByHash(ctx context.Context, hash types.Hash) (uint64, error) {
	args := r.Called(ctx, hash)
	return args.Get(0).(uint64), args.Error(1)
}

func (r *RPC) GetBlockTransactionCountByNumber(ctx context.Context, number types.BlockNumber) (uint64, error) {
	args := r.Called(ctx, number)
	return args.Get(0).(uint64), args.Error(1)
}

func (r *RPC) GetUncleCountByBlockHash(ctx context.Context, hash types.Hash) (uint64, error) {
	args := r.Called(ctx, hash)
	return args.Get(0).(uint64), args.Error(1)
}

func (r *RPC) GetUncleCountByBlockNumber(ctx context.Context, number types.BlockNumber) (uint64, error) {
	args := r.Called(ctx, number)
	return args.Get(0).(uint64), args.Error(1)
}

func (r *RPC) GetCode(ctx context.Context, account types.Address, block types.BlockNumber) ([]byte, error) {
	args := r.Called(ctx, account, block)
	return args.Get(0).([]byte), args.Error(1)
}

func (r *RPC) Sign(ctx context.Context, account types.Address, data []byte) (*types.Signature, error) {
	args := r.Called(ctx, account, data)
	return args.Get(0).(*types.Signature), args.Error(1)
}

func (r *RPC) SignTransaction(ctx context.Context, tx types.Transaction) ([]byte, *types.Transaction, error) {
	args := r.Called(ctx, tx)
	return args.Get(0).([]byte), args.Get(1).(*types.Transaction), args.Error(2)
}

func (r *RPC) SendTransaction(ctx context.Context, tx types.Transaction) (*types.Hash, error) {
	args := r.Called(ctx, tx)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (r *RPC) SendRawTransaction(ctx context.Context, data []byte) (*types.Hash, error) {
	args := r.Called(ctx, data)
	return args.Get(0).(*types.Hash), args.Error(1)
}

func (r *RPC) Call(ctx context.Context, call types.Call, block types.BlockNumber) ([]byte, error) {
	args := r.Called(ctx, call, block)
	return args.Get(0).([]byte), args.Error(1)
}

func (r *RPC) EstimateGas(ctx context.Context, call types.Call, block types.BlockNumber) (uint64, error) {
	args := r.Called(ctx, call, block)
	return args.Get(0).(uint64), args.Error(1)
}

func (r *RPC) BlockByHash(ctx context.Context, hash types.Hash, full bool) (*types.Block, error) {
	args := r.Called(ctx, hash, full)
	return args.Get(0).(*types.Block), args.Error(1)
}

func (r *RPC) BlockByNumber(ctx context.Context, number types.BlockNumber, full bool) (*types.Block, error) {
	args := r.Called(ctx, number, full)
	return args.Get(0).(*types.Block), args.Error(1)
}

func (r *RPC) GetTransactionByHash(ctx context.Context, hash types.Hash) (*types.OnChainTransaction, error) {
	args := r.Called(ctx, hash)
	return args.Get(0).(*types.OnChainTransaction), args.Error(1)
}

func (r *RPC) GetTransactionByBlockHashAndIndex(ctx context.Context, hash types.Hash, index uint64) (*types.OnChainTransaction, error) {
	args := r.Called(ctx, hash, index)
	return args.Get(0).(*types.OnChainTransaction), args.Error(1)
}

func (r *RPC) GetTransactionByBlockNumberAndIndex(ctx context.Context, number types.BlockNumber, index uint64) (*types.OnChainTransaction, error) {
	args := r.Called(ctx, number, index)
	return args.Get(0).(*types.OnChainTransaction), args.Error(1)
}

func (r *RPC) GetTransactionReceipt(ctx context.Context, hash types.Hash) (*types.TransactionReceipt, error) {
	args := r.Called(ctx, hash)
	return args.Get(0).(*types.TransactionReceipt), args.Error(1)
}

func (r *RPC) GetLogs(ctx context.Context, query types.FilterLogsQuery) ([]types.Log, error) {
	args := r.Called(ctx, query)
	return args.Get(0).([]types.Log), args.Error(1)
}

func (r *RPC) MaxPriorityFeePerGas(ctx context.Context) (*big.Int, error) {
	args := r.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (r *RPC) SubscribeLogs(ctx context.Context, query types.FilterLogsQuery) (chan types.Log, error) {
	args := r.Called(ctx, query)
	return args.Get(0).(chan types.Log), args.Error(1)
}

func (r *RPC) SubscribeNewHeads(ctx context.Context) (chan types.Block, error) {
	args := r.Called(ctx)
	return args.Get(0).(chan types.Block), args.Error(1)
}

func (r *RPC) SubscribeNewPendingTransactions(ctx context.Context) (chan types.Hash, error) {
	args := r.Called(ctx)
	return args.Get(0).(chan types.Hash), args.Error(1)
}

var _ rpc.RPC = (*RPC)(nil)
