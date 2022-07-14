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

package sets

import (
	"sync"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// NotifeeSet implements the network.Notifiee and allow to aggregate
// multiple instances of this interface.
type NotifeeSet struct {
	mu sync.RWMutex

	notifees []network.Notifiee
}

// NewNotifeeSet creates new instance of the NotifeeSet.
func NewNotifeeSet() *NotifeeSet {
	return &NotifeeSet{}
}

// Add adds new network.Notifiee to the set.
func (n *NotifeeSet) Add(notifees ...network.Notifiee) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.notifees = append(n.notifees, notifees...)
}

// Remove removes network.Notifiee from the set if already added.
func (n *NotifeeSet) Remove(notifees ...network.Notifiee) {
	n.mu.Lock()
	defer n.mu.Unlock()

	var notifeesDiff []network.Notifiee
	for _, a := range n.notifees {
		f := false
		for _, b := range notifees {
			if a == b {
				f = true
				break
			}
		}
		if !f {
			notifeesDiff = append(notifeesDiff, a)
		}
	}
	n.notifees = notifeesDiff
}

// Listen implements the network.Notifiee interface.
func (n *NotifeeSet) Listen(network network.Network, maddr multiaddr.Multiaddr) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, notifee := range n.notifees {
		notifee.Listen(network, maddr)
	}
}

// ListenClose implements the network.Notifiee interface.
func (n *NotifeeSet) ListenClose(network network.Network, maddr multiaddr.Multiaddr) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, notifee := range n.notifees {
		notifee.ListenClose(network, maddr)
	}
}

// Connected implements the network.Notifiee interface.
func (n *NotifeeSet) Connected(network network.Network, conn network.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, notifee := range n.notifees {
		notifee.Connected(network, conn)
	}
}

// Disconnected implements the network.Notifiee interface.
func (n *NotifeeSet) Disconnected(network network.Network, conn network.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, notifee := range n.notifees {
		notifee.Disconnected(network, conn)
	}
}

// OpenedStream implements the network.Notifiee interface.
func (n *NotifeeSet) OpenedStream(network network.Network, stream network.Stream) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, notifee := range n.notifees {
		notifee.OpenedStream(network, stream)
	}
}

// ClosedStream implements the network.Notifiee interface.
func (n *NotifeeSet) ClosedStream(network network.Network, stream network.Stream) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, notifee := range n.notifees {
		notifee.ClosedStream(network, stream)
	}
}

var _ network.Notifiee = (*NotifeeSet)(nil)
