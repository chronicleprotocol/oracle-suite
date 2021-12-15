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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"go.cryptoscope.co/ssb"
	refs "go.mindeco.de/ssb-refs"

	"github.com/chronicleprotocol/oracle-suite/cmd/keeman/rand"
)

type Caps000 struct {
	Shs  []byte
	Sign []byte
}

func (c Caps000) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Shs    string `json:"shs"`
		Sign   string `json:"sign,omitempty"`
		Invite string `json:"invite,omitempty"`
	}{
		Shs:  base64.URLEncoding.EncodeToString(c.Shs),
		Sign: base64.URLEncoding.EncodeToString(c.Sign),
	})
}

func NewCaps(seed []byte) (*Caps000, error) {
	randBytes, err := rand.SeededRandBytesGen(seed, 32)
	if err != nil {
		return nil, err
	}
	return &Caps000{
		Shs:  randBytes(),
		Sign: randBytes(),
	}, nil
}

func NewKeyPair(b []byte) (ssb.KeyPair, error) {
	return ssb.NewKeyPair(
		bytes.NewReader(b),
		refs.RefAlgoFeedSSB1,
	)
}

type Caps struct {
	Shs    string `json:"shs"`
	Sign   string `json:"sign,omitempty"`
	Invite string `json:"invite,omitempty"`
}

func LoadCapsFromConfigFile(fileName string) (Caps, error) {
	b, err := LoadFile(fileName)
	if err != nil {
		return Caps{}, err
	}
	var c struct {
		Caps Caps `json:"caps"`
	}
	return c.Caps, json.Unmarshal(b, &c)
}

func LoadCapsFile(fileName string) (Caps, error) {
	b, err := LoadFile(fileName)
	if err != nil {
		return Caps{}, err
	}
	var c Caps
	return c, json.Unmarshal(b, &c)
}

func LoadFile(fileName string) (b []byte, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("could not open file %s: %w", fileName, err)
	}
	defer func() {
		err = f.Close()
	}()
	b, err = ioutil.ReadAll(f)
	return b, err
}
