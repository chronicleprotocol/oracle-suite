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

type NodeConfiguredEvent struct{}
type NodeStartingEvent struct{}
type NodeHostStartedEvent struct{}
type NodePubSubStartedEvent struct{}
type NodeStartedEvent struct{}
type NodeTopicSubscribedEvent struct{ Topic string }
type NodeTopicUnsubscribedEvent struct{ Topic string }
type NodeStoppingEvent struct{}
type NodeStoppedEvent struct{}

// NodeEventHandlerFunc is a adapter for the NodeEventHandler interface.
type NodeEventHandlerFunc func(event interface{})

// Handle calls f(topic, event).
func (f NodeEventHandlerFunc) Handle(event interface{}) {
	f(event)
}

// NodeEventHandler can ba implemented by type that supports handling the Node
// system events.
type NodeEventHandler interface {
	// Handle is called on a new event.
	Handle(event interface{})
}

// NodeEventHandlerSet stores multiple instances of the NodeEventHandler interface.
type NodeEventHandlerSet struct {
	eventHandler []NodeEventHandler
}

// NewNodeEventHandlerSet creates new instance of the NodeEventHandlerSet.
func NewNodeEventHandlerSet() *NodeEventHandlerSet {
	return &NodeEventHandlerSet{}
}

// Add adds new NodeEventHandler to the set.
func (n *NodeEventHandlerSet) Add(eventHandler ...NodeEventHandler) {
	n.eventHandler = append(n.eventHandler, eventHandler...)
}

// Handle invokes all registered handlers for given topic.
func (n *NodeEventHandlerSet) Handle(event interface{}) {
	for _, eventHandler := range n.eventHandler {
		eventHandler.Handle(event)
	}
}

var _ NodeEventHandler = (*NodeEventHandlerSet)(nil)
