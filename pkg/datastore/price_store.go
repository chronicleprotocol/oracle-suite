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

package datastore

import (
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type FeederPrice struct {
	AssetPair string
	Feeder    ethereum.Address
}

type PriceStore interface {
	// Add adds a new price to the list. If a price from same feeder already
	// exists, the newer one will be used.
	Add(from ethereum.Address, msg *messages.Price)
	// All returns all prices.
	All() map[FeederPrice]*messages.Price
	// AssetPair returns all prices for given asset pair.
	AssetPair(assetPair string) []*messages.Price
	// Feeder returns the latest price for given asset pair sent by given feeder.
	Feeder(assetPair string, feeder ethereum.Address) *messages.Price
}
