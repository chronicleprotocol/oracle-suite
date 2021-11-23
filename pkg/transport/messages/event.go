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
	Date       time.Time
	Type       string
	Data       []byte
	Signatures map[string][]byte
}

func (e *Event) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.Event{
		Timestamp:  e.Date.Unix(),
		Type:       e.Type,
		Data:       e.Data,
		Signatures: e.Signatures,
	})
}

func (e *Event) UnmarshallBinary(data []byte) error {
	msg := &pb.Event{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return err
	}
	e.Date = time.Unix(msg.Timestamp, 0)
	e.Type = msg.Type
	e.Data = msg.Data
	e.Signatures = msg.Signatures
	return nil
}
