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
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// PubSubEventHandlerFunc is a adapter for the PubSubEventHandler interface.
type PubSubEventHandlerFunc func(topic string, event pubsub.PeerEvent)

// Handle calls f(topic, event).
func (f PubSubEventHandlerFunc) Handle(topic string, event pubsub.PeerEvent) {
	f(topic, event)
}

// PubSubEventHandler can ba implemented by type that supports handling the PubSub
// system events.
type PubSubEventHandler interface {
	// Handle is called on a new event.
	Handle(topic string, event pubsub.PeerEvent)
}

// PubSubEventHandlerSet stores multiple instances of the PubSubEventHandler interface.
type PubSubEventHandlerSet struct {
	eventHandler []PubSubEventHandler
}

// NewPubSubEventHandlerSet creates new instance of the PubSubEventHandlerSet.
func NewPubSubEventHandlerSet() *PubSubEventHandlerSet {
	return &PubSubEventHandlerSet{}
}

// Add adds new PubSubEventHandler to the set.
func (n *PubSubEventHandlerSet) Add(eventHandler ...PubSubEventHandler) {
	n.eventHandler = append(n.eventHandler, eventHandler...)
}

// Handle invokes all registered handlers for given topic.
func (n *PubSubEventHandlerSet) Handle(topic string, event pubsub.PeerEvent) {
	for _, eventHandler := range n.eventHandler {
		eventHandler.Handle(topic, event)
	}
}

var _ PubSubEventHandler = (*PubSubEventHandlerSet)(nil)
