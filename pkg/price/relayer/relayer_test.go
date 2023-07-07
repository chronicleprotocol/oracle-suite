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

package relayer

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	logMocks "github.com/chronicleprotocol/oracle-suite/pkg/log/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	priceMedian "github.com/chronicleprotocol/oracle-suite/pkg/price/median"
	medianMocks "github.com/chronicleprotocol/oracle-suite/pkg/price/median/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	storeMocks "github.com/chronicleprotocol/oracle-suite/pkg/price/store/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store/testutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

var (
	feedAddress  = types.MustAddressFromHex("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	priceAAABBB1 = &messages.Price{
		Price: &priceMedian.Price{
			Wat: "AAABBB",
			Val: big.NewInt(9),
			Age: time.Now(),
			Sig: types.Signature{
				V: big.NewInt(1),
				R: big.NewInt(1),
				S: big.NewInt(2),
			},
		},
		Trace: nil,
	}
	priceAAABBB2 = &messages.Price{
		Price: &priceMedian.Price{
			Wat: "AAABBB",
			Val: big.NewInt(10),
			Age: time.Now(),
			Sig: types.Signature{
				V: big.NewInt(1),
				R: big.NewInt(1),
				S: big.NewInt(2),
			},
		},
		Trace: nil,
	}
	priceAAABBB3 = &messages.Price{
		Price: &priceMedian.Price{
			Wat: "AAABBB",
			Val: big.NewInt(11),
			Age: time.Now(),
			Sig: types.Signature{
				V: big.NewInt(1),
				R: big.NewInt(1),
				S: big.NewInt(2),
			},
		},
		Trace: nil,
	}
)

