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

type ConfigRegistry struct {
	client  rpc.RPC
	address types.Address
}

func NewConfigRegistry(client rpc.RPC, address types.Address) *ConfigRegistry {
	return &ConfigRegistry{
		client:  client,
		address: address,
	}
}

func (c *ConfigRegistry) Client() rpc.RPC {
	return c.client
}

func (c *ConfigRegistry) Address() types.Address {
	return c.address
}

func (c *ConfigRegistry) Latest() contract.TypedSelfCaller[string] {
	method := abiConfigRegistry.Methods["latest"]
	return contract.NewTypedCall[string](
		contract.CallOpts{
			Client:       c.client,
			Address:      c.address,
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiConfigRegistry),
		},
	)
}
