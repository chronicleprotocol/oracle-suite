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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
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

type PingResult struct {
	Peer  peer.ID
	RTT   time.Duration
	Error error
}

func Pinger(ctx context.Context, ids <-chan peer.ID, pings chan<- any) Action {
	log := log.Named("Pinger")
	return func(rail *Node) error {
		pingService := ping.NewPingService(rail.host)
		go func() {
			for id := range ids {
				res := <-pingService.Ping(ctx, id)
				log.Infow("ping", "id", id, "rtt", res.RTT.String())
				pings <- PingResult{id, res.RTT, res.Error}
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
	UserAgent     string
}

func getPeerUserAgent(ps peerstore.Peerstore, pid peer.ID) string {
	av, _ := ps.Get(pid, "AgentVersion")
	if s, ok := av.(string); ok {
		return s
	}
	return ""
}

func Events(ch chan<- any) Action {
	return func(rail *Node) error {
		sub, err := rail.host.EventBus().Subscribe(event.WildcardSubscription)
		if err != nil {
			return err
		}
		ps := rail.host.Peerstore()
		go func() {
			rail.wg.Add(1)
			defer rail.wg.Done()

			for {
				select {
				case <-rail.ctx.Done():
					closeSub(sub)
					// close(ch)
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
							UserAgent:     getPeerUserAgent(ps, t.Peer),
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

type PeerMessage struct {
	Msg       pubsub.Message
	UserAgent string
}

func GossipSub(ctx context.Context, msgChan chan<- any, opts ...pubsub.Option) Action {
	return func(rail *Node) error {
		ps := rail.host.Peerstore()
		pubSub, err := pubsub.NewGossipSub(ctx, rail.host, opts...)
		if err != nil {
			return err
		}

		type cancel struct {
			cancel pubsub.RelayCancelFunc
			topic  string
			sub    *pubsub.Subscription
		}
		var cancels []cancel

		for _, topic := range messages.AllMessagesMap.Keys() {
			t, err := pubSub.Join(topic)
			if err != nil {
				log.Errorw("error join", "topic", topic, "error", err)
				continue
			}
			log.Debugw("joined", "topic", topic)

			c, err := t.Relay()
			if err != nil {
				log.Errorw("error relay", "topic", topic, "error", err)
				continue
			}
			log.Debugw("relaying", "topic", topic)

			sub, err := t.Subscribe()
			if err != nil {
				log.Errorw("error subscribing", "topic", topic, "error", err)
				continue
			}
			log.Debugw("subscribed", "topic", topic)

			cancels = append(cancels, cancel{c, topic, sub})
			go func(topic string) {
				for {
					msg, err := sub.Next(ctx)
					if err != nil {
						log.Errorw("error getting next", "topic", topic, "error", err)
						return
					}
					msgChan <- PeerMessage{
						Msg:       *msg,
						UserAgent: getPeerUserAgent(ps, msg.GetFrom()),
					}
				}
			}(topic)
		}
		rail.wg.Add(len(cancels))
		go func() {
			<-rail.ctx.Done()
			for _, c := range cancels {
				log.Debugw("canceling subscription", "topic", c.topic)
				c.sub.Cancel()

				log.Debugw("canceling relay", "topic", c.topic)
				c.cancel()

				rail.wg.Done()
			}
			log.Debugf("all topics abandoned")
		}()

		return nil
	}
}

func Gossip(ctx context.Context, ch chan<- any) Action {
	var gSubOpts []pubsub.Option
	if directPeers := env.Strings("CFG_LIBP2P_DIRECT_PEERS_ADDRS", nil); len(directPeers) > 0 {
		gSubOpts = append(gSubOpts, pubsub.WithDirectPeers(addrInfos(directPeers)))
	}
	return GossipSub(ctx, ch, gSubOpts...)
}
