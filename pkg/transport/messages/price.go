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
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages/pb"
)

const PriceMessageName = "price/v0"
const PriceV1MessageName = "price/v1"

var ErrPriceMalformedMessage = errors.New("malformed price message")
var ErrUnknownMessageVersion = errors.New("unknown message version")

type Price struct {
	// MessageVersion is the version of the message. The value 0 corresponds to
	// the price/v0 and 1 to the price/v1 message. Both messages contain the
	// data but the price/v1 uses protobuf to encode the data. After full
	// migration to the price/v1 message, the price/v0 must be removed
	// along with this field.
	MessageVersion uint8 `json:"-"`

	Price   *oracle.Price   `json:"price"`
	Trace   json.RawMessage `json:"trace"`
	Version string          `json:"version,omitempty"`
}

func (p *Price) Marshall() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Price) Unmarshall(b []byte) error {
	err := json.Unmarshal(b, p)
	if err != nil {
		return err
	}
	if p.Price == nil {
		return ErrPriceMalformedMessage
	}
	return nil
}

// MarshallBinary implements the transport.Message interface.
func (p *Price) MarshallBinary() ([]byte, error) {
	switch p.MessageVersion {
	case 1:
		data, err := proto.Marshal(&pb.Price{
			Wat:     p.Price.Wat,
			Val:     p.Price.Val.Bytes(),
			Age:     p.Price.Age.Unix(),
			Vrs:     ethereum.SignatureFromVRS(p.Price.V, p.Price.R, p.Price.S).Bytes(),
			StarkR:  p.Price.StarkR,
			StarkS:  p.Price.StarkS,
			StarkPK: p.Price.StarkPK,
			Trace:   p.Trace,
			Version: p.Version,
		})
		if err != nil {
			return nil, err
		}
		return data, nil
	case 0:
		return p.Marshall()
	}
	return nil, ErrUnknownMessageVersion
}

// UnmarshallBinary implements the transport.Message interface.
func (p *Price) UnmarshallBinary(data []byte) error {
	switch p.MessageVersion {
	case 1:
		msg := &pb.Price{}
		if err := proto.Unmarshal(data, msg); err != nil {
			return err
		}
		v, r, s := ethereum.SignatureFromBytes(msg.Vrs).VRS()
		p.Price = &oracle.Price{
			Wat:     msg.Wat,
			Val:     new(big.Int).SetBytes(msg.Val),
			Age:     time.Unix(msg.Age, 0),
			V:       v,
			R:       r,
			S:       s,
			StarkR:  msg.StarkR,
			StarkS:  msg.StarkS,
			StarkPK: msg.StarkPK,
		}
		p.Trace = msg.Trace
		p.Version = msg.Version
	case 0:
		return p.Unmarshall(data)
	}
	return ErrUnknownMessageVersion
}

func (p *Price) AsV0() *Price {
	c := p.copy()
	c.MessageVersion = 0
	return c
}

func (p *Price) AsV1() *Price {
	c := p.copy()
	c.MessageVersion = 1
	return c
}

func (p *Price) copy() *Price {
	c := &Price{
		MessageVersion: p.MessageVersion,
		Price: &oracle.Price{
			Wat:     p.Price.Wat,
			Val:     new(big.Int).Set(p.Price.Val),
			Age:     p.Price.Age,
			V:       p.Price.V,
			R:       p.Price.R,
			S:       p.Price.S,
			StarkR:  p.Price.StarkR,
			StarkS:  p.Price.StarkS,
			StarkPK: p.Price.StarkPK,
		},
		Trace:   p.Trace,
		Version: p.Version,
	}
	c.Trace = make([]byte, len(p.Trace))
	c.Price.StarkS = make([]byte, len(p.Price.StarkS))
	c.Price.StarkR = make([]byte, len(p.Price.StarkR))
	c.Price.StarkPK = make([]byte, len(p.Price.StarkPK))
	copy(c.Trace, p.Trace)
	copy(c.Price.StarkS, p.Price.StarkS)
	copy(c.Price.StarkR, p.Price.StarkR)
	copy(c.Price.StarkPK, p.Price.StarkPK)
	return c
}
