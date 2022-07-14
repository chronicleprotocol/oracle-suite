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

package geth

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
)

func TestMedian_Age(t *testing.T) {
	// Prepare test data:
	c := &mocks.Client{}
	a := ethereum.Address{}
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
	a := ethereum.Address{}
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
	a := ethereum.Address{}
	m := NewMedian(c, a)

	// Call Val function:
	bts := make([]byte, 32)
	val := new(big.Int).Mul(big.NewInt(42), big.NewInt(oracle.PriceMultiplier))
	val.FillBytes(bts)
	c.On("Storage", mock.Anything, a, common.BigToHash(big.NewInt(1))).Return(bts, nil)
	price, err := m.Val(context.Background())

	// Verify:
	assert.NoError(t, err)
	assert.Equal(t, val.String(), price.String())
}

func TestMedian_Poke(t *testing.T) {
	// Prepare test data:
	c := &mocks.Client{}
	a := ethereum.Address{}
	s := &mocks.Signer{}
	m := NewMedian(c, a)

	p1 := &oracle.Price{Wat: "AAABBB"}
	p1.SetFloat64Price(10)
	p1.Age = time.Unix(0xAAAAAAAA, 0)
	p1.V = 0xA1
	p1.R = [32]byte{0xA2}
	p1.S = [32]byte{0xA3}

	p2 := &oracle.Price{Wat: "AAABBB"}
	p2.SetFloat64Price(30)
	p2.Age = time.Unix(0xBBBBBBBB, 0)
	p2.V = 0xB1
	p2.R = [32]byte{0xB2}
	p2.S = [32]byte{0xB3}

	p3 := &oracle.Price{Wat: "AAABBB"}
	p3.SetFloat64Price(20)
	p3.Age = time.Unix(0xCCCCCCCC, 0)
	p3.V = 0xC1
	p3.R = [32]byte{0xC2}
	p3.S = [32]byte{0xC3}

	s.On("Signature", mock.Anything).Return(ethereum.SignatureFromBytes(bytes.Repeat([]byte{0xAA}, 65)), nil).Once()
	p1.Sign(s)
	s.On("Signature", mock.Anything).Return(ethereum.SignatureFromBytes(bytes.Repeat([]byte{0xBB}, 65)), nil).Once()
	p2.Sign(s)
	s.On("Signature", mock.Anything).Return(ethereum.SignatureFromBytes(bytes.Repeat([]byte{0xCC}, 65)), nil).Once()
	p3.Sign(s)

	c.On("SendTransaction", mock.Anything, mock.Anything).Return(&ethereum.Hash{}, nil)

	// Call Poke function:
	_, err := m.Poke(context.Background(), []*oracle.Price{p1, p2, p3}, false)
	assert.NoError(t, err)

	// Verify generated transaction:
	tx := c.Calls[0].Arguments.Get(1).(*ethereum.Transaction)
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

	assert.Equal(t, a, tx.Address)
	assert.Equal(t, (*big.Int)(nil), tx.MaxFee)
	assert.Equal(t, big.NewInt(gasLimit), tx.GasLimit)
	assert.Equal(t, uint64(0), tx.Nonce)
	assert.Equal(t, cd, hex.EncodeToString(tx.Data))
}
