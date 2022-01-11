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

const EventMessageName = "event/v0"

const eventMessageMaxDataFields = 10
const eventMessageMaxSignatureFields = 10
const eventMessageMaxKeyLen = 32
const eventMessageMaxFieldSize = 1024

type Event struct {
	// The date when the event message was created. It is *not* the date of
	// the event itself.
	Date time.Time
	// Type of the event.
	Type string
	// Unique ID of the event.
	ID []byte
	// Event index used to search for events.
	Index []byte
	// List of event data.
	Data map[string][]byte
	// List of event signatures.
	Signatures map[string][]byte
}

// MarshallBinary implements the transport.Message interface.
func (e *Event) MarshallBinary() ([]byte, error) {
	if err := e.validate(); err != nil {
		return nil, err
	}
	return proto.Marshal(&pb.Event{
		Timestamp:  e.Date.Unix(),
		Type:       e.Type,
		Id:         e.ID,
		Index:      e.Index,
		Data:       e.Data,
		Signatures: e.Signatures,
	})
}

// UnmarshallBinary implements the transport.Message interface.
func (e *Event) UnmarshallBinary(data []byte) error {
	msg := &pb.Event{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return err
	}
	e.Date = time.Unix(msg.Timestamp, 0)
	e.Type = msg.Type
	e.ID = msg.Id
	e.Index = msg.Index
	e.Data = msg.Data
	e.Signatures = msg.Signatures
	return e.validate()
}

func (e *Event) validate() error {
	if len(e.Type) > eventMessageMaxFieldSize {
		return errors.New("invalid event message, type length too large")
	}
	if len(e.ID) > eventMessageMaxFieldSize {
		return errors.New("invalid event message, ID size too large")
	}
	if len(e.Index) > eventMessageMaxFieldSize {
		return errors.New("invalid event message, index size too large")
	}
	if len(e.Data) > eventMessageMaxDataFields {
		return errors.New("invalid event message, too many data fields")
	}
	if len(e.Signatures) > eventMessageMaxSignatureFields {
		return errors.New("invalid event message, too many signatures")
	}
	for k, v := range e.Data {
		if len(k) > eventMessageMaxKeyLen {
			return errors.New("invalid event message, data key too long")
		}
		if len(v) > eventMessageMaxFieldSize {
			return errors.New("invalid event message, data size too large")
		}
	}
	for k, v := range e.Signatures {
		if len(k) > eventMessageMaxKeyLen {
			return errors.New("invalid event message, signature key too long")
		}
		if len(v) > eventMessageMaxFieldSize {
			return errors.New("invalid event message, signature size too large")
		}
	}
	return nil
}
