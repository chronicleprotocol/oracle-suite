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

type WatRegistry struct {
	client  rpc.RPC
	address types.Address
}

func NewWatRegistry(client rpc.RPC, address types.Address) *WatRegistry {
	return &WatRegistry{
		client:  client,
		address: address,
	}
}

func (w *WatRegistry) Address() types.Address {
	return w.address
}

func (w *WatRegistry) Bar(ctx context.Context, wat string) (int, error) {
	res, _, err := w.client.Call(
		ctx,
		types.Call{
			To:    &w.address,
			Input: errutil.Must(abiWatRegistry.Methods["bar"].EncodeArgs(stringToBytes32(wat))),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return 0, fmt.Errorf("WatRegistry: bar query failed: %v", err)
	}
	var bar uint8
	if err := abiWatRegistry.Methods["bar"].DecodeValues(res, &bar); err != nil {
		return 0, fmt.Errorf("WatRegistry: bar query failed: %v", err)
	}
	return int(bar), nil
}

func (w *WatRegistry) Feeds(ctx context.Context, wat string) ([]types.Address, error) {
	res, _, err := w.client.Call(
		ctx,
		types.Call{
			To:    &w.address,
			Input: errutil.Must(abiWatRegistry.Methods["feeds"].EncodeArgs(stringToBytes32(wat))),
		},
		types.LatestBlockNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("WatRegistry: feeds query failed: %v", err)
	}
	var feeds []types.Address
	if err := abiWatRegistry.Methods["feeds"].DecodeValues(res, &feeds); err != nil {
		return nil, fmt.Errorf("WatRegistry: feeds query failed: %v", err)
	}
	return feeds, nil
}
