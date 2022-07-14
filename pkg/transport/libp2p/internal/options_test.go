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
	"crypto/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_PeerPrivKey(t *testing.T) {
	// This tests checks the PeerPrivKey option.

	sk, _, _ := crypto.GenerateRSAKeyPair(2048, rand.Reader)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	n, err := NewNode(
		PeerPrivKey(sk),
	)
	require.NoError(t, err)
	require.NoError(t, n.Start(ctx))
	time.Sleep(time.Second)

	id, _ := peer.IDFromPrivateKey(sk)
	assert.Equal(t, id, n.Host().ID())
}

func TestNode_MessagePrivKey(t *testing.T) {
	// This tests checks the MessagePrivKey option.

	sk, _, _ := crypto.GenerateRSAKeyPair(2048, rand.Reader)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	n, err := NewNode(
		MessagePrivKey(sk),
	)
	require.NoError(t, err)
	require.NoError(t, n.Start(ctx))
	time.Sleep(time.Second)

	_, err = n.Subscribe("test")
	require.NoError(t, err)
	s, err := n.Subscription("test")
	require.NoError(t, err)

	err = s.Publish([]byte("makerdao"))
	require.NoError(t, err)

	// The public key used to sign the message should be derived from the key
	// passed to the MessagePrivKey function:
	id, _ := peer.IDFromPrivateKey(sk)
	msg := <-s.Next()
	assert.Equal(t, id, msg.GetFrom())
	// The public key extracted form a message must be different
	// than peer's public key:
	assert.NotEqual(t, n.Host().ID(), msg.GetFrom())
}

