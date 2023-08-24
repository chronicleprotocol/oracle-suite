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

package types

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

// BlockNumber is a type that can hold a block number or a tag.
type BlockNumber struct{ x big.Int }

const (
	earliestBlockNumber = -1
	latestBlockNumber   = -2
	pendingBlockNumber  = -3
)

var (
	EarliestBlockNumber = BlockNumber{x: *new(big.Int).SetInt64(earliestBlockNumber)}
	LatestBlockNumber   = BlockNumber{x: *new(big.Int).SetInt64(latestBlockNumber)}
	PendingBlockNumber  = BlockNumber{x: *new(big.Int).SetInt64(pendingBlockNumber)}
)

// StringToBlockNumber converts a string to a BlockNumber. A string can be a
// hex number, "earliest", "latest" or "pending".
func StringToBlockNumber(str string) BlockNumber {
	b := &BlockNumber{}
	_ = b.UnmarshalText([]byte(str))
	return *b
}

// BigToBlockNumber converts a big.Int to a BlockNumber.
func BigToBlockNumber(x *big.Int) BlockNumber {
	return BlockNumber{x: *new(big.Int).Set(x)}
}

// Uint64ToBlockNumber converts a uint64 to a BlockNumber.
func Uint64ToBlockNumber(x uint64) BlockNumber {
	return BlockNumber{x: *new(big.Int).SetUint64(x)}
}

func (t *BlockNumber) IsEarliest() bool {
	return t.Big().Int64() == earliestBlockNumber
}

func (t *BlockNumber) IsLatest() bool {
	return t.Big().Int64() == latestBlockNumber
}

func (t *BlockNumber) IsPending() bool {
	return t.Big().Int64() == pendingBlockNumber
}

func (t *BlockNumber) IsTag() bool {
	return t.Big().Sign() < 0
}

func (t *BlockNumber) Big() *big.Int {
	return new(big.Int).Set(&t.x)
}

func (t *BlockNumber) String() string {
	switch {
	case t.IsEarliest():
		return "earliest"
	case t.IsLatest():
		return "latest"
	case t.IsPending():
		return "pending"
	default:
		return "0x" + t.x.Text(16)
	}
}

// MarshalJSON implements json.Marshaler.
func (t BlockNumber) MarshalJSON() ([]byte, error) {
	b, err := t.MarshalText()
	if err != nil {
		return nil, err
	}
	return naiveQuote(b), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *BlockNumber) UnmarshalJSON(input []byte) error {
	return t.UnmarshalText(naiveUnquote(input))
}

func (t BlockNumber) MarshalText() ([]byte, error) {
	switch {
	case t.IsEarliest():
		return []byte("earliest"), nil
	case t.IsLatest():
		return []byte("latest"), nil
	case t.IsPending():
		return []byte("pending"), nil
	default:
		return bigIntToHex(&t.x), nil
	}
}

func (t *BlockNumber) UnmarshalText(input []byte) error {
	switch strings.TrimSpace(string(input)) {
	case "earliest":
		*t = BlockNumber{x: *new(big.Int).SetInt64(earliestBlockNumber)}
		return nil
	case "latest":
		*t = BlockNumber{x: *new(big.Int).SetInt64(latestBlockNumber)}
		return nil
	case "pending":
		*t = BlockNumber{x: *new(big.Int).SetInt64(pendingBlockNumber)}
		return nil
	default:
		u, err := hexToBigInt(input)
		if err != nil {
			return err
		}
		if u.Cmp(big.NewInt(math.MaxInt64)) > 0 {
			return fmt.Errorf("block number larger than int64")
		}
		*t = BlockNumber{x: *u}
		return nil
	}
}
