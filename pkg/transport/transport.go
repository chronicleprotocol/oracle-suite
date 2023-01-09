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

package transport

import (
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

// ReceivedMessage contains a Message received from Transport.
type ReceivedMessage struct {
	// Message contains the message content. It is nil when the Error field
	// is not nil.
	Message Message

	// Author is the author of the message.
	Author []byte

	// Data contains an optional data associated with the message. A type of
	// the data is different depending on a transport implementation.
	Data any

	// Error contains an optional error returned by transport layer.
	Error error
}

// Message is a message that can be sent over transport.
type Message interface {
	// MarshallBinary serializes the message into a byte slice.
	MarshallBinary() ([]byte, error)

	// UnmarshallBinary deserializes the message from a byte slice.
	UnmarshallBinary([]byte) error
}

// Transport implements a mechanism for exchanging messages between Oracles.
type Transport interface {
	supervisor.Service

	// Broadcast sends a message with a given topic.
	Broadcast(topic string, message Message) error

	// Messages returns a channel for incoming messages. A new channel is
	// created for each call, therefore this method should not be used in
	// loops. In case of an error, an error will be returned in the
	// ReceivedMessage structure.
	Messages(topic string) <-chan ReceivedMessage
}
