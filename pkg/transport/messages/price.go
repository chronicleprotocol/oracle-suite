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
	"errors"
	"math/big"
	"time"

	"github.com/defiweb/go-eth/types"
	"google.golang.org/protobuf/proto"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const PriceV0MessageName = "price/v0"
const PriceV1MessageName = "price/v1"

const priceMessageMaxSize = 1 * 1024 * 1024 // 1MB

var (
	ErrPriceMessageTooLarge       = errors.New("price message too large")
	ErrUnknownPriceMessageVersion = errors.New("unknown message version")
	ErrInvalidPriceMessage        = errors.New("invalid price message")
)

type Price struct {
	Price   *median.Price   `json:"price"`
	Trace   json.RawMessage `json:"trace"`
	Version string          `json:"version,omitempty"` // TODO: this should move to some meta field e.g. `feedVersion`

	// messageVersion is the version of the message. The value 0 corresponds to
	// the price/v0 and 1 to the price/v1 message. Both messages contain the
	// same data but the price/v1 uses protobuf to encode the data. After full
	// migration to the price/v1 message, the price/v0 must be removed
	// along with this field.
	messageVersion uint8
}

func (p *Price) Marshall() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Price) Unmarshall(b []byte) error {
	err := json.Unmarshal(b, p)
	if err != nil {
		return err
	}
	return nil
}

// MarshallBinary implements the transport.Message interface.
func (p *Price) MarshallBinary() ([]byte, error) {
	switch p.messageVersion {
	case 1:
		pbPrice := &pb.Price{
			Wat:     p.Price.Wat,
			Age:     p.Price.Age.Unix(),
			Vrs:     p.Price.Sig.Bytes(),
			Trace:   p.Trace,
			Version: p.Version,
		}
		if p.Price.Val != nil {
			pbPrice.Val = p.Price.Val.Bytes()
		}
		data, err := proto.Marshal(pbPrice)
		if err != nil {
			return nil, err
		}
		if len(data) > priceMessageMaxSize {
			return nil, ErrEventMessageTooLarge
		}
		return data, nil
	case 0:
		data, err := p.Marshall()
		if err != nil {
			return nil, err
		}
		if len(data) > priceMessageMaxSize {
			return nil, ErrPriceMessageTooLarge
		}
		return data, nil
	}
	return nil, ErrUnknownPriceMessageVersion
}

// UnmarshallBinary implements the transport.Message interface.
func (p *Price) UnmarshallBinary(data []byte) error {
	if len(data) > priceMessageMaxSize {
		return ErrPriceMessageTooLarge
	}
	switch json.Valid(data) {
	case true:
		p.messageVersion = 0
	case false:
		p.messageVersion = 1
	}
	switch p.messageVersion {
	case 1:
		msg := &pb.Price{}
		if err := proto.Unmarshal(data, msg); err != nil {
			return err
		}
		sig, err := types.SignatureFromBytes(msg.Vrs)
		if err != nil {
			return err
		}
		p.Price = &median.Price{
			Wat: msg.Wat,
			Val: new(big.Int).SetBytes(msg.Val),
			Age: time.Unix(msg.Age, 0),
			Sig: sig,
		}
		p.Trace = msg.Trace
		p.Version = msg.Version
	case 0:
		if err := p.Unmarshall(data); err != nil {
			return err
		}
	default:
		return ErrUnknownPriceMessageVersion
	}
	if p.Price == nil {
		return ErrInvalidPriceMessage
	}
	if p.Price.Val == nil {
		p.Price.Val = big.NewInt(0)
	}
	return nil
}

func (p *Price) AsV0() *Price {
	c := p.copy()
	c.messageVersion = 0
	return c
}

func (p *Price) AsV1() *Price {
	c := p.copy()
	c.messageVersion = 1
	return c
}

func (p *Price) copy() *Price {
	c := &Price{
		messageVersion: p.messageVersion,
		Price: &median.Price{
			Wat: p.Price.Wat,
			Age: p.Price.Age,
			Sig: p.Price.Sig,
		},
		Trace:   p.Trace,
		Version: p.Version,
	}
	if p.Price.Val != nil {
		c.Price.Val = new(big.Int).Set(p.Price.Val)
	}
	if p.Trace != nil {
		c.Trace = make([]byte, len(p.Trace))
		copy(c.Trace, p.Trace)
	}
	return c
}

func Price2DataPoint(msg *Price) *DataPoint {
	return &DataPoint{
		Model: msg.Price.Wat,
		Value: datapoint.Point{
			Value: value.Tick{
				Pair: value.Pair{
					Base:  msg.Price.Wat[:len(msg.Price.Wat)-3],
					Quote: msg.Price.Wat[len(msg.Price.Wat)-3:],
				},
				Price: bn.Float(msg.Price.Val),
			},
			Time: msg.Price.Age,
		},
		Signature: msg.Price.Sig,
	}
}
