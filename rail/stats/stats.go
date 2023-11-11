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

package stats

import (
	"github.com/defiweb/go-eth/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Peer struct {
	ID             peer.ID
	Addr           types.Address
	Connectedness  string
	Network        string
	Identification string
	Ping           string
	Topic          string
	Name           string
	Counter        int
	UserAgent      string
}
type PeerEvent any
