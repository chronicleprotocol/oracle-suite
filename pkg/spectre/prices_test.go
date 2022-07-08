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

package spectre

import (
	"math"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/store/testutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestPrices_len(t *testing.T) {
	ps := newPricesList([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	assert.Equal(t, 4, ps.len())
}

func TestPrices_messages(t *testing.T) {
	ps := newPricesList([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	assert.Len(t, ps.messages(), 4)
	assert.Contains(t, ps.messages(), testutil.PriceAAABBB1)
	assert.Contains(t, ps.messages(), testutil.PriceAAABBB2)
	assert.Contains(t, ps.messages(), testutil.PriceAAABBB3)
	assert.Contains(t, ps.messages(), testutil.PriceAAABBB4)
}

func TestPrices_oraclePrices(t *testing.T) {
	ps := newPricesList([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	assert.Len(t, ps.oraclePrices(), 4)
	assert.Contains(t, ps.oraclePrices(), testutil.PriceAAABBB1.Price)
	assert.Contains(t, ps.oraclePrices(), testutil.PriceAAABBB2.Price)
	assert.Contains(t, ps.oraclePrices(), testutil.PriceAAABBB3.Price)
	assert.Contains(t, ps.oraclePrices(), testutil.PriceAAABBB4.Price)
}

func TestPrices_truncate(t *testing.T) {
	msgs := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	}

	ps1 := newPricesList(msgs)
	ps1.truncate(5)
	assert.Len(t, ps1.messages(), 4)

	ps2 := newPricesList(msgs)
	ps2.truncate(4)
	assert.Len(t, ps2.messages(), 4)

	ps3 := newPricesList(msgs)
	ps3.truncate(3)
	assert.Len(t, ps3.messages(), 3)
}

func TestPrices_median_Even(t *testing.T) {
	ps := newPricesList([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	assert.Equal(t, big.NewInt(25), ps.median())
}

func TestPrices_Median_Odd(t *testing.T) {
	ps := newPricesList([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
	})

	assert.Equal(t, big.NewInt(20), ps.median())
}

func TestPrices_Median_Empty(t *testing.T) {
	ps := newPricesList([]*messages.Price{})

	assert.Equal(t, big.NewInt(0), ps.median())
}

func TestPrices_spread(t *testing.T) {
	ps := newPricesList([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

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
		t.Run("Case:"+strconv.Itoa(n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, ps.spread(big.NewInt(tt.price)))
		})
	}
}

func TestPrices_clearOlderThan(t *testing.T) {
	ps := newPricesList([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	ps.clearOlderThan(time.Unix(300, 0))

	assert.Len(t, ps.oraclePrices(), 2)
	assert.Contains(t, ps.oraclePrices(), testutil.PriceAAABBB3.Price)
	assert.Contains(t, ps.oraclePrices(), testutil.PriceAAABBB4.Price)
}
