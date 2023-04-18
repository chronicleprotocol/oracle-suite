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

package ethereum

import (
	"context"
	"math/big"

	"github.com/defiweb/go-eth/types"
)

// Client is an interface for Ethereum blockchain.
//
// Deprecated: use the github.com/defiweb/go-eth package instead.
type Client interface {
	// BlockNumber returns the current block number.
	BlockNumber(ctx context.Context) (*big.Int, error)

	// Block returns the block data. The block number can be changed by using
	// the WithBlockNumber context.
	Block(ctx context.Context) (*types.Block, error)

	// Call executes a message call transaction, which is directly
	// executed in the VM of the node, but never mined into the blockchain.
	Call(ctx context.Context, call types.Call) ([]byte, error)

	// CallBlocks executes the same call on multiple blocks (counting back from the latest)
	// and returns multiple results in a slice
	CallBlocks(ctx context.Context, call types.Call, blocks []int64) ([][]byte, error)

	// MultiCall works like the Call function but allows to execute multiple
	// calls at once.
	MultiCall(ctx context.Context, calls []types.Call) ([][]byte, error)

	// Storage returns the value of key in the contract storage of the
	// given account.
	Storage(ctx context.Context, address types.Address, key types.Hash) ([]byte, error)

	// Balance returns the wei balance of the given account.
	Balance(ctx context.Context, address types.Address) (*big.Int, error)

	// SendTransaction injects a signed transaction into the pending pool
	// for execution.
	SendTransaction(ctx context.Context, transaction *types.Transaction) (*types.Hash, error)

	// FilterLogs executes a filter query.
	FilterLogs(ctx context.Context, query types.FilterLogsQuery) ([]types.Log, error)
}

type contextKey string

const contextBlockNumber contextKey = "ethereum_block_number"

// WithBlockNumber sets the block number in the context.
//
// Deprecated: use the github.com/defiweb/go-eth package instead.
func WithBlockNumber(ctx context.Context, block *big.Int) context.Context {
	return context.WithValue(ctx, contextBlockNumber, block)
}

// BlockNumberFromContext returns the block number from the context.
//
// Deprecated: use the github.com/defiweb/go-eth package instead.
func BlockNumberFromContext(ctx context.Context) *big.Int {
	n, ok := ctx.Value(contextBlockNumber).(*big.Int)
	if ok {
		return n
	}
	return nil
}
