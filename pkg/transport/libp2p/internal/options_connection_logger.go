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
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

// ConnectionLogger logs connected and disconnected hosts,
func ConnectionLogger() Options {
	return func(n *Node) error {
		n.AddNotifee(&connectionLoggerNotifee{n: n})
		return nil
	}
}

type connectionLoggerNotifee struct {
	n *Node
}

// Listen implements the network.Notifiee interface.
func (n *connectionLoggerNotifee) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose implements the network.Notifiee interface.
func (n *connectionLoggerNotifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected implements the network.Notifiee interface.
func (n *connectionLoggerNotifee) Connected(_ network.Network, conn network.Conn) {
	n.n.tsLog.get().
		WithFields(log.Fields{
			"peerID": conn.RemotePeer().String(),
			"addr":   conn.RemoteMultiaddr().String(),
		}).
		Info("Connected to a host")
}

// Disconnected implements the network.Notifiee interface.
func (n *connectionLoggerNotifee) Disconnected(_ network.Network, conn network.Conn) {
	n.n.tsLog.get().
		WithFields(log.Fields{
			"peerID": conn.RemotePeer().String(),
			"addr":   conn.RemoteMultiaddr().String(),
		}).
		Info("Disconnected from a host")
}

// OpenedStream implements the network.Notifiee interface.
func (n *connectionLoggerNotifee) OpenedStream(network.Network, network.Stream) {}

// ClosedStream implements the network.Notifiee interface.
func (n *connectionLoggerNotifee) ClosedStream(network.Network, network.Stream) {}
