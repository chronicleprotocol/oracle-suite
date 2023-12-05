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

type ConfigResult struct {
	Bar   int     `abi:"bar"`
	Bloom FeedIDs `abi:"bloom"`
}

// WatRegistry allows interacting with the WatRegistry contract.
type WatRegistry struct {
	client  rpc.RPC
	address types.Address
}

// NewWatRegistry creates a new WatRegistry instance.
func NewWatRegistry(client rpc.RPC, address types.Address) *WatRegistry {
	return &WatRegistry{
		client:  client,
		address: address,
	}
}

// Client returns the RPC client used to interact with the WatRegistry
func (w *WatRegistry) Client() rpc.RPC {
	return w.client
}

// Address returns the address of the WatRegistry contract.
func (w *WatRegistry) Address() types.Address {
	return w.address
}

// Wats returns all of Chronicle Protocol's wat identifiers.
func (w *WatRegistry) Wats() contract.TypedSelfCaller[[]string] {
	method := abiWatRegistry.Methods["wats"]
	return contract.NewTypedCall[[]string](
		contract.CallOpts{
			Client:       w.client,
			Address:      w.address,
			Encoder:      contract.NewCallEncoder(method),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiWatRegistry),
		},
	)
}

// Exists returns whether wat `wat` exists in registry.
func (w *WatRegistry) Exists(wat string) contract.TypedSelfCaller[bool] {
	method := abiWatRegistry.Methods["exists"]
	return contract.NewTypedCall[bool](
		contract.CallOpts{
			Client:       w.client,
			Address:      w.address,
			Encoder:      contract.NewCallEncoder(method, wat),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiWatRegistry),
		},
	)
}

// Config returns wat `wat`'s configuration.
func (w *WatRegistry) Config(wat string) contract.TypedSelfCaller[ConfigResult] {
	method := abiWatRegistry.Methods["config"]
	return contract.NewTypedCall[ConfigResult](
		contract.CallOpts{
			Client:       w.client,
			Address:      w.address,
			Encoder:      contract.NewCallEncoder(method, wat),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiWatRegistry),
		},
	)
}

// Chains returns wat `wat`'s list of chain ids its deployed to.
func (w *WatRegistry) Chains(wat string) contract.TypedSelfCaller[[]uint64] {
	method := abiWatRegistry.Methods["chains"]
	return contract.NewTypedCall[[]uint64](
		contract.CallOpts{
			Client:       w.client,
			Address:      w.address,
			Encoder:      contract.NewCallEncoder(method, wat),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiWatRegistry),
		},
	)
}

// Deployment returns wat `wat`'s deployment address for chain `chainId`.
func (w *WatRegistry) Deployment(wat string, chainID uint64) contract.TypedSelfCaller[types.Address] {
	method := abiWatRegistry.Methods["deployment"]
	return contract.NewTypedCall[types.Address](
		contract.CallOpts{
			Client:       w.client,
			Address:      w.address,
			Encoder:      contract.NewCallEncoder(method, wat, chainID),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiWatRegistry),
		},
	)
}
