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

package contract

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestOpScribe_OpPoke(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRPC)
	scribe := NewOpScribe(mockClient, types.MustAddressFromHex("0x1122344556677889900112233445566778899002"))

	// Mocked data for the test
	pokeData := PokeData{
		Val: bn.DecFixedPoint(26064.535, 18),
		Age: time.Unix(1692913991, 0),
	}
	schnorrData := SchnorrData{
		Signature:   new(big.Int).SetBytes(hexutil.MustHexToBytes("0x1234567890123456789012345678901234567890123456789012345678901234")),
		Commitment:  types.MustAddressFromHex("0x1234567890123456789012345678901234567890"),
		SignersBlob: []byte{0x01, 0x02, 0x03, 0x04},
	}
	ecdsaData := types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00")

	calldata := hexutil.MustHexToBytes(
		"0x" +
			"6712af9e" +
			"000000000000000000000000000000000000000000000584f61606acd0134800" +
			"0000000000000000000000000000000000000000000000000000000064e7d147" +
			"00000000000000000000000000000000000000000000000000000000000000c0" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff" +
			"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff" +
			"1234567890123456789012345678901234567890123456789012345678901234" +
			"0000000000000000000000001234567890123456789012345678901234567890" +
			"0000000000000000000000000000000000000000000000000000000000000060" +
			"0000000000000000000000000000000000000000000000000000000000000004" +
			"0102030400000000000000000000000000000000000000000000000000000000",
	)

	mockClient.On(
		"Call",
		ctx,
		types.Call{
			To:    &scribe.address,
			Input: calldata,
		},
		types.LatestBlockNumber,
	).
		Return(
			[]byte{},
			&types.Call{},
			nil,
		)

	mockClient.On(
		"SendTransaction",
		ctx,
		types.Transaction{
			Call: types.Call{
				To:    &scribe.address,
				Input: calldata,
			},
		},
	).
		Return(
			&types.Hash{},
			&types.Transaction{},
			nil,
		)

	_, _, err := scribe.OpPoke(ctx, pokeData, schnorrData, ecdsaData)
	require.NoError(t, err)
}
