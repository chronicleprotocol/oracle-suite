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

package teleportevm

import (
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const SignatureKey = "ethereum"

// Signer signs Ethereum logger messages using Ethereum signature.
type Signer struct {
	signer ethereum.Signer
	types  []string
}

// NewSigner returns a new instance of the Signer struct.
func NewSigner(signer ethereum.Signer, types []string) *Signer {
	return &Signer{signer: signer, types: types}
}

// Sign implements the Signer interface.
func (l *Signer) Sign(event *messages.Event) (bool, error) {
	supports := false
	for _, t := range l.types {
		if t == event.Type {
			supports = true
			break
		}
	}
	if !supports {
		return false, nil
	}
	h, ok := event.Data["hash"]
	if !ok {
		return false, errors.New("missing hash field")
	}
	s, err := l.signer.Signature(h)
	if err != nil {
		return false, err
	}
	if event.Signatures == nil {
		event.Signatures = map[string]messages.EventSignature{}
	}
	event.Signatures[SignatureKey] = messages.EventSignature{
		Signer:    l.signer.Address().Bytes(),
		Signature: s.Bytes(),
	}
	return true, nil
}
