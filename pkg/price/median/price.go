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

package median

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const PriceMultiplier = 1e18

var ErrPriceNotSet = errors.New("unable to sign a price because the price is not set")
var ErrUnmarshallingFailure = errors.New("unable to unmarshal given JSON")

func errUnmarshalling(s string, err error) error {
	return fmt.Errorf("%w: %s: %s", ErrUnmarshallingFailure, s, err)
}

type Price struct {
	Wat string          // Wat is the asset name.
	Val *big.Int        // Val is the asset price multiplied by PriceMultiplier.
	Age time.Time       // Age is the time when the price was obtained.
	Sig types.Signature // Sig is the signature of the price.
}

// jsonPrice is the JSON representation of the Price structure.
type jsonPrice struct {
	Wat string `json:"wat"`
	Val string `json:"val"`
	Age int64  `json:"age"`
	V   string `json:"v"`
	R   string `json:"r"`
	S   string `json:"s"`
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

func (p *Price) From(r crypto.Recoverer) (*types.Address, error) {
	return r.RecoverMessage(p.Hash().Bytes(), p.Sig)
}

func (p *Price) Sign(signer wallet.Key) error {
	if p.Val == nil {
		return ErrPriceNotSet
	}
	signature, err := signer.SignMessage(p.Hash().Bytes())
	if err != nil {
		return err
	}
	p.Sig = *signature
	return nil
}

func (p *Price) Fields(r crypto.Recoverer) log.Fields {
	from := "*invalid signature*"
	if addr, err := p.From(r); err == nil {
		from = addr.String()
	}
	return log.Fields{
		"from": from,
		"wat":  p.Wat,
		"age":  p.Age.UTC().Format(time.RFC3339),
		"val":  p.Val.String(),
		"hash": hex.EncodeToString(p.Hash().Bytes()),
		"V":    hex.EncodeToString(p.Sig.V.Bytes()),
		"R":    hex.EncodeToString(p.Sig.R.Bytes()),
		"S":    hex.EncodeToString(p.Sig.S.Bytes()),
	}
}

func (p *Price) MarshalJSON() ([]byte, error) {
	bts := p.Sig.Bytes()
	v := bts[64]
	r := bts[:32]
	s := bts[32:64]
	return json.Marshal(jsonPrice{
		Wat: p.Wat,
		Val: p.Val.String(),
		Age: p.Age.Unix(),
		V:   hex.EncodeToString([]byte{v}),
		R:   hex.EncodeToString(r),
		S:   hex.EncodeToString(s),
	})
}

func (p *Price) UnmarshalJSON(bytes []byte) error {
	j := &jsonPrice{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return errUnmarshalling("price fields errors", err)
	}

	j.V = strings.TrimPrefix(j.V, "0x")
	j.R = strings.TrimPrefix(j.R, "0x")
	j.S = strings.TrimPrefix(j.S, "0x")

	if (len(j.V)+len(j.R)+len(j.S) != 0) && (len(j.V) != 2 || len(j.R) != 64 || len(j.S) != 64) {
		return errUnmarshalling("VRS fields contain invalid signature lengths", nil)
	}

	p.Wat = j.Wat
	p.Val, _ = new(big.Int).SetString(j.Val, 10)
	p.Age = time.Unix(j.Age, 0)

	v, err := hex.DecodeString(j.V)
	if err != nil {
		return errUnmarshalling("unable to decode V param", err)
	}
	r, err := hex.DecodeString(j.R)
	if err != nil {
		return errUnmarshalling("unable to decode R param", err)
	}
	s, err := hex.DecodeString(j.S)
	if err != nil {
		return errUnmarshalling("unable to decode S param", err)
	}

	p.Sig = types.SignatureFromVRS(
		new(big.Int).SetBytes(v),
		new(big.Int).SetBytes(r),
		new(big.Int).SetBytes(s),
	)

	return nil
}

// Hash is an equivalent of keccak256(abi.encodePacked(val_, age_, wat))) in Solidity.
func (p *Price) Hash() types.Hash {
	return crypto.Keccak256(p.h())
}

func (p *Price) h() []byte {
	// Median:
	val := make([]byte, 32)
	p.Val.FillBytes(val)

	// Time:
	age := make([]byte, 32)
	binary.BigEndian.PutUint64(age[24:], uint64(p.Age.Unix()))

	// Asset name:
	wat := make([]byte, 32)
	copy(wat, p.Wat)

	h := make([]byte, 96)
	copy(h[0:32], val)
	copy(h[32:64], age)
	copy(h[64:96], wat)
	return h
}
