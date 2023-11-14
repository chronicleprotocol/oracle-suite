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
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/libp2p/go-libp2p-pubsub"

	"github.com/chronicleprotocol/oracle-suite/rail/env"
)

type PeerMessage struct {
	Msg   *pubsub.Message
	Topic string
}

func MessagesIntoChan(ch chan<- any) Action {
	var gSubOpts []pubsub.Option
	if directPeers := env.Strings("CFG_LIBP2P_DIRECT_PEERS_ADDRS", nil); len(directPeers) > 0 {
		gSubOpts = append(gSubOpts, pubsub.WithDirectPeers(addrInfos(directPeers)))
	}
	return gossipSub(ch, gSubOpts...)
}

func gossipSub(msgChan chan<- any, opts ...pubsub.Option) Action {
	return func(rail *Node) error {
		pubSub, err := pubsub.NewGossipSub(rail.ctx, rail.host, opts...)
		if err != nil {
			return err
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

			cancels = append(cancels, cancel{topic, sub, cancelable(c)})

			go func(topic string) {
				for {
					message, err := sub.Next(rail.ctx)
					if err != nil {
						log.Errorw("error getting next", "topic", topic, "error", err)
						return
					}
					msgChan <- PeerMessage{
						Msg:   message,
						Topic: topic,
					}

				}
			}(topic)
		}

		rail.wg.Add(len(cancels))
		go func() {
			<-rail.ctx.Done()
			for _, c := range cancels {
				c.cancel()
				rail.wg.Done()
			}
			log.Debugf("all topics abandoned")
		}()

		return nil
	}
}

type cancel struct {
	topic string

	subscription *pubsub.Subscription
	relay        cancelable
}

func (c cancel) cancel() {
	log.Debugw("canceling subscription", "topic", c.topic)
	c.subscription.Cancel()

	log.Debugw("canceling relay", "topic", c.topic)
	c.relay.cancel()
}

type cancelable func()

func (c cancelable) cancel() {
	c()
}
