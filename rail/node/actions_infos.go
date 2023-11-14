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

package node

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

func AddrInfoChan(infos chan<- peer.AddrInfo) Action {
	return func(rail *Node) error {
		infos <- *host.InfoFromHost(rail.host)
		return nil
	}
}

func ConnectoPinger(ctx context.Context, infos <-chan peer.AddrInfo) Action {
	return func(rail *Node) error {
		pingService := ping.NewPingService(rail.host)
		go func() {
			for id := range infos {
				log.Debugw("connect", "id", id.ID, "addrs", id.Addrs)
				if err := rail.host.Connect(ctx, id); err != nil {
					log.Error(err)
					continue
				}
				res := <-pingService.Ping(ctx, id.ID)
				log.Infow("ping", "id", id.ID, "rtt", res.RTT.String())
			}
		}()
		return nil
	}
}
