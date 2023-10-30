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

package chronicle

import (
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
)

type TryGetResult struct {
	Ok      bool          `abi:"ok"`
	Address types.Address `abi:"address"`
}

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

func (w *Chainlog) TryGet(wat string) contract.TypedSelfCaller[TryGetResult] {
	method := abiChainlog.Methods["tryGet"]
	return contract.NewTypedCall[TryGetResult](
		contract.CallOpts{
			Client:       w.client,
			Address:      w.address,
			Encoder:      contract.NewCallEncoder(method, stringToBytes32(wat)),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiChainlog),
		},
	)
}
