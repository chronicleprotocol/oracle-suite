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

package starknet

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/chronicleprotocol/oracle-suite/internal/util/retry"
)

const retryAttempts = 3               // The maximum number of attempts to retry a call.
const retryInterval = 5 * time.Second // The delay between retry attempts.

type Scope string

const (
	ScopeTXNHash            Scope = "TXN_HASH"
	ScopeFullTXNs           Scope = "FULL_TXNS"
	ScopeFullTXNAndReceipts Scope = "FULL_TXN_AND_RECEIPTS"
)

// Client is a JSON RPC client of a Starknet node.
type Client struct {
	rpcClient *rpc.Client
}

// NewClient creates a new client for the given URL.
func NewClient(ctx context.Context, url string) (*Client, error) {
	rpcClient, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Client{rpcClient: rpcClient}, nil
}

// BlockNumber returns the current block number.
func (s *Client) BlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber uint64
	err := s.call(ctx, &blockNumber, "starknet_blockNumber")
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}

// GetBlockByNumber returns the block with the given number.
func (s *Client) GetBlockByNumber(ctx context.Context, blockNumber uint64, scope Scope) (*Block, error) {
	var block *Block
	err := s.call(ctx, &block, "starknet_getBlockByNumber", blockNumber, scope)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (s *Client) call(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	return retry.Retry(ctx, func() error {
		return s.rpcClient.CallContext(ctx, result, method, args...)
	}, retryAttempts, retryInterval)
}
