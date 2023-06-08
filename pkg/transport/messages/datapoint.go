package messages

import (
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages/pb"

	"google.golang.org/protobuf/proto"
)

const DataPointV1MessageName = "data_point/v1"

type DataPoint struct {
	// Model of the data point.
	Model string

	// Value is a binary representation of the data point.
	Value datapoint.Point

	Signature types.Signature
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
