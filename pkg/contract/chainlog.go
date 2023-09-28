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
	"context"
	"fmt"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

type Chainlog struct {
	client  rpc.RPC
	address types.Address
}

func NewChainlog(client rpc.RPC, address types.Address) *Chainlog {
	return &Chainlog{
		client:  client,
		address: address,
	}
}

func (w *Chainlog) Address() types.Address {
	return w.address
}

func (w *Chainlog) TryGet(ctx context.Context, wat string) (ok bool, address types.Address, ere error) {
	res, _, err := w.client.Call(
		ctx,
		types.Call{
			To:    &w.address,
			Input: errutil.Must(abiChainlog.Methods["tryGet"].EncodeArgs(stringToBytes32(wat))),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return false, types.ZeroAddress, fmt.Errorf("chainlog: tryGet query failed: %w", err)
	}
	if err := abiChainlog.Methods["tryGet"].DecodeValues(res, &ok, &address); err != nil {
		return false, types.ZeroAddress, fmt.Errorf("chainlog: tryGet query failed: %w", err)
	}
	return ok, address, nil
}
