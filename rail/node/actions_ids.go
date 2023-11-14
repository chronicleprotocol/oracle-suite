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
	"time"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

func IDExtractor() (Action, <-chan peer.ID) {
	idChan := make(chan peer.ID)
	return extractIDs(idChan), idChan
}
func extractIDs(ids chan<- peer.ID) Action {
	return func(rail *Node) error {
		sub, err := rail.host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
		if err != nil {
			return err
		}
		go func() {
			rail.wg.Add(1)
			defer rail.wg.Done()

			for {
				select {
				case <-rail.ctx.Done():
					closeSub(sub)
					return
				case e := <-sub.Out():
					t := e.(event.EvtPeerIdentificationCompleted)
					ids <- t.Peer
				}
			}
		}()
		return nil
	}
}

type PingResult struct {
	Peer  peer.ID
	RTT   time.Duration
	Error error
}

func PingIDsIntoChan(ids <-chan peer.ID, pings chan<- any) Action {
	log := log.Named("Pinger")
	return func(rail *Node) error {
		pingService := ping.NewPingService(rail.host)
		go func() {
			for id := range ids {
				res := <-pingService.Ping(rail.ctx, id)
				log.Infow("ping", "id", id, "rtt", res.RTT.String())
				pings <- PingResult{id, res.RTT, res.Error}
			}
		}()
		return nil
	}
}
