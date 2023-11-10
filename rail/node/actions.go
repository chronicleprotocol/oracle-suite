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
	"reflect"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/defiweb/go-eth/types"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	"github.com/chronicleprotocol/oracle-suite/rail/env"
)

type Action func(*Node) error

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

func Pinger(ctx context.Context, ids <-chan peer.ID) Action {
	return func(rail *Node) error {
		pingService := ping.NewPingService(rail.host)
		go func() {
			for id := range ids {
				res := <-pingService.Ping(ctx, id)
				log.Infow("ping", "id", id, "rtt", res.RTT.String())
			}
		}()
		return nil
	}
}

func LogListeningAddresses(rail *Node) error {
	addrs, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(rail.host))
	if err != nil {
		return err
	}
	log.Infow("listening", "addrs", addrs)
	return nil
}

func ExtractIDs(ids chan<- peer.ID) Action {
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

func Events(ch chan<- any) Action {
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
					close(ch)
					return
				case e := <-sub.Out():
					ch <- e
				}
			}
		}()
		return nil
	}
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
						"net", feeds.net(a),
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
					_ = t
					log.Debugw("event",
						reflect.TypeOf(t).String(), e,
					)
				}
			}
		}
	}()
	return nil
}

func closeSub(sub event.Subscription) {
	log.Debugw("closing", "subscription", sub.Name())
	go func() {
		log.Debugf("draining %T", sub.Out())
		for e := range sub.Out() {
			log.Debugf("got %T for %s", e, sub.Name())
		}
	}()
	if err := sub.Close(); err != nil {
		log.Errorw("error closing", "error", err, "subscription", sub.Name())
		return
	}
	log.Debugw("closed", "subscription", sub.Name())
}

var feeds = feedList{
	"prod": {
		"0x130431b4560Cd1d74A990AE86C337a33171FF3c6",
		"0x16655369Eb59F3e1cAFBCfAC6D3Dd4001328f747",
		"0x3CB645a8f10Fb7B0721eaBaE958F77a878441Cb9",
		"0x4b0E327C08e23dD08cb87Ec994915a5375619aa2",
		"0x4f95d9B4D842B2E2B1d1AC3f2Cf548B93Fd77c67",
		"0x60da93D9903cb7d3eD450D4F81D402f7C4F71dd9",
		"0x71eCFF5261bAA115dcB1D9335c88678324b8A987",
		"0x75ef8432566A79C86BBF207A47df3963B8Cf0753",
		"0x77EB6CF8d732fe4D92c427fCdd83142DB3B742f7",
		"0x83e23C207a67a9f9cB680ce84869B91473403e7d",
		"0x8aFBD9c3D794eD8DF903b3468f4c4Ea85be953FB",
		"0x8de9c5F1AC1D4d02bbfC25fD178f5DAA4D5B26dC",
		"0x8ff6a38A1CD6a42cAac45F08eB0c802253f68dfD",
		"0xa580BBCB1Cee2BCec4De2Ea870D20a12A964819e",
		"0xA8EB82456ed9bAE55841529888cDE9152468635A",
		"0xaC8519b3495d8A3E3E44c041521cF7aC3f8F63B3",
		"0xc00584B271F378A0169dd9e5b165c0945B4fE498",
		"0xC9508E9E3Ccf319F5333A5B8c825418ABeC688BA",
		"0xD09506dAC64aaA718b45346a032F934602e29cca",
		"0xD27Fa2361bC2CfB9A591fb289244C538E190684B",
		"0xd72BA9402E9f3Ff01959D6c841DDD13615FFff42",
		"0xd94BBe83b4a68940839cD151478852d16B3eF891",
		"0xDA1d2961Da837891f43235FddF66BAD26f41368b",
		"0xE6367a7Da2b20ecB94A25Ef06F3b551baB2682e6",
		"0xFbaF3a7eB4Ec2962bd1847687E56aAEE855F5D00",
		"0xfeEd00AA3F0845AFE52Df9ECFE372549B74C69D2",
	},
	"stage": {
		"0x0c4FC7D66b7b6c684488c1F218caA18D4082da18",
		"0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD",
		"0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9",
		"0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4",
	},
}

type feedList map[string][]string

func (f feedList) net(a types.Address) string {
	for net, feeds := range f {
		for _, feed := range feeds {
			if feed == a.String() {
				return net
			}
		}
	}
	return ""
}

func GossipSub(ctx context.Context, opts ...pubsub.Option) Action {
	return func(rail *Node) error {
		ps, err := pubsub.NewGossipSub(ctx, rail.host, opts...)
		if err != nil {
			return err
		}

		type cancel struct {
			cancel pubsub.RelayCancelFunc
			topic  string
		}
		var cancels []cancel

		for _, topic := range messages.AllMessagesMap.Keys() {
			t, err := ps.Join(topic)
			if err != nil {
				log.Errorw("error joining topic", "topic", topic, "error", err)
				continue
			}
			log.Debugw("joined topic", "topic", topic)

			c, err := t.Relay()
			if err != nil {
				log.Errorw("error enabling relay", "topic", topic, "error", err)
				continue
			}
			log.Debugw("enabled relay", "topic", topic)

			cancels = append(cancels, cancel{c, topic})
		}
		rail.wg.Add(len(cancels))
		go func() {
			<-rail.ctx.Done()
			for _, c := range cancels {
				log.Debugw("canceling relay", "topic", c.topic)
				c.cancel()
				rail.wg.Done()
			}
			log.Debugf("all relays canceled")
		}()

		return nil
	}
}

func Gossip(ctx context.Context) Action {
	var gSubOpts []pubsub.Option
	if directPeers := env.Strings("CFG_LIBP2P_DIRECT_PEERS_ADDRS", nil); len(directPeers) > 0 {
		gSubOpts = append(gSubOpts, pubsub.WithDirectPeers(addrInfos(directPeers)))
	}
	return GossipSub(ctx, gSubOpts...)
}
