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

package p2p

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scoringEventTracer struct {
	prunes int
	mu     sync.Mutex
}

func (s *scoringEventTracer) Trace(evt *pubsub_pb.TraceEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if evt.GetType() == pubsub_pb.TraceEvent_PRUNE {
		s.prunes++
	}
}

func (s *scoringEventTracer) Prunes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prunes
}

func TestNode_PeerScoring(t *testing.T) {
	// This test verifies that the integration with peer scoring is working
	// correctly.

	var peerScoreParams = &pubsub.PeerScoreParams{
		AppSpecificScore:            func(id peer.ID) float64 { return 0 },
		AppSpecificWeight:           1,
		IPColocationFactorWeight:    -1,
		IPColocationFactorThreshold: 1,
		DecayInterval:               1 * time.Minute,
		DecayToZero:                 0.01,
		RetainScore:                 10 * time.Second,
		Topics:                      make(map[string]*pubsub.TopicScoreParams),
	}

	var topicScoreParams = &pubsub.TopicScoreParams{
		TimeInMeshCap:                  0,
		TimeInMeshQuantum:              1 * time.Second,
		TopicWeight:                    1,
		FirstMessageDeliveriesWeight:   100,
		FirstMessageDeliveriesDecay:    0.999,
		FirstMessageDeliveriesCap:      300,
		InvalidMessageDeliveriesWeight: -100,
		InvalidMessageDeliveriesDecay:  0.999,
	}

	var thresholds = &pubsub.PeerScoreThresholds{
		GossipThreshold:   -100,
		PublishThreshold:  -200,
		GraylistThreshold: -300,
		AcceptPXThreshold: 0,
	}

	peers, err := getNodeInfo(2)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	et := &scoringEventTracer{}
	n0, err := NewNode(
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		PubsubEventTracer(et),
		PeerScoring(peerScoreParams, thresholds, func(topic string) *pubsub.TopicScoreParams {
			return topicScoreParams
		}),
	)
	require.NoError(t, err)
	require.NoError(t, n0.Start(ctx))
	time.Sleep(time.Second)

	n1, err := NewNode(
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start(ctx))
	time.Sleep(time.Second)

	// Add validator to the n0 node which will reject all received messages
	// from the second node:
	n0.AddValidator(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		if n0.Host().ID() == id {
			return pubsub.ValidationAccept
		}
		if bytes.Equal(msg.Data, []byte("valid")) {
			return pubsub.ValidationAccept
		}
		return pubsub.ValidationReject
	})

	require.NoError(t, n1.Connect(peers[0].PeerAddrs[0]))
	_, err = n0.Subscribe("test")
	require.NoError(t, err)
	_, err = n1.Subscribe("test")
	require.NoError(t, err)

	s0, err := n0.Subscription("test")
	require.NoError(t, err)
	s1, err := n1.Subscription("test")
	require.NoError(t, err)

	// Wait for the peers to connect to each other:
	waitFor(t, func() bool {
		return len(n0.PubSub().ListPeers("test")) > 0 && len(n1.PubSub().ListPeers("test")) > 0
	})

	s0wait := countMessages(s0, 2*time.Second)

	// Send a few valid messages to boost a peer score:
	for i := 0; i < 5; i++ {
		err := s1.Publish([]byte("valid"))
		if err != nil {
			panic(err)
		}
	}

	// Now send 4 invalid messages, that should be enough to lower the score below 0:
	for i := 0; i < 4; i++ {
		err := s1.Publish([]byte("invalid"))
		if err != nil {
			panic(err)
		}
	}
	<-s0wait

	// Wait 1 second for the "heartbeat".
	time.Sleep(1 * time.Second)

	// The n1 node should be pruned:
	assert.Equal(t, 1, et.Prunes())
}
