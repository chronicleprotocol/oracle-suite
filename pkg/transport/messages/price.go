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

	"github.com/makerdao/oracle-suite/pkg/oracle"
)

var PriceMessageName = "price/v0"

var ErrPriceMalformedMessage = errors.New("malformed price message")

type Price struct {
	Price *oracle.Price   `json:"price"`
	Trace json.RawMessage `json:"trace"`
}

func (p *Price) Marshall() ([]byte, error) {
	// TODO: remove
	return json.Marshal(p)
}

func (p *Price) Unmarshall(b []byte) error {
	// TODO: remove
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
	return p.Marshall()
}

// UnmarshallBinary implements the transport.Message interface.
func (p *Price) UnmarshallBinary(data []byte) error {
	return p.Unmarshall(data)
}
