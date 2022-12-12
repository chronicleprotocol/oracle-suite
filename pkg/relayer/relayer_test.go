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

package relayer

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	oracleMocks "github.com/chronicleprotocol/oracle-suite/pkg/price/oracle/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	storeMocks "github.com/chronicleprotocol/oracle-suite/pkg/price/store/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store/testutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestRelayer_relay(t *testing.T) {
	// Test prices:
	address := ethereum.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	priceAAABBB1 := &messages.Price{
		Price: &oracle.Price{
			Wat: "AAABBB",
			Val: big.NewInt(9),
			Age: time.Now(),
			V:   1,
			R:   [32]byte{1},
			S:   [32]byte{2},
		},
		Trace: nil,
	}
	priceAAABBB2 := &messages.Price{
		Price: &oracle.Price{
			Wat: "AAABBB",
			Val: big.NewInt(10),
			Age: time.Now(),
			V:   1,
			R:   [32]byte{1},
			S:   [32]byte{2},
		},
		Trace: nil,
	}
	priceAAABBB3 := &messages.Price{
		Price: &oracle.Price{
			Wat: "AAABBB",
			Val: big.NewInt(11),
			Age: time.Now(),
			V:   1,
			R:   [32]byte{1},
			S:   [32]byte{2},
		},
		Trace: nil,
	}

	// Services:
	sig := &ethereumMocks.Signer{}
	tra := local.New([]byte("test"), 0, map[string]transport.Message{
		messages.PriceV1MessageName: &messages.Price{},
	})
	med := &oracleMocks.Median{}
	sto := &storeMocks.Storage{}
	pri, err := store.New(store.Config{
		Storage:   sto,
		Signer:    sig,
		Transport: tra,
		Pairs:     []string{"AAABBB"},
		Logger:    null.New(),
	})
	require.NoError(t, err)
	rel, err := New(Config{
		Signer:     sig,
		PriceStore: pri,
		Interval:   time.Second,
		Pairs: []*Pair{{
			AssetPair:        "AAABBB",
			OracleSpread:     1.0,
			OracleExpiration: 10 * time.Second,
			Median:           med,
		}},
		Logger: null.New(),
	})
	require.NoError(t, err)

	// Start services:
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*5)
	_ = tra.Start(ctx)
	_ = pri.Start(ctx)
	_ = rel.Start(ctx)
	defer func() {
		ctxCancel()
		<-tra.Wait()
		<-pri.Wait()
		<-rel.Wait()
	}()

	// Prepare mocks:
	sto.On("GetByAssetPair", ctx, "AAABBB").Return(
		[]*messages.Price{priceAAABBB1, priceAAABBB2, priceAAABBB3},
		nil,
	)
	med.On("Bar", ctx).Return(int64(3), nil)
	med.On("Age", ctx).Return(time.Now().Add(-30*time.Second), nil)
	med.On("Val", ctx).Return(big.NewInt(10), nil)
	sig.On("Recover", mock.Anything, mock.Anything).Return(&address, nil)
	med.On("Poke",
		ctx,
		[]*oracle.Price{priceAAABBB1.Price, priceAAABBB2.Price, priceAAABBB3.Price},
		true,
	).Return(
		&ethereum.Hash{},
		nil,
	)
}

func Test_oraclePrices(t *testing.T) {
	ms := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	}
	ps := messagesToPrices(&ms)
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
		{
			price: 0,
			want:  math.Inf(1),
		},
		{
			price: 20,
			want:  25,
		},
		{
			price: 25,
			want:  0,
		},
		{
			price: 50,
			want:  50,
		},
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
	ps := messagesToPrices(&ms)
	assert.Len(t, ps, 2)
	assert.Contains(t, ps, testutil.PriceAAABBB3.Price)
	assert.Contains(t, ps, testutil.PriceAAABBB4.Price)
}
