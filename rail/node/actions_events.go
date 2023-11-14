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
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type LocalReachability struct {
	Peer         peer.ID
	Reachability network.Reachability
}
type LocalAddresses struct {
	Peer      peer.ID
	Addresses []multiaddr.Multiaddr
}
type PeerConnectedness struct {
	Peer          peer.ID
	Connectedness network.Connectedness
}

func EventsIntoChan(ch chan<- any) Action {
	return func(rail *Node) error {
		sub, err := rail.host.EventBus().Subscribe(event.WildcardSubscription)
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
					switch t := e.(type) {
					case event.EvtLocalReachabilityChanged:
						evt := LocalReachability{
							Peer:         rail.host.ID(),
							Reachability: t.Reachability,
						}
						ch <- evt
					case event.EvtLocalAddressesUpdated:
						evt := LocalAddresses{
							Peer: rail.host.ID(),
						}
						for _, ma := range t.Current {
							evt.Addresses = append(evt.Addresses, ma.Address)
						}
						ch <- evt
					case event.EvtPeerConnectednessChanged:
						ch <- PeerConnectedness{
							Peer:          t.Peer,
							Connectedness: t.Connectedness,
						}
					default:
						ch <- e
					}
				}
			}
		}()
		return nil
	}
}
