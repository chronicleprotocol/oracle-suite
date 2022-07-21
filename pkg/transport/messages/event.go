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

package messages

import (
	"errors"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages/pb"
)

const EventV1MessageName = "event/v1"

const eventMessageMaxSize = 1 * 1024 * 1024 // 1MB

var ErrEventMessageTooLarge = errors.New("event message too large")

type EventSignature struct {
	Signer    []byte
	Signature []byte
}

type Event struct {
	// Type of the event.
	Type string
	// Unique ID of the event.
	ID []byte
	// Event index used to search for events.
	Index []byte
	// The date of the event.
	EventDate time.Time
	// The date when the event message was created.
	MessageDate time.Time
	// List of event data.
	Data map[string][]byte
	// List of event signatures.
	Signatures map[string]EventSignature
}

// MarshallBinary implements the transport.Message interface.
func (e *Event) MarshallBinary() ([]byte, error) {
	signatures := map[string]*pb.Event_Signature{}
	for k, s := range e.Signatures {
		signatures[k] = &pb.Event_Signature{
			Signer:    s.Signer,
			Signature: s.Signature,
		}
	}
	data, err := proto.Marshal(&pb.Event{
		Type:             e.Type,
		Id:               e.ID,
		Index:            e.Index,
		EventTimestamp:   e.EventDate.Unix(),
		MessageTimestamp: e.MessageDate.Unix(),
		Data:             e.Data,
		Signatures:       signatures,
	})
	if err != nil {
		return nil, err
	}
	if len(data) > eventMessageMaxSize {
		return nil, ErrEventMessageTooLarge
	}
	return data, nil
}

// UnmarshallBinary implements the transport.Message interface.
func (e *Event) UnmarshallBinary(data []byte) error {
	if len(data) > eventMessageMaxSize {
		return ErrEventMessageTooLarge
	}
	msg := &pb.Event{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return err
	}
	signatures := map[string]EventSignature{}
	for k, s := range msg.Signatures {
		signatures[k] = EventSignature{
			Signer:    s.Signer,
			Signature: s.Signature,
		}
	}
	e.Type = msg.Type
	e.ID = msg.Id
	e.Index = msg.Index
	e.EventDate = time.Unix(msg.EventTimestamp, 0)
	e.MessageDate = time.Unix(msg.MessageTimestamp, 0)
	e.Data = msg.Data
	e.Signatures = signatures
	return nil
}
