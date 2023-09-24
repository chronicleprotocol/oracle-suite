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

package signer

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
)

// Hash for the AAABBB asset pair, with the price set to 42 and the age to 1605371361:
var priceHash = "0x5e7aa8f6514c872b2020a7f63c72a382e813dc0624a2fb3c28367fee763be154"

func TestTick_Supports(t *testing.T) {
	t.Run("supported data point", func(t *testing.T) {
		k := &mocks.Key{}
		s := NewTickSigner(k)
		assert.True(t, s.Supports(context.Background(), datapoint.Point{Value: value.Tick{}}))
	})
	t.Run("unsupported data point", func(t *testing.T) {
		k := &mocks.Key{}
		s := NewTickSigner(k)
		assert.False(t, s.Supports(context.Background(), datapoint.Point{Value: value.StaticValue{}}))
	})
}

func TestTick_Sign(t *testing.T) {
	k := &mocks.Key{}
	s := NewTickSigner(k)

	expSig := types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65))
	k.On("SignHash", types.MustHashFromHex(priceHash, types.PadNone)).Return(expSig, nil).Once()

	retSig, err := s.Sign(context.Background(), "AAABBB", datapoint.Point{
		Value:     value.NewTick(value.Pair{Base: "AAA", Quote: "BBB"}, 42, 0),
		Time:      time.Unix(1605371361, 0),
		SubPoints: nil,
		Meta:      nil,
		Error:     nil,
	})
	require.NoError(t, err)

	assert.Equal(t, *expSig, *retSig)
}

func TestTick_Recover(t *testing.T) {
	r := &mocks.Recoverer{}
	s := NewTickRecoverer(r)

	msgSig := types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65))
	expAddr := types.MustAddressFromHexPtr("0x1234567890123456789012345678901234567890")
	r.On("RecoverHash", types.MustHashFromHex(priceHash, types.PadNone), *msgSig).Return(expAddr, nil).Once()

	retAddr, err := s.Recover(context.Background(), "AAABBB", datapoint.Point{
		Value:     value.NewTick(value.Pair{Base: "AAA", Quote: "BBB"}, 42, 0),
		Time:      time.Unix(1605371361, 0),
		SubPoints: nil,
		Meta:      nil,
		Error:     nil,
	}, *msgSig)
	require.NoError(t, err)

	assert.Equal(t, *expAddr, *retAddr)
}
