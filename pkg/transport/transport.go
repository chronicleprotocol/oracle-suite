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

package transport

import (
	"encoding/json"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

// Message is a message that can be sent over transport.
type Message interface {
	// MarshallBinary serializes the message into a byte slice.
	MarshallBinary() ([]byte, error)

	// UnmarshallBinary deserializes the message from a byte slice.
	UnmarshallBinary([]byte) error
}

type WithAppInfo interface {
	Message
	SetAppInfo(info AppInfo)
	GetAppInfo() AppInfo
}

// Service implements a mechanism for exchanging messages between Oracles.
type Service interface {
	supervisor.Service
	Transport
}

type Transport interface {
	// Broadcast sends a message with a given topic.
	Broadcast(topic string, message Message) error

	// Messages returns a channel for incoming messages. A new channel is
	// created for each call, therefore this method should not be used in
	// loops. In case of an error, an error will be returned in the
	// ReceivedMessage structure.
	Messages(topic string) <-chan ReceivedMessage
}

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

	// Meta contains optional information about the message.
	Meta Meta
}

func (p ReceivedMessage) MarshalJSON() ([]byte, error) {
	type msg struct {
		Message Message `json:"message,omitempty"`
		Author  string  `json:"author,omitempty"`
		Error   error   `json:"error,omitempty"`
		Meta    Meta    `json:"meta,omitempty"`
	}
	return json.Marshal(msg{
		Message: p.Message,
		Author:  fmt.Sprintf("%x", p.Author),
		Error:   p.Error,
		Meta:    p.Meta,
	})
}

type AppInfo struct {
	Name    string `json:"-"`
	Version string `json:"-"`
}

func (a *AppInfo) SetAppInfo(info AppInfo) {
	a.Name = info.Name
	a.Version = info.Version
}

func (a *AppInfo) GetAppInfo() AppInfo {
	return *a
}

type Meta struct {
	Transport            string `json:"transport"`
	Topic                string `json:"topic"`
	MessageID            string `json:"message_id"`
	PeerID               string `json:"peer_id"`
	PeerAddr             string `json:"peer_addr"`
	ReceivedFromPeerID   string `json:"received_from_peer_id"`
	ReceivedFromPeerAddr string `json:"received_from_peer_addr"`
	UserAgent            string `json:"user_agent"`
}

func ReceivedMessageFields(p ReceivedMessage) log.Fields {
	c := p.Meta.Transport
	if p.Meta.Topic != "" {
		c += ":" + p.Meta.Topic
	}
	return log.Fields{
		"channel":              c,
		"messageID":            p.Meta.MessageID,
		"peerID":               p.Meta.PeerID,
		"peerAddr":             p.Meta.PeerAddr,
		"receivedFromPeerID":   p.Meta.ReceivedFromPeerID,
		"receivedFromPeerAddr": p.Meta.ReceivedFromPeerAddr,
		"userAgent":            p.Meta.UserAgent,
	}
}
