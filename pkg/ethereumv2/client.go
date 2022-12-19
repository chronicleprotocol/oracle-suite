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

package ethereumv2

import (
	"context"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereumv2/types"
)

// Client is a lightweight Ethereum RPC.
type Client interface {
	// BlockNumber performs eth_blockNumber RPC call.
	//
	// It returns the current block number.
	BlockNumber(ctx context.Context) (uint64, error)
	// BlockByNumber performs eth_getBlockByNumber RPC call.
	//
	// It returns the block with the given number.
	BlockByNumber(ctx context.Context, number types.BlockNumber) (*types.BlockTxHashes, error)
	// FullBlockByNumber performs eth_getBlockByNumber RPC call.
	//
	// It returns the block with the given number along with all transactions.
	FullBlockByNumber(ctx context.Context, number types.BlockNumber) (*types.BlockTxObjects, error)
	// GetTransactionCount performs eth_getTransactionCount RPC call.
	//
	// It returns the number of transactions sent from the given address.
	GetTransactionCount(ctx context.Context, account types.Address, block types.BlockNumber) (uint64, error)
	// SendRawTransaction performs eth_sendRawTransaction RPC call.
	//
	// It sends an encoded transaction to the network.
	SendRawTransaction(ctx context.Context, data types.Bytes) (*types.Hash, error)
	// GetStorageAt performs eth_getStorageAt RPC call.
	//
	// It returns the value of key in the contract storage at the given
	// address.
	GetStorageAt(ctx context.Context, acc types.Address, key types.Hash, block types.BlockNumber) (*types.Hash, error)
	// FilterLogs performs eth_getLogs RPC call.
	//
	// FilterLogs returns logs that match the given query.
	FilterLogs(ctx context.Context, q types.FilterLogsQuery) ([]types.Log, error)
}
