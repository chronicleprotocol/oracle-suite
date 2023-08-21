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
	// Model of the data point.
	Model string `json:"model"`

	// Value is a binary representation of the data point.
	Value datapoint.Point `json:"value"`

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
	value, err := d.Value.MarshalBinary()
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

func (d *DataPoint) LogFields() log.Fields {
	if d == nil {
		return nil
	}
	f := log.Fields{
		"model":     d.Model,
		"signature": d.Signature.String(),
	}
	for k, v := range d.Value.LogFields() {
		f[k] = v
	}
	return f
}
