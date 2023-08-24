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
	"fmt"

	goethABI "github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
)

// simulateTransaction simulates a transaction by calling the contract method
// and checking for revert or panic.
func simulateTransaction(ctx context.Context, rpc rpc.RPC, tx types.Transaction) error {
	res, err := rpc.Call(ctx, tx.Call, types.LatestBlockNumber)
	if err != nil {
		return err
	}
	if goethABI.IsRevert(res) {
		return fmt.Errorf("transaction reverted: %v", goethABI.DecodeRevert(res))
	}
	if goethABI.IsPanic(res) {
		return fmt.Errorf("transaction panicked: %v", goethABI.DecodePanic(res))
	}
	return nil
}

// bytesToString converts a string terminated by a null byte to a Go string.
func bytesToString(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		return string(b)
	}
	return string(b[:n])
}
