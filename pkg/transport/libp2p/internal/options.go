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

	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/internal/sets"
)

type Options func(n *Node) error

// DialTimeout sets dial timeout for libp2p nodes.
func DialTimeout(t time.Duration) Options {
	return func(n *Node) error {
		n.hostOpts = append(n.hostOpts, libp2p.WithDialTimeout(t))
		return nil
	}
}

// ListenAddrs configures node to listen on the given addresses.
func ListenAddrs(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		n.hostOpts = append(n.hostOpts, libp2p.ListenAddrs(addrs...))
		return nil
	}
}

// PeerPrivKey configures node to use given key as its identity.
func PeerPrivKey(sk crypto.PrivKey) Options {
	return func(n *Node) error {
		n.hostOpts = append(n.hostOpts, libp2p.Identity(sk))
		return nil
	}
}

// MessagePrivKey configures node to use given key to sign messages.
func MessagePrivKey(sk crypto.PrivKey) Options {
	return func(n *Node) error {
		pid, err := peer.IDFromPublicKey(sk.GetPublic())
		if err != nil {
			return err
		}

		// It's necessary to add this key to the peerstore,
		// otherwise it'll be impossible to use it to sign messages:
		err = n.Peerstore().AddPrivKey(pid, sk)
		if err != nil {
			return err
		}

		n.pubsubOpts = append(n.pubsubOpts, pubsub.WithMessageAuthor(pid))
		return nil
	}
}

// DisablePubSub disables the pubsub system.
func DisablePubSub() Options {
	return func(n *Node) error {
		n.disablePubSub = true
		return nil
	}
}

// UserAgent sets the libp2p user-agent sent along with the identify protocol.
func UserAgent(userAgent string) Options {
	return func(n *Node) error {
		n.hostOpts = append(n.hostOpts, libp2p.UserAgent(userAgent))
		return nil
	}
}

// ConnectionLimit limits the number of connections. When the number of
// connections reaches a high value, then as many connections will be
// dropped until we reach a low number of connections.
func ConnectionLimit(low, high int, grace time.Duration) Options {
	return func(n *Node) error {
		var err error
		n.connmgr, err = connmgr.NewConnManager(low, high, connmgr.WithGracePeriod(grace))
		return err
	}
}

// DirectPeers enforces direct connection with given peers. Note that the
// direct connection should be symmetrically configured at both ends.
func DirectPeers(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		if len(addrs) == 0 {
			return nil
		}
		n.tsLog.get().
			WithField("addrs", addrs).
			Info("Adding direct peers")
		var addrInfos []peer.AddrInfo
		for _, maddr := range addrs {
			ai, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				return err
			}
			addrInfos = append(addrInfos, *ai)
		}
		n.pubsubOpts = append(
			n.pubsubOpts,
			pubsub.WithDirectPeers(addrInfos),
		)
		connect := func() {
			for _, addrInfo := range addrInfos {
				if n.host.Network().Connectedness(addrInfo.ID) != network.NotConnected {
					continue
				}
				n.tsLog.get().
					WithField("peerID", addrInfo.ID.Pretty()).
					WithField("addrs", addrInfo.Addrs).
					Info("Connecting to the direct peer")
				err := n.host.Connect(n.ctx, addrInfo)
				if err != nil {
					n.tsLog.get().
						WithField("peerID", addrInfo.ID.Pretty()).
						WithField("addrs", addrInfo.Addrs).
						WithError(err).
						Warn("Unable to connect to the direct peer")
				}
			}
		}
		connectRoutine := func() {
			t := time.NewTicker(2 * time.Minute)
			connect()
			for {
				select {
				case <-n.ctx.Done():
					t.Stop()
					return
				case <-t.C:
					connect()
				}
			}
		}
		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event interface{}) {
			if _, ok := event.(sets.NodeStartedEvent); ok {
				go connectRoutine()
			}
		}))
		return nil
	}
}

// PubsubEventTracer provides a tracer for the pubsub system.
func PubsubEventTracer(tracer pubsub.EventTracer) Options {
	return func(n *Node) error {
		n.pubsubOpts = append(n.pubsubOpts, pubsub.WithEventTracer(tracer))
		return nil
	}
}

// Discovery configures node to use kad-dht for node discovery.
func Discovery(bootstrapAddrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		var err error
		var kadDHT *dht.IpfsDHT

		addrs, err := peer.AddrInfosFromP2pAddrs(bootstrapAddrs...)
		if err != nil {
			return err
		}

		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event interface{}) {
			switch event.(type) {
			case sets.NodeHostStartedEvent:
				n.tsLog.get().
					WithField("bootstrapAddrs", bootstrapAddrs).
					Info("Starting KAD-DHT discovery")
				for _, addr := range addrs {
					// bootstrap nodes aren't protected by KAD-DHT so we have
					// do it manually
					n.connmgr.Protect(addr.ID, "bootstrap")
					n.peerstore.AddAddrs(addr.ID, addr.Addrs, peerstore.PermanentAddrTTL)
				}
				kadDHT, err = dht.New(n.ctx, n.host, dht.BootstrapPeers(addrs...), dht.Mode(dht.ModeServer))
				if err != nil {
					n.tsLog.get().
						WithError(err).
						Error("Unable to initialize KAD-DHT")
					return
				}
				if err = kadDHT.Bootstrap(n.ctx); err != nil {
					n.tsLog.get().
						WithError(err).
						Error("Unable to bootstrap KAD-DHT")
					return
				}
				n.pubsubOpts = append(n.pubsubOpts, pubsub.WithDiscovery(discovery.NewRoutingDiscovery(kadDHT)))
			case sets.NodeStoppingEvent:
				if kadDHT == nil {
					return
				}
				err = kadDHT.Close()
				if err != nil {
					n.tsLog.get().
						WithError(err).
						Error("Unable to close KAD-DHT")
				}
			}
		}))
		return nil
	}
}
