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

package oracle

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const PriceMultiplier = 1e18

var ErrPriceNotSet = errors.New("unable to sign a price because the price is not set")
var ErrUnmarshallingFailure = errors.New("unable to unmarshal given JSON")

func errUnmarshalling(s string, err error) error {
	return fmt.Errorf("%w: %s: %s", ErrUnmarshallingFailure, s, err)
}

type Price struct {
	Wat string    // Wat is the asset name.
	Val *big.Int  // Val is the asset price multiplied by PriceMultiplier.
	Age time.Time // Age is the time when the price was obtained.

	// Signature:
	V byte
	R [32]byte
	S [32]byte

	// StarkWare signature:
	StarkR  []byte
	StarkS  []byte
	StarkPK []byte
}

// jsonPrice is the JSON representation of the Price structure.
type jsonPrice struct {
	Wat     string `json:"wat"`
	Val     string `json:"val"`
	Age     int64  `json:"age"`
	V       string `json:"v"`
	R       string `json:"r"`
	S       string `json:"s"`
	StarkR  string `json:"stark_r"`
	StarkS  string `json:"stark_s"`
	StarkPK string `json:"stark_pk"`
}

func (p *Price) SetFloat64Price(price float64) {
	pf := new(big.Float).SetFloat64(price)
	pf = new(big.Float).Mul(pf, new(big.Float).SetFloat64(PriceMultiplier))
	pi, _ := pf.Int(nil)

	p.Val = pi
}

func (p *Price) Float64Price() float64 {
	x := new(big.Float).SetInt(p.Val)
	x = new(big.Float).Quo(x, new(big.Float).SetFloat64(PriceMultiplier))
	f, _ := x.Float64()

	return f
}

func (p *Price) From(signer ethereum.Signer) (*ethereum.Address, error) {
	from, err := signer.Recover(p.Signature(), p.hash())
	if err != nil {
		return nil, err
	}

	return from, nil
}

func (p *Price) Sign(signer ethereum.Signer) error {
	if p.Val == nil {
		return ErrPriceNotSet
	}

	signature, err := signer.Signature(p.hash())
	if err != nil {
		return err
	}

	p.V, p.R, p.S = signature.VRS()

	return nil
}

func (p *Price) Signature() ethereum.Signature {
	return ethereum.SignatureFromVRS(p.V, p.R, p.S)
}

func (p *Price) Fields(signer ethereum.Signer) log.Fields {
	from := "*invalid signature*"
	if addr, err := p.From(signer); err == nil {
		from = addr.String()
	}

	return log.Fields{
		"from":    from,
		"wat":     p.Wat,
		"age":     p.Age.UTC().Format(time.RFC3339),
		"val":     p.Val.String(),
		"hash":    hex.EncodeToString(p.hash()),
		"V":       hex.EncodeToString([]byte{p.V}),
		"R":       hex.EncodeToString(p.R[:]),
		"S":       hex.EncodeToString(p.S[:]),
		"starkR":  encodeHexNumber(p.StarkR),
		"starkS":  encodeHexNumber(p.StarkS),
		"starkPK": encodeHexNumber(p.StarkPK),
	}
}

func (p *Price) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonPrice{
		Wat:     p.Wat,
		Val:     p.Val.String(),
		Age:     p.Age.Unix(),
		V:       hex.EncodeToString([]byte{p.V}),
		R:       hex.EncodeToString(p.R[:]),
		S:       hex.EncodeToString(p.S[:]),
		StarkR:  encodeHexNumber(p.StarkR),
		StarkS:  encodeHexNumber(p.StarkS),
		StarkPK: encodeHexNumber(p.StarkPK),
	})
}

func (p *Price) UnmarshalJSON(bytes []byte) error {
	j := &jsonPrice{}
	err := json.Unmarshal(bytes, j)
	if err != nil {
		return errUnmarshalling("price fields errors", err)
	}

	j.V = strings.TrimPrefix(j.V, "0x")
	j.R = strings.TrimPrefix(j.R, "0x")
	j.S = strings.TrimPrefix(j.S, "0x")

	if (len(j.V)+len(j.R)+len(j.S) != 0) && (len(j.V) != 2 || len(j.R) != 64 || len(j.S) != 64) {
		return errUnmarshalling("VRS fields contain invalid signature lengths", err)
	}

	p.Wat = j.Wat
	p.Val, _ = new(big.Int).SetString(j.Val, 10)
	p.Age = time.Unix(j.Age, 0)

	if len(j.V) != 0 {
		v := [1]byte{}
		_, err = hex.Decode(v[:], []byte(j.V))
		if err != nil {
			return errUnmarshalling("unable to decode V param", err)
		}
		p.V = v[0]
	}

	if len(j.R) != 0 {
		_, err = hex.Decode(p.R[:], []byte(j.R))
		if err != nil {
			return errUnmarshalling("unable to decode R param", err)
		}
	}

	if len(j.S) != 0 {
		_, err = hex.Decode(p.S[:], []byte(j.S))
		if err != nil {
			return errUnmarshalling("unable to decode S param", err)
		}
	}

	p.StarkR, err = decodeHexNumber(j.StarkR)
	if err != nil {
		return errUnmarshalling("unable to decode StarkR param", err)
	}

	p.StarkS, err = decodeHexNumber(j.StarkS)
	if err != nil {
		return errUnmarshalling("unable to decode StarkS param", err)
	}

	p.StarkPK, err = decodeHexNumber(j.StarkPK)
	if err != nil {
		return errUnmarshalling("unable to decode StarkPK param", err)
	}

	return nil
}

func decodeHexNumber(s string) ([]byte, error) {
	n, ok := (&big.Int{}).SetString(strings.TrimPrefix(s, "0x"), 16)
	if !ok {
		return nil, errors.New("unable to parse hex number")
	}
	return n.Bytes(), nil
}

func encodeHexNumber(b []byte) string {
	n := (&big.Int{}).SetBytes(b)
	return "0x" + n.Text(16)
}

// hash is an equivalent of keccak256(abi.encodePacked(val_, age_, wat))) in Solidity.
func (p *Price) hash() []byte {
	// Median:
	median := make([]byte, 32)
	p.Val.FillBytes(median)

	// Time:
	age := make([]byte, 32)
	binary.BigEndian.PutUint64(age[24:], uint64(p.Age.Unix()))

	// Asset name:
	wat := make([]byte, 32)
	copy(wat, p.Wat)

	hash := make([]byte, 96)
	copy(hash[0:32], median)
	copy(hash[32:64], age)
	copy(hash[64:96], wat)

	return ethereum.SHA3Hash(hash)
}
