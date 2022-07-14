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

package nodes

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

const medianTestTTL = 10 * time.Second

func TestMedianAggregatorNode_Children(t *testing.T) {
	m := NewMedianAggregatorNode(provider.Pair{Base: "A", Quote: "B"}, 3)

	c1 := NewOriginNode(OriginPair{Pair: provider.Pair{Base: "A", Quote: "B"}, Origin: "a"}, medianTestTTL, medianTestTTL)
	c2 := NewOriginNode(OriginPair{Pair: provider.Pair{Base: "A", Quote: "B"}, Origin: "b"}, medianTestTTL, medianTestTTL)
	c3 := NewOriginNode(OriginPair{Pair: provider.Pair{Base: "A", Quote: "B"}, Origin: "c"}, medianTestTTL, medianTestTTL)

	m.AddChild(c1)
	m.AddChild(c2)
	m.AddChild(c3)

	assert.Len(t, m.Children(), 3)
	assert.Same(t, c1, m.Children()[0])
	assert.Same(t, c2, m.Children()[1])
	assert.Same(t, c3, m.Children()[2])
}

func TestMedianAggregatorNode_Pair(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	m := NewMedianAggregatorNode(p, 3)

	assert.Equal(t, m.Pair(), p)
}

func TestMedianAggregatorNode_Price_ThreeOriginPrices(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 3)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"}, medianTestTTL, medianTestTTL)
	c2 := NewOriginNode(OriginPair{Pair: p, Origin: "b"}, medianTestTTL, medianTestTTL)
	c3 := NewOriginNode(OriginPair{Pair: p, Origin: "c"}, medianTestTTL, medianTestTTL)

	_ = c1.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Time:      n,
		},
		Origin: "a",
		Error:  nil,
	})

	_ = c2.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Time:      n,
		},
		Origin: "b",
		Error:  nil,
	})

	_ = c3.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     30,
			Bid:       30,
			Ask:       30,
			Volume24h: 30,
			Time:      n,
		},
		Origin: "c",
		Error:  nil,
	})

	m.AddChild(c1)
	m.AddChild(c2)
	m.AddChild(c3)

	expected := AggregatorPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 0,
			Time:      n,
		},
		OriginPrices:     []OriginPrice{c1.Price(), c2.Price(), c3.Price()},
		AggregatorPrices: nil,
		Parameters:       map[string]string{"method": "median", "minimumSuccessfulSources": "3"},
		Error:            nil,
	}

	assert.Equal(t, expected, m.Price())
}

func TestMedianAggregatorNode_Price_ThreeAggregatorPrices(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 3)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"}, medianTestTTL, medianTestTTL)
	c2 := NewOriginNode(OriginPair{Pair: p, Origin: "b"}, medianTestTTL, medianTestTTL)
	c3 := NewOriginNode(OriginPair{Pair: p, Origin: "c"}, medianTestTTL, medianTestTTL)

	i1 := NewMedianAggregatorNode(p, 1)
	i1.AddChild(c1)

	i2 := NewMedianAggregatorNode(p, 1)
	i2.AddChild(c2)

	i3 := NewMedianAggregatorNode(p, 1)
	i3.AddChild(c3)

	_ = c1.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Time:      n,
		},
		Origin: "a",
		Error:  nil,
	})

	_ = c2.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Time:      n,
		},
		Origin: "b",
		Error:  nil,
	})

	_ = c3.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     30,
			Bid:       30,
			Ask:       30,
			Volume24h: 30,
			Time:      n,
		},
		Origin: "c",
		Error:  nil,
	})

	m.AddChild(i1)
	m.AddChild(i2)
	m.AddChild(i3)

	expected := AggregatorPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 0,
			Time:      n,
		},
		OriginPrices: nil,
		AggregatorPrices: []AggregatorPrice{
			{
				PairPrice: PairPrice{
					Pair:      p,
					Price:     10,
					Bid:       10,
					Ask:       10,
					Volume24h: 0,
					Time:      n,
				},
				OriginPrices:     []OriginPrice{c1.Price()},
				AggregatorPrices: nil,
				Parameters:       map[string]string{"method": "median", "minimumSuccessfulSources": "1"},
				Error:            nil,
			},
			{
				PairPrice: PairPrice{
					Pair:      p,
					Price:     20,
					Bid:       20,
					Ask:       20,
					Volume24h: 0,
					Time:      n,
				},
				OriginPrices:     []OriginPrice{c2.Price()},
				AggregatorPrices: nil,
				Parameters:       map[string]string{"method": "median", "minimumSuccessfulSources": "1"},
				Error:            nil,
			},
			{
				PairPrice: PairPrice{
					Pair:      p,
					Price:     30,
					Bid:       30,
					Ask:       30,
					Volume24h: 0,
					Time:      n,
				},
				OriginPrices:     []OriginPrice{c3.Price()},
				AggregatorPrices: nil,
				Parameters:       map[string]string{"method": "median", "minimumSuccessfulSources": "1"},
				Error:            nil,
			},
		},
		Parameters: map[string]string{"method": "median", "minimumSuccessfulSources": "3"},
		Error:      nil,
	}

	assert.Equal(t, expected, m.Price())
}

