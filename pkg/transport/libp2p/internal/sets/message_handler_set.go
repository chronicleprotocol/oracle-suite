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

// MessageHandler can ba implemented by type that supports handling the PubSub
// system messages.
type MessageHandler interface {
	// Published is called when new message is published.
	Published(topic string, msg []byte)
	// Received is called when new message is received.
	Received(topic string, msg *pubsub.Message, result pubsub.ValidationResult)
}

// MessageHandlerSet stores multiple instances of the MessageHandler interface.
type MessageHandlerSet struct {
	messageHandler []MessageHandler
}

// NewMessageHandlerSet creates new instance of the MessageHandlerSet.
func NewMessageHandlerSet() *MessageHandlerSet {
	return &MessageHandlerSet{}
}

// Add adds new MessageHandler to the set.
func (n *MessageHandlerSet) Add(messageHandler ...MessageHandler) {
	n.messageHandler = append(n.messageHandler, messageHandler...)
}

// Published invokes all registered handlers.
func (n *MessageHandlerSet) Published(topic string, msg []byte) {
	for _, messageHandler := range n.messageHandler {
		messageHandler.Published(topic, msg)
	}
}

// Received invokes all registered handlers.
func (n *MessageHandlerSet) Received(topic string, msg *pubsub.Message, result pubsub.ValidationResult) {
	for _, messageHandler := range n.messageHandler {
		messageHandler.Received(topic, msg, result)
	}
}

var _ MessageHandler = (*MessageHandlerSet)(nil)
