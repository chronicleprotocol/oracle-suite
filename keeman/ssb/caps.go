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

package ssb

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
	"go.cryptoscope.co/ssb"
	refs "go.mindeco.de/ssb-refs"

	"github.com/makerdao/oracle-suite/keeman/rand"
)

type Caps struct {
	Shs  []byte
	Sign []byte
}

func (c Caps) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Shs    string `json:"shs"`
		Sign   string `json:"sign,omitempty"`
		Invite string `json:"invite,omitempty"`
	}{
		Shs:  base64.URLEncoding.EncodeToString(c.Shs),
		Sign: base64.URLEncoding.EncodeToString(c.Sign),
	})
}

func NewCaps(privateKey *ecdsa.PrivateKey) (*Caps, error) {
	randBytes, err := rand.SeededRandBytesGen(crypto.FromECDSA(privateKey), 32)
	if err != nil {
		return nil, err
	}
	return &Caps{
		Shs:  randBytes(),
		Sign: randBytes(),
	}, nil
}

func NewKeyPair(privateKey *ecdsa.PrivateKey) (ssb.KeyPair, error) {
	return ssb.NewKeyPair(
		bytes.NewReader(crypto.FromECDSA(privateKey)),
		refs.RefAlgoFeedSSB1,
	)
}
