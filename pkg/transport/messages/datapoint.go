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

package messages

import (
	"encoding/json"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages/pb"

	"google.golang.org/protobuf/proto"
)

const DataPointV1MessageName = "data_point/v1"

type DataPoint struct {
	// Model is the name of the data model.
	Model string `json:"model"`

	// Value is a binary representation of the data point.
	Value datapoint.Point `json:"value"`

	// Signature is the feed signature of the data point.
	Signature types.Signature `json:"signature"`
}

func (d *DataPoint) Marshall() ([]byte, error) {
	return json.Marshal(d)
}

func (d *DataPoint) Unmarshall(b []byte) error {
	err := json.Unmarshal(b, d)
	if err != nil {
		return err
	}
	return nil
}

// MarshallBinary implements the transport.Message interface.
func (d *DataPoint) MarshallBinary() ([]byte, error) {
	// Copy of the data point without the trace to reduce the size of the
	// message. In the future, this problem should be solved by using a
	// compression algorithm.
	cpy := datapoint.Point{
		Value:     d.Value.Value,
		Time:      d.Value.Time,
		Meta:      d.Value.Meta,
		Error:     d.Value.Error,
		SubPoints: d.Value.SubPoints,
	}
	value, err := cpy.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&pb.DataPointMessage{
		Model:     d.Model,
		Value:     value,
		Signature: d.Signature.Bytes(),
	})
}

// UnmarshallBinary implements the transport.Message interface.
func (d *DataPoint) UnmarshallBinary(data []byte) error {
	msg := &pb.DataPointMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return err
	}
	err := d.Value.UnmarshalBinary(msg.Value)
	if err != nil {
		return err
	}
	sig, err := types.SignatureFromBytes(msg.Signature)
	if err != nil {
		return err
	}
	d.Model = msg.Model
	d.Signature = sig
	return nil
}

func DataPointMessageLogFields(d DataPoint) log.Fields {
	f := log.Fields{
		"point.model":     d.Model,
		"point.signature": d.Signature.String(),
	}
	for k, v := range datapoint.PointLogFields(d.Value) {
		f[k] = v
	}
	return f
}