func TestMedianAggregatorNode_Price_NotEnoughSources(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 3)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"}, medianTestTTL, medianTestTTL)

	_ = c1.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Time:      n,
		},
		Origin: "a",
		Error:  nil,
	})

	m.AddChild(c1)

	price := m.Price()

	assert.True(t, errors.As(price.Error, &ErrNotEnoughSources{}))

	// If possible, the median should be calculated for the rest of the prices:
	assert.Equal(t, float64(10), price.Price)
	assert.Equal(t, float64(10), price.Bid)
	assert.Equal(t, float64(10), price.Ask)
}

func TestMedianAggregatorNode_Price_ChildPriceWithError(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 2)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"}, medianTestTTL, medianTestTTL)
	c2 := NewOriginNode(OriginPair{Pair: p, Origin: "b"}, medianTestTTL, medianTestTTL)

	_ = c1.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Time:      n,
		},
		Origin: "a",
		Error:  nil,
	})

	_ = c2.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Time:      n,
		},
		Origin: "b",
		Error:  errors.New("something"),
	})

	m.AddChild(c1)
	m.AddChild(c2)

	price := m.Price()

	assert.True(t, errors.As(price.Error, &ErrNotEnoughSources{}))

	// If possible, the median should be calculated for the rest of the prices:
	assert.Equal(t, float64(10), price.Price)
	assert.Equal(t, float64(10), price.Bid)
	assert.Equal(t, float64(10), price.Ask)
}

func TestMedianAggregatorNode_Price_IncompatiblePairs(t *testing.T) {
	p1 := provider.Pair{Base: "A", Quote: "B"}
	p2 := provider.Pair{Base: "C", Quote: "D"}
	n := time.Now()
	m := NewMedianAggregatorNode(p1, 2)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, medianTestTTL, medianTestTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, medianTestTTL, medianTestTTL)

	_ = c1.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p1,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Time:      n,
		},
		Origin: "a",
		Error:  nil,
	})

	_ = c2.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p2,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Time:      n,
		},
		Origin: "b",
		Error:  nil,
	})

	m.AddChild(c1)
	m.AddChild(c2)

	price := m.Price()

	assert.True(t, errors.As(price.Error, &ErrIncompatiblePairs{}))

	// If possible, the median should be calculated for the rest of the prices:
	assert.Equal(t, float64(10), price.Price)
	assert.Equal(t, float64(10), price.Bid)
	assert.Equal(t, float64(10), price.Ask)
}

func TestMedianAggregatorNode_Price_NoChildrenNodes(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	m := NewMedianAggregatorNode(p, 2)

	price := m.Price()

	assert.True(t, errors.As(price.Error, &ErrNotEnoughSources{}))
}

func TestMedianAggregatorNode_Price_FilterOutPricesLteZero(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 1)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"}, medianTestTTL, medianTestTTL)
	c2 := NewOriginNode(OriginPair{Pair: p, Origin: "b"}, medianTestTTL, medianTestTTL)
	c3 := NewOriginNode(OriginPair{Pair: p, Origin: "c"}, medianTestTTL, medianTestTTL)

	_ = c1.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     10,
			Bid:       0,
			Ask:       0,
			Volume24h: 0,
			Time:      n,
		},
		Origin: "a",
		Error:  nil,
	})

	_ = c2.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     0,
			Bid:       10,
			Ask:       0,
			Volume24h: 0,
			Time:      n,
		},
		Origin: "b",
		Error:  nil,
	})

	_ = c3.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p,
			Price:     0,
			Bid:       0,
			Ask:       10,
			Volume24h: 0,
			Time:      n,
		},
		Origin: "c",
		Error:  nil,
	})

	m.AddChild(c1)
	m.AddChild(c2)
	m.AddChild(c3)

	price := m.Price()

	assert.Equal(t, float64(10), price.Price)
	assert.Equal(t, float64(10), price.Bid)
	assert.Equal(t, float64(10), price.Ask)
}

func Test_median(t *testing.T) {
	tests := []struct {
		name   string
		prices []float64
		want   float64
	}{
		{
			name:   "no-prices",
			prices: []float64{},
			want:   float64(0),
		},
		{
			name:   "one-price",
			prices: []float64{10},
			want:   float64(10),
		},
		{
			name:   "three-prices",
			prices: []float64{-20, 10, 20},
			want:   float64(10),
		},
		{
			name:   "four-prices",
			prices: []float64{10, 20, 30, 40},
			want:   float64(25),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := median(tt.prices)
			assert.Equal(t, tt.want, got)
		})
	}
}
