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

func Pinger(ids <-chan peer.AddrInfo) Action {
	return func(rail *Rail) error {
		ctx := context.Background()
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

func LogListeningAddresses(rail *Rail) error {
	addrs, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(rail.host))
	if err != nil {
		return err
	}
	log.Infow("listening", "addrs", addrs)
	return nil
}

func PingAll() Action {
	return func(rail *Rail) error {
		ctx := context.Background()
		sub, err := rail.host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
		if err != nil {
			return err
		}
		go func() {
			<-rail.wait
			log.Debugw("closing subscription")
			if err := sub.Close(); err != nil {
				log.Errorw("error closing subscription", "error", err)
			}
		}()
		pingService := ping.NewPingService(rail.host)
		go func() {
			for e := range sub.Out() {
				t := e.(event.EvtPeerIdentificationCompleted)
				res := <-pingService.Ping(ctx, t.Peer)
				log.Infow("ping", "id", t.Peer, "rtt", res.RTT.String())
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
		<-rail.wait
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
