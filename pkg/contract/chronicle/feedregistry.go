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

// FeedRegistry allows interacting with the FeedRegistry contract.
type FeedRegistry struct {
	client  rpc.RPC
	address types.Address
}

// NewFeedRegistry creates a new FeedRegistry instance.
func NewFeedRegistry(client rpc.RPC, address types.Address) *FeedRegistry {
	return &FeedRegistry{
		client:  client,
		address: address,
	}
}

// Client returns the RPC client used to interact with the FeedRegistry.
func (w *FeedRegistry) Client() rpc.RPC {
	return w.client
}

// Address returns the address of the FeedRegistry contract.
func (w *FeedRegistry) Address() types.Address {
	return w.address
}

// Feeds returns all of Chronicle Protocol's feeds.
func (w *FeedRegistry) Feeds() contract.TypedSelfCaller[[]types.Address] {
	method := abiFeedRegistry.Methods["feeds"]
	return contract.NewTypedCall[[]types.Address](
		contract.CallOpts{
			Client:       w.client,
			Address:      w.address,
			Encoder:      contract.NewCallEncoder(method),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiFeedRegistry),
		},
	)
}

// FeedExists returns whether address `feed` is a feed.
func (w *FeedRegistry) FeedExists(feed types.Address) contract.TypedSelfCaller[bool] {
	method := abiFeedRegistry.Methods["feeds(address)"]
	return contract.NewTypedCall[bool](
		contract.CallOpts{
			Client:       w.client,
			Address:      w.address,
			Encoder:      contract.NewCallEncoder(method, feed),
			Decoder:      contract.NewCallDecoder(method),
			ErrorDecoder: contract.NewContractErrorDecoder(abiFeedRegistry),
		},
	)
}
