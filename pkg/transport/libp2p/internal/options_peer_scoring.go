//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package internal

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/internal/sets"
)

// PeerScoring configures peer scoring parameters used in a pubsub system.
func PeerScoring(
	params *pubsub.PeerScoreParams,
	thresholds *pubsub.PeerScoreThresholds,
	topicScoreParams func(topic string) *pubsub.TopicScoreParams) Options {

	return func(n *Node) error {
		n.pubsubOpts = append(
			n.pubsubOpts,
			pubsub.WithPeerScore(params, thresholds),
			pubsub.WithPeerScoreInspect(func(m map[peer.ID]*pubsub.PeerScoreSnapshot) {
				for id, ps := range m {
					n.tsLog.get().
						WithField("peerID", id).
						WithField("score", ps).
						Debug("Peer score")
				}
			}, time.Minute),
		)

		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event interface{}) {
			if e, ok := event.(sets.NodeTopicSubscribedEvent); ok {
				var err error
				defer func() {
					if err != nil {
						n.tsLog.get().
							WithError(err).
							WithField("topic", e.Topic).
							Warn("Unable to set topic score parameters")
					}
				}()
				sub, err := n.Subscription(e.Topic)
				if err != nil {
					return
				}
				if sp := topicScoreParams(e.Topic); sp != nil {
					n.tsLog.get().
						WithField("topic", e.Topic).
						WithField("params", sp).
						Info("Topic score params")
					err = sub.topic.SetScoreParams(sp)
					if err != nil {
						return
					}
				}
			}
		}))
		return nil
	}
}
