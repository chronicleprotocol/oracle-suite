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

// ReceivedMessage contains a Message received from Transport with
// an additional data.
type ReceivedMessage struct {
	// Message contains the message content. It is nil when the Error field
	// is not nil.
	Message Message
	// Data contains an optional data associated with the message. A type of
	// the data is different depending on a transport implementation.
	Data interface{}
	// Error contains an optional error returned by transport.
	Error error
}

type Message interface {
	MarshallBinary() ([]byte, error)
	UnmarshallBinary([]byte) error
}

// Transport is the interface for different implementations of a
// publish–subscribe messaging solutions for the Oracle network.
type Transport interface {
	// Broadcast sends a message with a given topic.
	Broadcast(topic string, message Message) error
	// Messages returns a channel that will deliver incoming messages.
	// In case of an error, error will be returned in a ReceivedMessage
	// structure.
	Messages(topic string) chan ReceivedMessage
	// Start starts listening for messages.
	Start() error
	// Wait return channel which send a nil or error when transport's context
	// is cancelled.
	Wait() chan error
}
