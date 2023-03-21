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

package feeder

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	priceMocks "github.com/chronicleprotocol/oracle-suite/pkg/price/provider/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

var (
	PriceAAABBB = &provider.Price{
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
	PriceXXXYYY = &provider.Price{
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
	InvalidPriceAAABBB = &provider.Price{
		Type:       "median",
		Parameters: nil,
		Pair: provider.Pair{
			Base:  "AAA",
			Quote: "BBB",
		},
		Price:     0,
		Bid:       0,
		Ask:       0,
		Volume24h: 0,
		Time:      time.Unix(0, 0),
		Prices:    nil,
		Error:     "err",
	}
	PriceAAABBBHash = types.MustHashFromHex("9315c7118c32ce6c778bf691147c554afd2dc816b5c6bd191ac03784f69aa004", types.PadNone)
	PriceXXXYYYHash = types.MustHashFromHex("8dd1c8d47ec9eafda294cfc8c0c8d4041a13d7a89536a89eb6685a79d9fa6bc4", types.PadNone)
)

func TestFeeder_Broadcast(t *testing.T) {
	tests := []struct {
		name    string
		prices  int
		mocks   func(*priceMocks.Provider, *ethereumMocks.Key)
		asserts func(t *testing.T, pricesV0, pricesV1 []*messages.Price)
	}{
		{
			name:   "valid-prices",
			prices: 2,
			mocks: func(pro *priceMocks.Provider, sig *ethereumMocks.Key) {
				pro.On("Price", provider.Pair{Base: "AAA", Quote: "BBB"}).Return(PriceAAABBB, nil).Times(1)
				pro.On("Price", provider.Pair{Base: "XXX", Quote: "YYY"}).Return(PriceXXXYYY, nil).Times(1)
				sig.On("SignMessage", PriceAAABBBHash.Bytes()).Return(types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65)), nil)
				sig.On("SignMessage", PriceXXXYYYHash.Bytes()).Return(types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65)), nil)
			},
			asserts: func(t *testing.T, pricesV0, pricesV1 []*messages.Price) {
				require.Len(t, pricesV0, 2)
				require.Len(t, pricesV1, 2)
				assertPrice(t, PriceAAABBB, pricesV0[0])
				assertPrice(t, PriceXXXYYY, pricesV0[1])
				assertPrice(t, PriceAAABBB, pricesV1[0])
				assertPrice(t, PriceXXXYYY, pricesV1[1])
			},
		},
		{
			name:   "invalid-price",
			prices: 1,
			mocks: func(pro *priceMocks.Provider, sig *ethereumMocks.Key) {
				pro.On("Price", provider.Pair{Base: "AAA", Quote: "BBB"}).Return(InvalidPriceAAABBB, nil).Times(1)
				pro.On("Price", provider.Pair{Base: "XXX", Quote: "YYY"}).Return(PriceXXXYYY, nil).Times(1)
				sig.On("SignMessage", PriceXXXYYYHash.Bytes()).Return(types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65)), nil)
			},
			asserts: func(t *testing.T, pricesV0, pricesV1 []*messages.Price) {
				require.Len(t, pricesV0, 1)
				require.Len(t, pricesV1, 1)
				assertPrice(t, PriceXXXYYY, pricesV0[0])
				assertPrice(t, PriceXXXYYY, pricesV1[0])
			},
		},
		{
			name:   "price-unavailable",
			prices: 1,
			mocks: func(pro *priceMocks.Provider, sig *ethereumMocks.Key) {
				pro.On("Price", provider.Pair{Base: "AAA", Quote: "BBB"}).Return((*provider.Price)(nil), errors.New("err")).Times(1)
				pro.On("Price", provider.Pair{Base: "XXX", Quote: "YYY"}).Return(PriceXXXYYY, nil).Times(1)
				sig.On("SignMessage", PriceXXXYYYHash.Bytes()).Return(types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65)), nil)
			},
			asserts: func(t *testing.T, pricesV0, pricesV1 []*messages.Price) {
				require.Len(t, pricesV0, 1)
				require.Len(t, pricesV1, 1)
				assertPrice(t, PriceXXXYYY, pricesV0[0])
				assertPrice(t, PriceXXXYYY, pricesV1[0])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
			defer ctxCancel()

			// Prepare feeder services.
			priceProvider := &priceMocks.Provider{}
			signer := &ethereumMocks.Key{}

			ticker := timeutil.NewTicker(0)
			localTransport := local.New([]byte("test"), 0, map[string]transport.Message{
				messages.PriceV0MessageName: (*messages.Price)(nil),
				messages.PriceV1MessageName: (*messages.Price)(nil),
			})

			// Prepare mocks.
			tt.mocks(priceProvider, signer)

			// Start feeder.
			feeder, err := New(Config{
				Pairs:         []string{"AAA/BBB", "XXX/YYY"},
				PriceProvider: priceProvider,
				Signer:        signer,
				Transport:     localTransport,
				Interval:      ticker,
			})
			require.NoError(t, err)
			require.NoError(t, localTransport.Start(ctx))
			require.NoError(t, feeder.Start(ctx))
			defer func() {
				ctxCancel()
				<-feeder.Wait()
				<-localTransport.Wait()
			}()

			ticker.Tick()

			// Wait for two messages.
			var pricesV0, pricesV1 []*messages.Price
			v0ch := localTransport.Messages(messages.PriceV0MessageName)
			v1ch := localTransport.Messages(messages.PriceV1MessageName)
			for {
				select {
				case msg := <-v0ch:
					price := msg.Message.(*messages.Price)
					pricesV0 = append(pricesV0, price)
				case msg := <-v1ch:
					price := msg.Message.(*messages.Price)
					pricesV1 = append(pricesV1, price)
				}
				if len(pricesV0) >= tt.prices && len(pricesV1) >= tt.prices {
					break
				}
			}

			sort.Slice(pricesV0, func(i, j int) bool {
				return pricesV0[i].Price.Wat < pricesV0[j].Price.Wat
			})
			sort.Slice(pricesV1, func(i, j int) bool {
				return pricesV1[i].Price.Wat < pricesV1[j].Price.Wat
			})

			// Assert results.
			tt.asserts(t, pricesV0, pricesV1)
		})
	}
}