func TestNode_Discovery(t *testing.T) {
	// This test checks whether all nodes in the network can discover each
	// other when Discovery option is used.
	//
	// Topology:
	//   n0 <--[discovery]--> n1 <--[discovery]--> n2

	peers, err := getNodeInfo(3)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	n0, err := NewNode(
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		Discovery(nil),
	)
	require.NoError(t, err)
	require.NoError(t, n0.Start(ctx))
	time.Sleep(time.Second)

	n1, err := NewNode(
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
		Discovery(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start(ctx))
	time.Sleep(time.Second)

	n2, err := NewNode(
		PeerPrivKey(peers[2].PrivKey),
		ListenAddrs(peers[2].ListenAddrs),
		Discovery(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n2.Start(ctx))
	time.Sleep(time.Second)

	_, err = n0.Subscribe("test")
	require.NoError(t, err)
	_, err = n1.Subscribe("test")
	require.NoError(t, err)
	_, err = n2.Subscribe("test")
	require.NoError(t, err)

	// Every peer should see two other peers:
	waitFor(t, func() bool {
		lp := n0.PubSub().ListPeers("test")
		return containsPeerID(lp, peers[1].ID) && containsPeerID(lp, peers[2].ID)
	})
	waitFor(t, func() bool {
		lp := n1.PubSub().ListPeers("test")
		return containsPeerID(lp, peers[0].ID) && containsPeerID(lp, peers[2].ID)
	})
	waitFor(t, func() bool {
		lp := n2.PubSub().ListPeers("test")
		return containsPeerID(lp, peers[0].ID) && containsPeerID(lp, peers[1].ID)
	})
}

func TestNode_Discovery_AddrNotLeaking(t *testing.T) {
	// This test checks that the addresses of nodes that do not use discovery
	// are not revealed to other nodes in the network.
	//
	// Topology:
	//   n0 <--[discovery]--> n1 <--[direct]--> n2
	//
	// - n0 node should only be connected to n1 node (through discovery)
	// - n1 node should only be connected to n0 node (through discovery) and n1 (through direct connection)
	// - n2 node should only be connected to n1 node (through direct connection)
	// - the n0 node's address must not be exposed to n2 node

	peers, err := getNodeInfo(3)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	n0, err := NewNode(
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		Discovery(nil),
	)
	require.NoError(t, err)
	require.NoError(t, n0.Start(ctx))
	time.Sleep(time.Second)

	n1, err := NewNode(
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
		DirectPeers(peers[2].PeerAddrs),
		Discovery(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start(ctx))
	time.Sleep(time.Second)

	n2, err := NewNode(
		PeerPrivKey(peers[2].PrivKey),
		ListenAddrs(peers[2].ListenAddrs),
		DirectPeers(peers[1].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n2.Start(ctx))
	time.Sleep(time.Second)

	_, err = n0.Subscribe("test")
	require.NoError(t, err)
	_, err = n1.Subscribe("test")
	require.NoError(t, err)
	_, err = n2.Subscribe("test")
	require.NoError(t, err)

	waitFor(t, func() bool {
		lp := n0.PubSub().ListPeers("test")
		return len(lp) == 1 && containsPeerID(lp, peers[1].ID)
	})
	waitFor(t, func() bool {
		lp := n1.PubSub().ListPeers("test")
		return len(lp) == 2 && containsPeerID(lp, peers[0].ID) && containsPeerID(lp, peers[2].ID)
	})
	waitFor(t, func() bool {
		lp := n2.PubSub().ListPeers("test")
		return len(lp) == 1 && containsPeerID(lp, peers[1].ID)
	})
}

func TestNode_ConnectionLimit(t *testing.T) {
	// This test checks whether the connection number is properly limited when
	// the ConnectionLimit option is used.

	peers, err := getNodeInfo(5)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	n, err := NewNode(
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		ConnectionLimit(1, 1, 0),
	)
	require.NoError(t, err)
	require.NoError(t, n.Start(ctx))
	time.Sleep(time.Second)

	for i := 2; i < len(peers); i++ {
		n, err := NewNode(
			PeerPrivKey(peers[i].PrivKey),
			ListenAddrs(peers[i].ListenAddrs),
			Discovery(nil),
		)
		require.NoError(t, err)
		require.NoError(t, n.Start(ctx))
		time.Sleep(time.Second)

		require.NoError(t, n.Connect(peers[0].PeerAddrs[0]))
	}

	n.Host().ConnManager().TrimOpenConns(context.Background())
	time.Sleep(time.Second)

	conns := 0
	for _, p := range n.Host().Peerstore().Peers() {
		if n.Host().Network().Connectedness(p) == network.Connected {
			conns++
		}
	}

	assert.Equal(t, conns, 1)
}

func TestNode_DirectPeers(t *testing.T) {
	// This test checks whether the direct connection between peers configured
	// using the DirectPeers option is always maintained.

	peers, err := getNodeInfo(5)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	n0, err := NewNode(
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		ConnectionLimit(1, 1, 0),
		DirectPeers(peers[1].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n0.Start(ctx))
	time.Sleep(time.Second)

	n1, err := NewNode(
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
		DirectPeers(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start(ctx))
	time.Sleep(time.Second)

	for i := 2; i < len(peers); i++ {
		n, err := NewNode(
			PeerPrivKey(peers[i].PrivKey),
			ListenAddrs(peers[i].ListenAddrs),
		)
		require.NoError(t, err)
		require.NoError(t, n.Start(ctx))

		require.NoError(t, n.Connect(peers[0].PeerAddrs[0]))
		// Connection with tagged hosts are less likely to be dropped.
		// By tagging them we can be sure it wasn't coincidence that
		// the connection to n1 host is maintained after call to
		// the TrimOpenConns method.
		n0.Host().ConnManager().TagPeer(n.Host().ID(), "test", 1)
	}

	// The connection between n0 and n1 nodes should be persisted even
	// with a connection limit:
	n0.Host().ConnManager().TrimOpenConns(context.Background())
	time.Sleep(time.Second)
	assert.Equal(t, network.Connected, n0.Host().Network().Connectedness(n1.Host().ID()))
}
