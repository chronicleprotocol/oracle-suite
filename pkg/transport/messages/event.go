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
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/makerdao/oracle-suite/pkg/transport/messages/pb"
)

var EventMessageName = "event/v0"

type Event struct {
	// The date when the event message was created. It is *not* the date of
	// the event itself.
	Date time.Time
	// Type of the event.
	Type string
	// Unique ID of the event.
	ID []byte
	// Specifies a group to which the event belongs to, like transaction hash.
	Group []byte
	// List of event data.
	Data map[string][]byte
	// List of event signatures.
	Signatures map[string][]byte
}

// MarshallBinary implements the transport.Message interface.
func (e *Event) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.Event{
		Timestamp:  e.Date.Unix(),
		Type:       e.Type,
		Id:         e.ID,
		Group:      e.Group,
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
	e.Group = msg.Group
	e.Data = msg.Data
	e.Signatures = msg.Signatures
	return nil
}
