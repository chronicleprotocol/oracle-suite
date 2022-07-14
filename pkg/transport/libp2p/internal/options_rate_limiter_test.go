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
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_RateLimiter_PeerLimit(t *testing.T) {
	// This test checks if peer limits works correctly. The limit is set to
	// 128 bytes/s, we will try to send two messages of 128 bytes each.
	// The second one must be rejected because it will exceed the 128 bytes/s
	// limit. Then we wait one second and try to send another 128 byte message.
	// This time the message should be accepted.

	peers, err := getNodeInfo(2)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	n0, err := NewNode(
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		RateLimiter(RateLimiterConfig{
			BytesPerSecond:      128,
			BurstSize:           128,
			RelayBytesPerSecond: 128,
			RelayBurstSize:      128,
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

	require.NoError(t, n1.Connect(peers[0].PeerAddrs[0]))
	_, err = n0.Subscribe("test")
	require.NoError(t, err)
	_, err = n1.Subscribe("test")
	require.NoError(t, err)

	s1, err := n0.Subscription("test")
	require.NoError(t, err)
	s2, err := n1.Subscription("test")
	require.NoError(t, err)

	// Wait for the peers to connect to each other:
	waitFor(t, func() bool {
		return len(n0.PubSub().ListPeers("test")) > 0 && len(n1.PubSub().ListPeers("test")) > 0
	})

	// Send messages:
	msgsCh := countMessages(s1, 2*time.Second)
	msg := []byte(strings.Repeat("a", 128))
	require.NoError(t, s2.Publish(msg))
	require.NoError(t, s2.Publish(msg)) // exceeds limit
	time.Sleep(1 * time.Second)
	require.NoError(t, s2.Publish(msg))

	// Only two messages should arrive, rest messages exceed the peer limit:
	assert.Equal(t, 2, (<-msgsCh)[n1.Host().ID()])
}

func TestNode_RateLimiter_PeerBurst(t *testing.T) {
	// This test checks if data burst for a peer works correctly. The value for
	// the data limit is smaller than the message size. We will try to send two
	// messages. The first one should be accepted because its size is within the
	// burst limit. The second one should be rejected because it exceeds the
	// burst limit.

	peers, err := getNodeInfo(2)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	n0, err := NewNode(
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		RateLimiter(RateLimiterConfig{
			BytesPerSecond:      1,
			BurstSize:           128,
			RelayBytesPerSecond: 1,
			RelayBurstSize:      128,
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

	require.NoError(t, n1.Connect(peers[0].PeerAddrs[0]))
	_, err = n0.Subscribe("test")
	require.NoError(t, err)
	_, err = n1.Subscribe("test")
	require.NoError(t, err)

	s1, err := n0.Subscription("test")
	require.NoError(t, err)
	s2, err := n1.Subscription("test")
	require.NoError(t, err)

	// Wait for the peers to connect to each other:
	waitFor(t, func() bool {
		return len(n0.PubSub().ListPeers("test")) > 0 && len(n1.PubSub().ListPeers("test")) > 0
	})

	// Send messages:
	msgsCh := countMessages(s1, 2*time.Second)
	msg := []byte(strings.Repeat("a", 128))
	require.NoError(t, s2.Publish(msg))
	require.NoError(t, s2.Publish(msg))

	// Only one message should arrive, second one exceeds the peer limit:
	assert.Equal(t, 1, (<-msgsCh)[n1.Host().ID()])
}
