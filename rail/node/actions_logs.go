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
	"reflect"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func LogListeningAddresses(rail *Node) error {
	addrs, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(rail.host))
	if err != nil {
		return err
	}
	log.Infow("listening", "addrs", addrs)
	return nil
}

func LogEvents(rail *Node) error {
	ps := rail.host.Peerstore()
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
				case event.EvtLocalAddressesUpdated:
					var mas []multiaddr.Multiaddr
					for _, ma := range t.Current {
						mas = append(mas, ma.Address)
					}
					list, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{Addrs: mas, ID: rail.host.ID()})
					if err != nil {
						log.Errorw("error converting addr info", "error", err)
						continue
					}
					log.Infow("new listening", "addrs", list)
				case event.EvtPeerIdentificationCompleted:
					prots, err := ps.GetProtocols(t.Peer)
					if err != nil {
						log.Errorw("error getting protocols", "error", err)
						continue
					}
					log.Infow("protocols",
						t.Peer.String(), prots,
					)
				case event.EvtPeerConnectednessChanged:
					a := ethkey.PeerIDToAddress(t.Peer)
					log.Infow("connectedness",
						"state", t.Connectedness.String(),
						"peer", t.Peer.String(),
						"addr", a.String(),
						"net", Net(a),
					)
				case event.EvtLocalReachabilityChanged:
					log.Infow("reachability",
						"state", t.Reachability.String(),
					)
				case event.EvtNATDeviceTypeChanged:
					log.Infow("NAT device",
						"type", t.NatDeviceType.String(),
						"proto", t.TransportProtocol.String(),
					)
				default:
					log.Debugw("event", reflect.TypeOf(t).String(), e)
				}
			}
		}
	}()
	return nil
}
