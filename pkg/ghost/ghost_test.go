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

package ghost

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	priceMocks "github.com/chronicleprotocol/oracle-suite/pkg/price/provider/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

var (
	PriceAAABBB1 = &provider.Price{
		Type:       "median",
		Parameters: nil,
		Pair: provider.Pair{
			Base:  "AAA",
			Quote: "BBB",
		},
		Price:     110,
		Bid:       109,
		Ask:       111,
		Volume24h: 110,
		Time:      time.Unix(100, 0),
		Prices:    nil,
		Error:     "",
	}
	PriceXXXYYY1 = &provider.Price{
		Type:       "median",
		Parameters: nil,
		Pair: provider.Pair{
			Base:  "XXX",
			Quote: "YYY",
		},
		Price:     210,
		Bid:       209,
		Ask:       211,
		Volume24h: 210,
		Time:      time.Unix(200, 0),
		Prices:    nil,
		Error:     "",
	}
)

func TestGhost(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer ctxCancel()

	pri := &priceMocks.Provider{}
	sig := &ethereumMocks.Signer{}
	tra := local.New([]byte("test"), 0, map[string]transport.Message{
		messages.PriceV0MessageName: (*messages.Price)(nil),
		messages.PriceV1MessageName: (*messages.Price)(nil),
	})
	_ = tra.Start(ctx)

	ps, err := New(Config{
		Pairs:         []string{"AAA/BBB", "XXX/YYY"},
		PriceProvider: pri,
		Signer:        sig,
		Transport:     tra,
		Interval:      time.Second,
		Logger:        null.New(),
	})
	require.NoError(t, err)
	require.NoError(t, ps.Start(ctx))

	pri.On("Price", provider.Pair{Base: "AAA", Quote: "BBB"}).Return(PriceAAABBB1, nil)
	pri.On("Price", provider.Pair{Base: "XXX", Quote: "YYY"}).Return(PriceXXXYYY1, nil)
	sig.On("Signature", errutil.Must(hex.DecodeString("9315c7118c32ce6c778bf691147c554afd2dc816b5c6bd191ac03784f69aa004"))).Return(ethereum.SignatureFromBytes(bytes.Repeat([]byte{0xAA}, 65)), nil)
	sig.On("Signature", errutil.Must(hex.DecodeString("8dd1c8d47ec9eafda294cfc8c0c8d4041a13d7a89536a89eb6685a79d9fa6bc4"))).Return(ethereum.SignatureFromBytes(bytes.Repeat([]byte{0xAA}, 65)), nil)

	// Wait for two messages. They should be sent after 2 seconds.
	var pricesV0, pricesV1 []*messages.Price
	for {
		select {
		case msg := <-tra.Messages(messages.PriceV0MessageName):
			price := msg.Message.(*messages.Price)
			pricesV0 = append(pricesV0, price)
		case msg := <-tra.Messages(messages.PriceV1MessageName):
			price := msg.Message.(*messages.Price)
			pricesV1 = append(pricesV1, price)
		}
		if len(pricesV0) == 2 && len(pricesV1) == 2 {
			break
		}
	}
	sort.Slice(pricesV0, func(i, j int) bool {
		return pricesV0[i].Price.Wat < pricesV0[j].Price.Wat
	})
	sort.Slice(pricesV1, func(i, j int) bool {
		return pricesV1[i].Price.Wat < pricesV1[j].Price.Wat
	})

	require.Len(t, pricesV0, 2)
	require.Len(t, pricesV1, 2)

	assertPrice(t, PriceAAABBB1, pricesV0[0])
	assertPrice(t, PriceXXXYYY1, pricesV0[1])
	assertPrice(t, PriceAAABBB1, pricesV1[0])
	assertPrice(t, PriceXXXYYY1, pricesV1[1])
}

func assertPrice(t *testing.T, expected *provider.Price, actual *messages.Price) {
	p, _ := new(big.Float).SetInt(actual.Price.Val).Float64()
	assert.Equal(t, actual.Price.Age.Unix(), expected.Time.Unix())
	assert.Equal(t, actual.Price.Wat, expected.Pair.Base+expected.Pair.Quote)
	assert.Equal(t, p/oracle.PriceMultiplier, expected.Price)
	assert.Equal(t, actual.Price.V, byte(0xAA))
	assert.Equal(t, actual.Price.R, [32]byte(common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")))
	assert.Equal(t, actual.Price.S, [32]byte(common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")))
}
