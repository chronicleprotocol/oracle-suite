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

package geth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

const (
	mainnetChainID = 1
	kovanChainID   = 42
	rinkebyChainID = 4
	gorliChainID   = 5
	ropstenChainID = 3
	xdaiChainID    = 100
)

// Addresses of multicall contracts. They're used to implement
// the Client.MultiCall function.
//
// https://github.com/makerdao/multicall
var multiCallContracts = map[uint64]types.Address{
	mainnetChainID: types.MustAddressFromHex("0xeefba1e63905ef1d7acba5a8513c70307c1ce441"),
	kovanChainID:   types.MustAddressFromHex("0x2cc8688c5f75e365aaeeb4ea8d6a480405a48d2a"),
	rinkebyChainID: types.MustAddressFromHex("0x42ad527de7d4e9d9d011ac45b31d8551f8fe9821"),
	gorliChainID:   types.MustAddressFromHex("0x77dca2c955b15e9de4dbbcf1246b4b85b651e50e"),
	ropstenChainID: types.MustAddressFromHex("0x53c43764255c17bd724f74c4ef150724ac50a3ed"),
	xdaiChainID:    types.MustAddressFromHex("0xb5b692a88bdfc81ca69dcb1d924f59f0413a602a"),
}

// Deprecated: use the github.com/defiweb/go-eth package instead.
type Client struct{ client rpc.RPC }

// Deprecated: use the github.com/defiweb/go-eth package instead.
func NewClient(client rpc.RPC) *Client {
	return &Client{client: client}
}

func (c *Client) BlockNumber(ctx context.Context) (*big.Int, error) {
	bn, err := c.client.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	return bn, nil
}

func (c *Client) Block(ctx context.Context) (*types.Block, error) {
	return c.client.BlockByNumber(ctx, blockNumberFromContext(ctx), false)
}

func (c *Client) Call(ctx context.Context, call types.Call) ([]byte, error) {
	return c.client.Call(ctx, call, blockNumberFromContext(ctx))
}

func (c *Client) CallBlocks(ctx context.Context, call types.Call, blocks []int64) ([][]byte, error) {
	blockNumber, err := c.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block number: %w", err)
	}
	var res [][]byte
	for _, block := range blocks {
		bn := types.BlockNumberFromBigInt(new(big.Int).Sub(blockNumber, big.NewInt(block)))
		r, err := c.client.Call(ctx, call, bn)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

func (c *Client) MultiCall(ctx context.Context, calls []types.Call) ([][]byte, error) {
	type multicallCall struct {
		Target types.Address `abi:"target"`
		Data   []byte        `abi:"callData"`
	}
	var (
		multicallCalls   []multicallCall
		multicallResults [][]byte
	)
	for _, call := range calls {
		if call.To == nil {
			return nil, fmt.Errorf("multicall: call to nil address")
		}
		multicallCalls = append(multicallCalls, multicallCall{
			Target: *call.To,
			Data:   call.Input,
		})
	}
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("multicall: getting chain id failed: %w", err)
	}
	multicallContract, ok := multiCallContracts[chainID]
	if !ok {
		return nil, fmt.Errorf("multicall: unsupported chain id %d", chainID)
	}
	callata, err := multicallMethod.EncodeArgs(multicallCalls)
	if err != nil {
		return nil, fmt.Errorf("multicall: encoding arguments failed: %w", err)
	}
	resp, err := c.client.Call(ctx, types.Call{
		To:    &multicallContract,
		Input: callata,
	}, types.LatestBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("multicall: call failed: %w", err)
	}
	if err := multicallMethod.DecodeValues(resp, nil, &multicallResults); err != nil {
		return nil, fmt.Errorf("multicall: decoding results failed: %w", err)
	}
	return multicallResults, nil
}

func (c *Client) Storage(ctx context.Context, address types.Address, key types.Hash) ([]byte, error) {
	hash, err := c.client.GetStorageAt(ctx, address, key, blockNumberFromContext(ctx))
	if err != nil {
		return nil, err
	}
	return hash.Bytes(), nil
}

func (c *Client) Balance(ctx context.Context, address types.Address) (*big.Int, error) {
	balance, err := c.client.GetBalance(ctx, address, blockNumberFromContext(ctx))
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (c *Client) SendTransaction(ctx context.Context, transaction *types.Transaction) (*types.Hash, error) {
	var err error

	if transaction.From == nil {
		return nil, fmt.Errorf("transaction must have a sender")
	}

	tx := &types.Transaction{
		Call: types.Call{
			From:                 transaction.From,
			To:                   transaction.To,
			MaxPriorityFeePerGas: transaction.MaxPriorityFeePerGas,
			MaxFeePerGas:         transaction.MaxFeePerGas,
			GasLimit:             transaction.GasLimit,
		},
		Nonce:     transaction.Nonce,
		ChainID:   transaction.ChainID,
		Signature: transaction.Signature,
	}
	tx.Input = make([]byte, len(transaction.Input))
	copy(tx.Input, transaction.Input)

	// Fill optional values if necessary:
	if tx.Nonce == nil {
		nonce, err := c.client.GetTransactionCount(ctx, *tx.From, types.LatestBlockNumber)
		if err != nil {
			return nil, err
		}
		tx.SetNonce(nonce)
	}
	if tx.MaxFeePerGas == nil {
		suggestedGasTipPrice, err := c.client.MaxPriorityFeePerGas(ctx)
		if err != nil {
			return nil, err
		}
		tx.SetMaxPriorityFeePerGas(suggestedGasTipPrice)
	}
	if tx.MaxFeePerGas == nil {
		suggestedGasPrice, err := c.client.GasPrice(ctx)
		if err != nil {
			return nil, err
		}
		tx.SetMaxFeePerGas(new(big.Int).Mul(suggestedGasPrice, big.NewInt(2)))
	}
	if tx.ChainID == nil {
		chainID, err := c.client.ChainID(ctx)
		if err != nil {
			return nil, err
		}
		tx.SetChainID(chainID)
	}

	// Send transaction:
	hash, err := c.client.SendTransaction(ctx, *tx)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (c *Client) FilterLogs(ctx context.Context, query types.FilterLogsQuery) ([]types.Log, error) {
	return c.client.GetLogs(ctx, query)
}

func blockNumberFromContext(ctx context.Context) types.BlockNumber {
	bn := ethereum.BlockNumberFromContext(ctx)
	if bn == nil {
		return types.LatestBlockNumber
	}
	return types.BlockNumberFromBigInt(bn)
}

var multicallMethod = abi.MustParseMethod(`
	function aggregate(
		(address target, bytes callData)[] memory calls
	) public returns (
		uint256 blockNumber, 
		bytes[] memory returnData
	)`,
)
