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

package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
)

type client struct {
	rpc rpc.RPC
}

func (c *client) Storage(ctx context.Context, addr types.Address, idx int64) ([]byte, error) {
	hash, err := c.rpc.GetStorageAt(ctx, addr, types.MustHashFromBigInt(big.NewInt(idx)), types.LatestBlockNumber)
	if err != nil {
		return nil, err
	}
	return hash.Bytes(), nil
}

func (c *client) Call(ctx context.Context, call types.Call) ([]byte, error) {
	return c.rpc.Call(ctx, call, types.LatestBlockNumber)
}

func (c *client) MultiCall(ctx context.Context, calls []types.Call) ([][]byte, error) {
	type mCall struct {
		Target types.Address `abi:"target"`
		Data   []byte        `abi:"callData"`
	}
	var mCalls []mCall
	var mResults [][]byte
	for _, call := range calls {
		if call.To == nil {
			return nil, fmt.Errorf("multicall: call to nil address")
		}
		mCalls = append(mCalls, mCall{
			Target: *call.To,
			Data:   call.Input,
		})
	}
	chainID, err := c.rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("multicall: getting chain id failed: %w", err)
	}
	mContract, ok := multiCallContracts[chainID]
	if !ok {
		return nil, fmt.Errorf("multicall: unsupported chain id %d", chainID)
	}
	callata, err := mMethod.EncodeArgs(mCalls)
	if err != nil {
		return nil, fmt.Errorf("multicall: encoding arguments failed: %w", err)
	}
	resp, err := c.rpc.Call(ctx, types.Call{
		To:    &mContract,
		Input: callata,
	}, types.LatestBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("multicall: call failed: %w", err)
	}
	if err := mMethod.DecodeValues(resp, nil, &mResults); err != nil {
		return nil, fmt.Errorf("multicall: decoding results failed: %w", err)
	}
	return mResults, nil
}

var mMethod = abi.MustParseMethod(`
	function aggregate(
		(address target, bytes callData)[] memory calls
	) public returns (
		uint256 blockNumber, 
		bytes[] memory returnData
	)`,
)

func (c *client) Send(ctx context.Context, call types.Call) (*types.Hash, error) {
	// TODO implement me
	panic("implement me")
}