func TestRelayer_relay(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(ctx context.Context, priceStorage *storeMocks.Storage, median *medianMocks.Median, recoverer *ethereumMocks.Recoverer, log *logMocks.Logger)
		asserts func(t *testing.T, priceStorage *storeMocks.Storage, median *medianMocks.Median, recoverer *ethereumMocks.Recoverer, log *logMocks.Logger)
	}{
		{
			name: "single-price",
			mocks: func(ctx context.Context, priceStorage *storeMocks.Storage, median *medianMocks.Median, recoverer *ethereumMocks.Recoverer, log *logMocks.Logger) {
				priceStorage.On("GetByAssetPair", ctx, "AAABBB").Return([]*messages.Price{priceAAABBB1}, nil)
				recoverer.On("RecoverMessage", mock.Anything, mock.Anything).Return(&feedAddress, nil)
				median.On("Feeds", ctx).Return([]types.Address{feedAddress}, nil)
				median.On("Bar", ctx).Return(int64(1), nil)
				median.On("Age", ctx).Return(time.Now().Add(-30*time.Second), nil)
				median.On("Val", ctx).Return(big.NewInt(10), nil)
				median.On("Poke", ctx, []*priceMedian.Price{priceAAABBB1.Price}, true).Return(&types.Hash{}, nil)
			},
		},
		{
			name: "three-prices",
			mocks: func(ctx context.Context, priceStorage *storeMocks.Storage, median *medianMocks.Median, recoverer *ethereumMocks.Recoverer, log *logMocks.Logger) {
				priceStorage.On("GetByAssetPair", ctx, "AAABBB").Return([]*messages.Price{priceAAABBB1, priceAAABBB2, priceAAABBB3}, nil)
				recoverer.On("RecoverMessage", mock.Anything, mock.Anything).Return(&feedAddress, nil)
				median.On("Feeds", ctx).Return([]types.Address{feedAddress}, nil)
				median.On("Bar", ctx).Return(int64(3), nil)
				median.On("Age", ctx).Return(time.Now().Add(-30*time.Second), nil)
				median.On("Val", ctx).Return(big.NewInt(10), nil)
				median.On("Poke", ctx, []*priceMedian.Price{priceAAABBB1.Price, priceAAABBB2.Price, priceAAABBB3.Price}, true).Return(&types.Hash{}, nil)
			},
		},
		{
			name: "spread-too-low",
			mocks: func(ctx context.Context, priceStorage *storeMocks.Storage, median *medianMocks.Median, recoverer *ethereumMocks.Recoverer, log *logMocks.Logger) {
				priceStorage.On("GetByAssetPair", ctx, "AAABBB").Return([]*messages.Price{priceAAABBB2}, nil)
				recoverer.On("RecoverMessage", mock.Anything, mock.Anything).Return(&feedAddress, nil)
				median.On("Feeds", ctx).Return([]types.Address{feedAddress}, nil)
				median.On("Bar", ctx).Return(int64(1), nil)
				median.On("Age", ctx).Return(time.Now().Add(-5*time.Second), nil)
				median.On("Val", ctx).Return(big.NewInt(10), nil)
			},
		},
		{
			name: "unknown-feed",
			mocks: func(ctx context.Context, priceStorage *storeMocks.Storage, median *medianMocks.Median, recoverer *ethereumMocks.Recoverer, log *logMocks.Logger) {
				priceStorage.On("GetByAssetPair", ctx, "AAABBB").Return([]*messages.Price{priceAAABBB1}, nil)
				recoverer.On("RecoverMessage", mock.Anything, mock.Anything).Return(&feedAddress, nil)
				median.On("Feeds", ctx).Return([]types.Address{}, nil)
				median.On("Bar", ctx).Return(int64(1), nil)
				median.On("Age", ctx).Return(time.Now().Add(-30*time.Second), nil)
				median.On("Val", ctx).Return(big.NewInt(10), nil)
			},
		},
		{
			name: "not-enough-prices",
			mocks: func(ctx context.Context, priceStorage *storeMocks.Storage, median *medianMocks.Median, recoverer *ethereumMocks.Recoverer, log *logMocks.Logger) {
				priceStorage.On("GetByAssetPair", ctx, "AAABBB").Return([]*messages.Price{priceAAABBB1, priceAAABBB2}, nil)
				recoverer.On("RecoverMessage", mock.Anything, mock.Anything).Return(&feedAddress, nil)
				median.On("Feeds", ctx).Return([]types.Address{feedAddress}, nil)
				median.On("Bar", ctx).Return(int64(3), nil)
				median.On("Age", ctx).Return(time.Now().Add(-30*time.Second), nil)
				median.On("Val", ctx).Return(big.NewInt(10), nil)
			},
			asserts: func(t *testing.T, priceStorage *storeMocks.Storage, median *medianMocks.Median, recoverer *ethereumMocks.Recoverer, log *logMocks.Logger) {
				hasErr := false
				for _, c := range log.Mock().Calls {
					if c.Method == "WithError" && strings.Contains(c.Arguments[0].(error).Error(), "not enough prices to achieve quorum") {
						hasErr = true
						break
					}
				}
				assert.True(t, hasErr)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*5)
			defer ctxCancel()

			// Prepare relayer dependencies.
			pokeTicker := timeutil.NewTicker(0)
			addressesTicker := timeutil.NewTicker(0)
			recovererMock := &ethereumMocks.Recoverer{}
			medianMock := &medianMocks.Median{}
			if tt.name != "unknown-feed" {
				medianMock.On("Address").Return(feedAddress)
			}
			storageMock := &storeMocks.Storage{}
			localTransport := local.New([]byte("test"), 0, map[string]transport.Message{
				messages.PriceV1MessageName: &messages.Price{},
			})
			mockLogger := logMocks.New()
			priceStore, err := store.New(store.Config{
				Storage:   storageMock,
				Transport: localTransport,
				Pairs:     []string{"AAABBB"},
				Logger:    null.New(),
				Recoverer: recovererMock,
			})
			require.NoError(t, err)

			// Prepare mocks.
			mockLogger.Mock().On("WithError", mock.Anything).Return(mockLogger)
			mockLogger.Mock().On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
			mockLogger.Mock().On("WithFields", mock.Anything).Return(mockLogger)
			mockLogger.Mock().On("Warn", mock.Anything).Return()
			mockLogger.Mock().On("Info", mock.Anything)
			mockLogger.Mock().On("Debug", mock.Anything)
			tt.mocks(ctx, storageMock, medianMock, recovererMock, mockLogger)

			// Prepare relayer.
			relayer, err := New(Config{
				PriceStore: priceStore,
				PokeTicker: pokeTicker,
				Pairs: []*Pair{{
					AssetPair:                 "AAABBB",
					Spread:                    1.0,
					Expiration:                10 * time.Second,
					Median:                    medianMock,
					FeedAddressesUpdateTicker: addressesTicker,
				}},
				Logger:    mockLogger,
				Recoverer: recovererMock,
			})
			require.NoError(t, err)

			// Start relayer.
			_ = localTransport.Start(ctx)
			_ = priceStore.Start(ctx)
			_ = relayer.Start(ctx)
			defer func() {
				ctxCancel()
				<-localTransport.Wait()
				<-priceStore.Wait()
				<-relayer.Wait()
			}()

			pokeTicker.Tick()
			addressesTicker.Tick()

			assert.Eventually(t, func() bool {
				for _, m := range mockLogger.Mock().Calls {
					if m.Method == "Warn" && m.Arguments[0].([]any)[0] == "Unable to update Oracle" {
						return true
					}
					if m.Method == "Info" && m.Arguments[0].([]any)[0] == "Oracle price is still valid" {
						return true
					}
					if m.Method == "Info" && m.Arguments[0].([]any)[0] == "Oracle updated" {
						return true
					}
				}
				return false
			}, time.Second*5, time.Millisecond*100)

			storageMock.AssertExpectations(t)
			medianMock.AssertExpectations(t)

			if tt.asserts != nil {
				tt.asserts(t, storageMock, medianMock, recovererMock, mockLogger)
			}
		})
	}
}

