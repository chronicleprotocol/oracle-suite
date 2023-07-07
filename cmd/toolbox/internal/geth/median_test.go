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

package geth

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
)

func TestMedian_Age(t *testing.T) {
	// Prepare test data:
	c := &mocks.Client{}
	a := types.Address{}
	m := NewMedian(c, a)

	// Call Age function:
	bts := make([]byte, 32)
	big.NewInt(123456).FillBytes(bts)
	c.On("Call", mock.Anything, mock.Anything).Return(bts, nil)
	age, err := m.Age(context.Background())

	// Verify:
	assert.NoError(t, err)
	assert.Equal(t, int64(123456), age.Unix())
}

func TestMedian_Bar(t *testing.T) {
	// Prepare test data:
	c := &mocks.Client{}
	a := types.Address{}
	m := NewMedian(c, a)

	// Call Bar function:
	bts := make([]byte, 32)
	big.NewInt(13).FillBytes(bts)
	c.On("Call", mock.Anything, mock.Anything).Return(bts, nil)
	bar, err := m.Bar(context.Background())

	// Verify:
	assert.NoError(t, err)
	assert.Equal(t, int64(13), bar)
}

func TestMedian_Price(t *testing.T) {
	// Prepare test data:
	c := &mocks.Client{}
	a := types.Address{}
	m := NewMedian(c, a)

	// Call Val function:
	bts := make([]byte, 32)
	val := new(big.Int).Mul(big.NewInt(42), big.NewInt(median.PriceMultiplier))
	val.FillBytes(bts)
	c.On("Storage", mock.Anything, a, types.MustHashFromBigInt(big.NewInt(1))).Return(bts, nil)
	price, err := m.Val(context.Background())

	// Verify:
	assert.NoError(t, err)
	assert.Equal(t, val.String(), price.String())
}

func TestMedian_Poke(t *testing.T) {
	// Prepare test data:
	c := &mocks.Client{}
	a := types.Address{}
	s := &mocks.Key{}
	m := NewMedian(c, a)

	p1 := &median.Price{Wat: "AAABBB"}
	p1.SetFloat64Price(10)
	p1.Age = time.Unix(0xAAAAAAAA, 0)
	p1.Sig.V = big.NewInt(0xA1)
	p1.Sig.R = big.NewInt(0xA2)
	p1.Sig.S = big.NewInt(0xA3)

	p2 := &median.Price{Wat: "AAABBB"}
	p2.SetFloat64Price(30)
	p2.Age = time.Unix(0xBBBBBBBB, 0)
	p2.Sig.V = big.NewInt(0xB1)
	p2.Sig.R = big.NewInt(0xB2)
	p2.Sig.S = big.NewInt(0xB3)

	p3 := &median.Price{Wat: "AAABBB"}
	p3.SetFloat64Price(20)
	p3.Age = time.Unix(0xCCCCCCCC, 0)
	p3.Sig.V = big.NewInt(0xC1)
	p3.Sig.R = big.NewInt(0xC2)
	p3.Sig.S = big.NewInt(0xC3)

	s.On("SignMessage", mock.Anything).Return(types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65)), nil).Once()
	p1.Sign(s)
	s.On("SignMessage", mock.Anything).Return(types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xBB}, 65)), nil).Once()
	p2.Sign(s)
	s.On("SignMessage", mock.Anything).Return(types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xCC}, 65)), nil).Once()
	p3.Sign(s)

	c.On("SendTransaction", mock.Anything, mock.Anything).Return(&types.Hash{}, nil)

	// Call Poke function:
	_, err := m.Poke(context.Background(), []*median.Price{p1, p2, p3}, false)
	assert.NoError(t, err)

	// Verify generated transaction:
	tx := c.Calls[0].Arguments.Get(1).(*types.Transaction)
	cd := "89bbb8b2" +
		// Offsets:
		"00000000000000000000000000000000000000000000000000000000000000a0" +
		"0000000000000000000000000000000000000000000000000000000000000120" +
		"00000000000000000000000000000000000000000000000000000000000001a0" +
		"0000000000000000000000000000000000000000000000000000000000000220" +
		"00000000000000000000000000000000000000000000000000000000000002a0" +
		// Val:
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"0000000000000000000000000000000000000000000000008ac7230489e80000" +
		"000000000000000000000000000000000000000000000001158e460913d00000" +
		"000000000000000000000000000000000000000000000001a055690d9db80000" +
		// Age:
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"00000000000000000000000000000000000000000000000000000000aaaaaaaa" +
		"00000000000000000000000000000000000000000000000000000000cccccccc" +
		"00000000000000000000000000000000000000000000000000000000bbbbbbbb" +
		// V:
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"00000000000000000000000000000000000000000000000000000000000000aa" +
		"00000000000000000000000000000000000000000000000000000000000000cc" +
		"00000000000000000000000000000000000000000000000000000000000000bb" +
		// R:
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc" +
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" +
		// S:
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc" +
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	assert.Equal(t, a, *tx.To)
	assert.Equal(t, (*big.Int)(nil), tx.MaxFeePerGas)
	assert.Equal(t, uint64(gasLimit), *tx.GasLimit)
	assert.Nil(t, tx.Nonce)
	assert.Equal(t, cd, hex.EncodeToString(tx.Input))
}
