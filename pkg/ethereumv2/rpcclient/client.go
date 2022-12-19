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

package rpcclient

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereumv2/types"
)

// Client is a lightweight Ethereum RPC client that aims to be compatible with
// different Ethereum based blockchains.
//
// Unlike the official go-ethereum client, this client does not verify
// responses from the server. This is necessary because responses from some
// blockchains (such as Arbitrum) are not compatible with the go-ethereum
// client, so they do not pass verification.
type Client struct {
	rpc *rpc.Client
}

// New returns a new Client instance.
func New(rpc *rpc.Client) *Client {
	return &Client{rpc: rpc}
}

// TODO: eth_call

// BlockNumber implements the ethereumv2.Client.
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	var number types.Number
	if err := c.rpc.CallContext(ctx, &number, "eth_blockNumber"); err != nil {
		return 0, err
	}
	return number.Big().Uint64(), nil
}

// TODO: eth_getBlockByHash

// BlockByNumber implements the ethereumv2.Client.
func (c *Client) BlockByNumber(ctx context.Context, number types.BlockNumber) (*types.BlockTxHashes, error) {
	var block *types.BlockTxHashes
	if err := c.rpc.CallContext(ctx, &block, "eth_getBlockByNumber", number, false); err != nil {
		return nil, err
	}
	return block, nil
}

// FullBlockByNumber implements the ethereumv2.Client.
func (c *Client) FullBlockByNumber(ctx context.Context, number types.BlockNumber) (*types.BlockTxObjects, error) {
	var block *types.BlockTxObjects
	if err := c.rpc.CallContext(ctx, &block, "eth_getBlockByNumber", number, true); err != nil {
		return nil, err
	}
	return block, nil
}

// TODO: eth_getTransactionByHash

// GetTransactionCount implements the ethereumv2.Client.
func (c *Client) GetTransactionCount(ctx context.Context, acc types.Address, block types.BlockNumber) (uint64, error) {
	var count types.Number
	if err := c.rpc.CallContext(ctx, &count, "eth_getTransactionCount", acc, block); err != nil {
		return 0, err
	}
	return count.Big().Uint64(), nil
}

// TODO: eth_getTransactionReceipt
// TODO: eth_getBlockTransactionCountByHash
// TODO: eth_getBlockTransactionCountByNumber
// TODO: eth_getTransactionByBlockHashAndIndex
// TODO: eth_getTransactionByBlockNumberAndIndex
// TODO: eth_getBlockReceipts

// SendRawTransaction implements the ethereumv2.Client.
func (c *Client) SendRawTransaction(ctx context.Context, data types.Bytes) (*types.Hash, error) {
	txHash := &types.Hash{}
	if err := c.rpc.CallContext(ctx, txHash, "eth_sendRawTransaction", data); err != nil {
		return nil, err
	}
	return txHash, nil
}

// TODO: eth_sendPrivateTransaction
// TODO: eth_cancelPrivateTransaction
// TODO: eth_getBalance

// GetStorageAt implements the ethereumv2.Client.
func (c *Client) GetStorageAt(
	ctx context.Context,
	account types.Address,
	key types.Hash,
	block types.BlockNumber) (*types.Hash, error) {

	bytes := &types.Hash{}
	if err := c.rpc.CallContext(ctx, bytes, "eth_getStorageAt", account, key, block); err != nil {
		return nil, err
	}
	return bytes, nil
}

// TODO: eth_getCode
// TODO: eth_accounts
// TODO: eth_getProof
// TODO: eth_call

// FilterLogs implements the ethereumv2.Client.
func (c *Client) FilterLogs(ctx context.Context, q types.FilterLogsQuery) ([]types.Log, error) {
	var logs []types.Log
	if err := c.rpc.CallContext(ctx, &logs, "eth_getLogs", q); err != nil {
		return nil, err
	}
	return logs, nil
}

// TODO: eth_getFilterChanges
// TODO: eth_getFilterLogs
// TODO: eth_newBlockFilter
// TODO: eth_newFilter
// TODO: eth_newPendingTransactionFilter
// TODO: eth_uninstallFilter
// TODO: eth_protocolVersion
// TODO: eth_gasPrice
// TODO: eth_estimateGas
// TODO: eth_feeHistory
// TODO: eth_maxPriorityFeePerGas
// TODO: eth_chainId
// TODO: net_version
// TODO: net_listening
// TODO: eth_getUncleByBlockHashAndIndex
// TODO: eth_getUncleByBlockNumberAndIndex
// TODO: eth_getUncleCountByBlockHash
// TODO: eth_getUncleCountByBlockNumber