func Test_oraclePrices(t *testing.T) {
	ms := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	}
	ps := toOraclePrices(&ms)
	assert.Len(t, ps, 4)
	assert.Contains(t, ps, testutil.PriceAAABBB1.Price)
	assert.Contains(t, ps, testutil.PriceAAABBB2.Price)
	assert.Contains(t, ps, testutil.PriceAAABBB3.Price)
	assert.Contains(t, ps, testutil.PriceAAABBB4.Price)
}

func Test_truncate(t *testing.T) {
	ms := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	}
	truncate(&ms, 2)
	assert.Len(t, ms, 2)
}

func Test_median_Even(t *testing.T) {
	ms := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	}
	assert.Equal(t, big.NewInt(25), calcMedian(&ms))
}

func Test_Median_Odd(t *testing.T) {
	ms := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
	}
	assert.Equal(t, big.NewInt(20), calcMedian(&ms))
}

func Test_Median_Empty(t *testing.T) {
	var ms []*messages.Price
	assert.Equal(t, big.NewInt(0), calcMedian(&ms))
}

func Test_spread(t *testing.T) {
	ms := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	}
	tests := []struct {
		price int64
		want  float64
	}{
		{price: 0, want: math.Inf(1)},
		{price: 20, want: 25},
		{price: 25, want: 0},
		{price: 50, want: 50},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, calcSpread(&ms, big.NewInt(tt.price)))
		})
	}
}

func Test_clearOlderThan(t *testing.T) {
	ms := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	}
	clearOlderThan(&ms, time.Unix(300, 0))
	ps := toOraclePrices(&ms)
	assert.Len(t, ps, 2)
	assert.Contains(t, ps, testutil.PriceAAABBB3.Price)
	assert.Contains(t, ps, testutil.PriceAAABBB4.Price)
}
