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

package store

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

var (
	aaabbb1 = &messages.MuSigSignature{
		MuSigMessage: &messages.MuSigMessage{
			Signers: []types.Address{types.MustAddressFromHex("0x1111111111111111111111111111111111111111")},
			MsgMeta: messages.MuSigMeta{Meta: messages.MuSigMetaTickV1{
				Wat: "AAA/BBB",
				Val: bn.DecFixedPoint(100, chronicle.ScribePricePrecision),
				Age: time.Unix(100, 0),
			}},
		},
		Commitment:       types.MustAddressFromHex("0x1111111111111111111111111111111111111111"),
		SchnorrSignature: big.NewInt(123456789),
		ComputedAt:       time.Unix(100, 0),
	}
	aaabbb2 = &messages.MuSigSignature{
		MuSigMessage: &messages.MuSigMessage{
			Signers: []types.Address{types.MustAddressFromHex("0x1111111111111111111111111111111111111111")},
			MsgMeta: messages.MuSigMeta{Meta: messages.MuSigMetaTickV1{
				Wat: "AAA/BBB",
				Val: bn.DecFixedPoint(110, chronicle.ScribePricePrecision),
				Age: time.Unix(90, 0),
			}},
		},
		Commitment:       types.MustAddressFromHex("0x1111111111111111111111111111111111111111"),
		SchnorrSignature: big.NewInt(123456789),
		ComputedAt:       time.Unix(90, 0),
	}
	xxxyyy1 = &messages.MuSigSignature{
		MuSigMessage: &messages.MuSigMessage{
			Signers: []types.Address{types.MustAddressFromHex("0x1111111111111111111111111111111111111111")},
			MsgMeta: messages.MuSigMeta{Meta: messages.MuSigMetaTickV1{
				Wat: "XXX/YYY",
				Val: bn.DecFixedPoint(100, chronicle.ScribePricePrecision),
				Age: time.Unix(90, 0),
			}},
		},
		Commitment:       types.MustAddressFromHex("0x1111111111111111111111111111111111111111"),
		SchnorrSignature: big.NewInt(123456789),
		ComputedAt:       time.Unix(90, 0),
	}
	xxxyyy2 = &messages.MuSigSignature{
		MuSigMessage: &messages.MuSigMessage{
			Signers: []types.Address{types.MustAddressFromHex("0x1111111111111111111111111111111111111111")},
			MsgMeta: messages.MuSigMeta{Meta: messages.MuSigMetaTickV1{
				Wat: "XXX/YYY",
				Val: bn.DecFixedPoint(110, chronicle.ScribePricePrecision),
				Age: time.Unix(100, 0),
			}},
		},
		Commitment:       types.MustAddressFromHex("0x1111111111111111111111111111111111111111"),
		SchnorrSignature: big.NewInt(123456789),
		ComputedAt:       time.Unix(100, 0),
	}
)

func TestStore(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	testTransport := local.New(
		[]byte("test"),
		0,
		map[string]transport.Message{messages.MuSigSignatureV1MessageName: (*messages.MuSigSignature)(nil)},
	)

	testStore := New(Config{
		Transport:  testTransport,
		DataModels: []string{"AAA/BBB", "XXX/YYY"},
		Logger:     null.New(),
	})

	require.NoError(t, testTransport.Start(ctx))
	require.NoError(t, testStore.Start(ctx))
	time.Sleep(100 * time.Millisecond) // Wait for services to start.

	assert.NoError(t, testTransport.Broadcast(messages.MuSigSignatureV1MessageName, aaabbb1))
	assert.NoError(t, testTransport.Broadcast(messages.MuSigSignatureV1MessageName, aaabbb2))
	assert.NoError(t, testTransport.Broadcast(messages.MuSigSignatureV1MessageName, xxxyyy1))
	assert.NoError(t, testTransport.Broadcast(messages.MuSigSignatureV1MessageName, xxxyyy2))

	// Wait to be sure that the store has processed the messages.
	assert.Eventually(t, func() bool {
		a := testStore.SignaturesByDataModel("AAA/BBB")
		b := testStore.SignaturesByDataModel("XXX/YYY")
		return len(a) == 1 && len(b) == 1
	}, 1*time.Second, 100*time.Millisecond)

	a := testStore.SignaturesByDataModel("AAA/BBB")
	b := testStore.SignaturesByDataModel("XXX/YYY")

	assert.Equal(t, "100", a[0].MsgMeta.TickV1().Val.String())
	assert.Equal(t, "110", b[0].MsgMeta.TickV1().Val.String())
}