func TestFeeder_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "minimal-valid-config",
			cfg: Config{
				PriceProvider: &priceMocks.Provider{},
				Signer:        &ethereumMocks.Key{},
				Transport:     local.New([]byte("test"), 0, nil),
			},
			wantErr: false,
		},
		{
			name: "invalid-pair",
			cfg: Config{
				PriceProvider: &priceMocks.Provider{},
				Signer:        &ethereumMocks.Key{},
				Transport:     local.New([]byte("test"), 0, nil),
				Pairs:         []string{"AAABBB"},
			},
			wantErr: true,
		},
		{
			name: "missing-price-provider",
			cfg: Config{
				PriceProvider: nil,
				Signer:        &ethereumMocks.Key{},
				Transport:     local.New([]byte("test"), 0, nil),
			},
			wantErr: true,
		},
		{
			name: "missing-signer",
			cfg: Config{
				PriceProvider: &priceMocks.Provider{},
				Signer:        nil,
				Transport:     local.New([]byte("test"), 0, nil),
			},
			wantErr: true,
		},
		{
			name: "missing-transport",
			cfg: Config{
				PriceProvider: &priceMocks.Provider{},
				Signer:        &ethereumMocks.Key{},
				Transport:     nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFeeder_Start(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer ctxCancel()

	pro := &priceMocks.Provider{}
	sig := &ethereumMocks.Key{}
	tra := local.New([]byte("test"), 0, map[string]transport.Message{})
	_ = tra.Start(ctx)
	defer func() {
		<-tra.Wait()
	}()

	gho, err := New(Config{
		Pairs:         []string{},
		PriceProvider: pro,
		Signer:        sig,
		Transport:     tra,
		Interval:      timeutil.NewTicker(time.Second),
	})
	require.NoError(t, err)
	require.Error(t, gho.Start(nil)) // Start without context should fail.
	require.NoError(t, gho.Start(ctx))
	require.Error(t, gho.Start(ctx)) // Second start should fail.
	ctxCancel()
}

func assertPrice(t *testing.T, expected *provider.Price, actual *messages.Price) {
	p, _ := new(big.Float).SetInt(actual.Price.Val).Float64()
	assert.Equal(t, actual.Price.Age.Unix(), expected.Time.Unix())
	assert.Equal(t, actual.Price.Wat, expected.Pair.Base+expected.Pair.Quote)
	assert.Equal(t, p/median.PriceMultiplier, expected.Price)
	assert.Equal(t, actual.Price.Sig.V, big.NewInt(0xAA))
	assert.Equal(t, actual.Price.Sig.R.Bytes(), hexutil.MustHexToBytes("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	assert.Equal(t, actual.Price.Sig.S.Bytes(), hexutil.MustHexToBytes("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
}
