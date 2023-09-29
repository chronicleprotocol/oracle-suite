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

package service

import (
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
)

func Bootstrap(ctx context.Context, boots ...peer.AddrInfo) libp2p.Option {
	return libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
		if len(boots) == 0 {
			return nil, nil
		}
		log.Infow("creating DHT router", "boots", boots)
		return dual.New(
			ctx, host,
			dual.DHTOption(
				dht.Mode(dht.ModeAutoServer),
			),
			dual.WanDHTOption(
				dht.BootstrapPeers(boots...),
				dht.Mode(dht.ModeAutoServer),
			),
		)
	})
}
