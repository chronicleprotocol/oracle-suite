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
	"reflect"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

type Action func(*Rail) error

func AddrInfoChan(ids chan<- peer.AddrInfo) Action {
	return func(rail *Rail) error {
		ids <- *host.InfoFromHost(rail.host)
		return nil
	}
}

func ConnectoPinger(ctx context.Context, ids <-chan peer.AddrInfo) Action {
	return func(rail *Rail) error {
		pingService := ping.NewPingService(rail.host)
		go func() {
			for id := range ids {
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
	return func(rail *Rail) error {
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

func LogListeningAddresses(rail *Rail) error {
	addrs, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(rail.host))
	if err != nil {
		return err
	}
	log.Infow("listening", "addrs", addrs)
	return nil
}

func ExtractIDs(ids chan<- peer.ID) Action {
	return func(rail *Rail) error {
		sub, err := rail.host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
		if err != nil {
			return err
		}
		go func() {
			<-rail.errCh
			log.Debugw("closing subscription")
			if err := sub.Close(); err != nil {
				log.Errorw("error closing subscription", "error", err)
			}
		}()
		go func() {
			for e := range sub.Out() {
				t := e.(event.EvtPeerIdentificationCompleted)
				ids <- t.Peer
			}
		}()
		return nil
	}
}

func LogEvents(rail *Rail) error {
	ps := rail.host.Peerstore()
	sub, err := rail.host.EventBus().Subscribe(event.WildcardSubscription)
	if err != nil {
		return err
	}
	go func() {
		<-rail.errCh
		log.Debugw("closing subscription")
		if err := sub.Close(); err != nil {
			log.Errorw("error closing subscription", "error", err)
		}
	}()
	go func() {
		for e := range sub.Out() {
			switch t := e.(type) {
			case event.EvtLocalAddressesUpdated:
				log.Debugw("event", reflect.TypeOf(t).String(), e)
				var mas []multiaddr.Multiaddr
				for _, ma := range t.Current {
					mas = append(mas, ma.Address)
				}
				list, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{Addrs: mas, ID: rail.host.ID()})
				if err != nil {
					log.Errorw("error converting addr info", "error", err)
				}
				log.Infow("new listening", "addrs", list)
			case event.EvtPeerIdentificationCompleted:
				log.Debugw("event", reflect.TypeOf(t).String(), e)
				prots, err := ps.GetProtocols(t.Peer)
				if err != nil {
					log.Errorw("error getting protocols", "error", err)
				}
				log.Infow("protocols", t.Peer.String(), prots)
			default:
				log.Debugw("event", reflect.TypeOf(t).String(), e)
			}
		}
	}()
	return nil
}

func GossipSub(ctx context.Context, opts ...pubsub.Option) Action {
	return func(rail *Rail) error {
		ps, err := pubsub.NewGossipSub(ctx, rail.host, opts...)
		if err != nil {
			return err
		}

		var cancels []pubsub.RelayCancelFunc

		for _, topic := range messages.AllMessagesMap.Keys() {
			t, err := ps.Join(topic)
			if err != nil {
				log.Errorw("error joining topic", "topic", topic, "error", err)
			}
			c, err := t.Relay()
			if err != nil {
				log.Errorw("error enabling relay", "topic", topic, "error", err)
			}
			cancels = append(cancels, c)
		}

		go func() {
			<-rail.ctx.Done()
			log.Debug("canceling all topics")
			for _, c := range cancels {
				c()
			}
			log.Debug("all topics canceled")
		}()

		return nil
	}
}
