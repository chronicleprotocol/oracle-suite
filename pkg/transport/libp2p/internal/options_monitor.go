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

package internal

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/internal/sets"
)

type MonitorConfig struct {
	// ShowLogOnChange enables logging immediately when the number of peers
	// or connections changes.
	ShowLogOnChange bool

	// ShowLogInterval specifies the interval between logging.
	// When ShowLogOnChange is true, then the delay resets after each change.
	ShowLogInterval time.Duration
}

type monitorNotifee struct {
	mu sync.RWMutex

	stopped   bool
	notifeeCh chan struct{}
}

// Listen implements the network.Notifiee interface.
func (n *monitorNotifee) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose implements the network.Notifiee interface.
func (n *monitorNotifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected implements the network.Notifiee interface.
func (n *monitorNotifee) Connected(_ network.Network, _ network.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if !n.stopped {
		n.notifeeCh <- struct{}{}
	}
}

// Disconnected implements the network.Notifiee interface.
func (n *monitorNotifee) Disconnected(_ network.Network, _ network.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.stopped {
		n.notifeeCh <- struct{}{}
	}
}

// Stop stops monitoring notifee.
func (n *monitorNotifee) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.stopped = true
	close(n.notifeeCh)
}

func Monitor(cfg MonitorConfig) Options {
	return func(n *Node) error {
		printLog := func() {
			n.tsLog.get().
				WithFields(log.Fields{
					"peerCount": len(n.host.Network().Peers()),
					"connCount": len(n.host.Network().Conns()),
				}).
				Info("Connection monitor")
		}

		notifeeCh := make(chan struct{})
		notifee := &monitorNotifee{notifeeCh: notifeeCh}
		if cfg.ShowLogOnChange {
			n.AddNotifee(notifee)
		}

		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event interface{}) {
			if _, ok := event.(sets.NodeStartedEvent); ok {
				go func() {
					t := time.NewTicker(cfg.ShowLogInterval)
					for {
						select {
						case <-notifeeCh:
							printLog()
							t.Reset(cfg.ShowLogInterval)
						case <-t.C:
							printLog()
						case <-n.ctx.Done():
							notifee.Stop()
							t.Stop()
							n.RemoveNotifee(notifee)
							return
						}
					}
				}()
			}
		}))
		return nil
	}
}
