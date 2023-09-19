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
	"bytes"
	"context"
	"errors"

	goethABI "github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/rpc/transport"
	"github.com/defiweb/go-eth/types"
)

// simulateTransaction simulates a transaction by calling the contract method
// and checking for revert or panic.
func simulateTransaction(ctx context.Context, rpc rpc.RPC, c *goethABI.Contract, tx types.Transaction) error {
	_, _, err := rpc.Call(ctx, tx.Call, types.LatestBlockNumber)
	if err != nil {
		var rpcErr *transport.RPCError
		if errors.As(err, &rpcErr) {
			data, ok := rpcErr.Data.([]byte)
			if !ok {
				return err
			}
			if err := c.ToError(data); err != nil {
				return err
			}
		}
		return err
	}
	return nil
}

// stringToBytes32 converts a Go string to bytes32.
func stringToBytes32(s string) []byte {
	b := make([]byte, 32)
	copy(b, s)
	return b
}

// bytes32ToString converts bytes32 to a Go string.
func bytes32ToString(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		return string(b)
	}
	return string(b[:n])
}
