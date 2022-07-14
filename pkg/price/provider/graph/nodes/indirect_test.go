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

const testTTL = 10 * time.Second

func TestIndirectAggregatorNode_Children(t *testing.T) {
	m := NewIndirectAggregatorNode(provider.Pair{Base: "A", Quote: "B"})

	c1 := NewOriginNode(OriginPair{Pair: provider.Pair{Base: "A", Quote: "B"}, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: provider.Pair{Base: "A", Quote: "B"}, Origin: "b"}, testTTL, testTTL)
	c3 := NewOriginNode(OriginPair{Pair: provider.Pair{Base: "A", Quote: "B"}, Origin: "c"}, testTTL, testTTL)

	m.AddChild(c1)
	m.AddChild(c2)
	m.AddChild(c3)

	assert.Len(t, m.Children(), 3)
	assert.Same(t, c1, m.Children()[0])
	assert.Same(t, c2, m.Children()[1])
	assert.Same(t, c3, m.Children()[2])
}

func TestIndirectAggregatorNode_Pair(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	m := NewIndirectAggregatorNode(p)

	assert.Equal(t, m.Pair(), p)
}

func TestIndirectAggregatorNode_Price_ThreeOriginPrices(t *testing.T) {
	p1 := provider.Pair{Base: "A", Quote: "B"}
	p2 := provider.Pair{Base: "B", Quote: "C"}
	p3 := provider.Pair{Base: "C", Quote: "D"}
	pf := provider.Pair{Base: "A", Quote: "D"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)
	c3 := NewOriginNode(OriginPair{Pair: p3, Origin: "c"}, testTTL, testTTL)

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

	_ = c3.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p3,
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
			Pair:      pf,
			Price:     6000,
			Bid:       6000,
			Ask:       6000,
			Volume24h: 0,
			Time:      n,
		},
		OriginPrices:     []OriginPrice{c1.Price(), c2.Price(), c3.Price()},
		AggregatorPrices: nil,
		Parameters:       map[string]string{"method": "indirect"},
		Error:            nil,
	}

	assert.Equal(t, expected, m.Price())
}

func TestIndirectAggregatorNode_Price_ThreeAggregatorPrices(t *testing.T) {
	p1 := provider.Pair{Base: "A", Quote: "B"}
	p2 := provider.Pair{Base: "B", Quote: "C"}
	p3 := provider.Pair{Base: "C", Quote: "D"}
	pf := provider.Pair{Base: "A", Quote: "D"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)
	c3 := NewOriginNode(OriginPair{Pair: p3, Origin: "c"}, testTTL, testTTL)

	i1 := NewIndirectAggregatorNode(p1)
	i1.AddChild(c1)

	i2 := NewIndirectAggregatorNode(p2)
	i2.AddChild(c2)

	i3 := NewIndirectAggregatorNode(p3)
	i3.AddChild(c3)

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

	_ = c3.Ingest(OriginPrice{
		PairPrice: PairPrice{
			Pair:      p3,
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
			Pair:      pf,
			Price:     6000,
			Bid:       6000,
			Ask:       6000,
			Volume24h: 0,
			Time:      n,
		},
		OriginPrices: nil,
		AggregatorPrices: []AggregatorPrice{
			{
				PairPrice: PairPrice{
					Pair:      p1,
					Price:     10,
					Bid:       10,
					Ask:       10,
					Volume24h: 10,
					Time:      n,
				},
				OriginPrices:     []OriginPrice{c1.Price()},
				AggregatorPrices: nil,
				Parameters:       map[string]string{"method": "indirect"},
				Error:            nil,
			},
			{
				PairPrice: PairPrice{
					Pair:      p2,
					Price:     20,
					Bid:       20,
					Ask:       20,
					Volume24h: 20,
					Time:      n,
				},
				OriginPrices:     []OriginPrice{c2.Price()},
				AggregatorPrices: nil,
				Parameters:       map[string]string{"method": "indirect"},
				Error:            nil,
			},
			{
				PairPrice: PairPrice{
					Pair:      p3,
					Price:     30,
					Bid:       30,
					Ask:       30,
					Volume24h: 30,
					Time:      n,
				},
				OriginPrices:     []OriginPrice{c3.Price()},
				AggregatorPrices: nil,
				Parameters:       map[string]string{"method": "indirect"},
				Error:            nil,
			},
		},
		Parameters: map[string]string{"method": "indirect"},
		Error:      nil,
	}

	assert.Equal(t, expected, m.Price())
}

func TestIndirectAggregatorNode_Price_ChildPriceWithError(t *testing.T) {
	p1 := provider.Pair{Base: "A", Quote: "B"}
	p2 := provider.Pair{Base: "B", Quote: "C"}
	pf := provider.Pair{Base: "A", Quote: "C"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)

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
		Error:  errors.New("something"),
	})

	m.AddChild(c1)
	m.AddChild(c2)

	assert.True(t, errors.As(m.Price().Error, &ErrPrice{}))
}

func TestIndirectAggregatorNode_Price_ResolveToWrongPair(t *testing.T) {
	// Below pairs will be resolved resolve to the A/D but the A/C is expected:
	p1 := provider.Pair{Base: "A", Quote: "B"}
	p2 := provider.Pair{Base: "B", Quote: "D"}
	pf := provider.Pair{Base: "A", Quote: "C"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)

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

	assert.True(t, errors.As(m.Price().Error, &ErrResolve{}))
}

func TestIndirectAggregatorNode_Price_UnableToResolve(t *testing.T) {
	// It's impossible to resolve below pairs, because the A/B and C/D have no common part:
	p1 := provider.Pair{Base: "A", Quote: "B"}
	p2 := provider.Pair{Base: "C", Quote: "D"}
	pf := provider.Pair{Base: "A", Quote: "C"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)

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

	// Resolved to the A/D pair but the A/C pair was expected:
	assert.True(t, errors.As(m.Price().Error, &ErrNoCommonPart{}))
}

func TestIndirectAggregatorNode_Price_DivByZero(t *testing.T) {
	p1 := provider.Pair{Base: "A", Quote: "B"}
	p2 := provider.Pair{Base: "C", Quote: "B"}
	pf := provider.Pair{Base: "A", Quote: "C"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)

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
			Price:     0,
			Bid:       0,
			Ask:       0,
			Volume24h: 0,
			Time:      n,
		},
		Origin: "b",
		Error:  nil,
	})

	m.AddChild(c1)
	m.AddChild(c2)

	// Conversion between the A/B and the C/B requires division. The C/B pair
	// reports 0 as its price so the ErrDivByZero error should be returned:
	assert.True(t, errors.As(m.Price().Error, &ErrDivByZero{}))
}

func Test_crossRate(t *testing.T) {
	tests := []struct {
		name    string
		prices  []PairPrice
		want    PairPrice
		wantErr bool
	}{
		{
			name:    "no-pair",
			prices:  []PairPrice{},
			want:    PairPrice{},
			wantErr: false,
		},
		{
			name: "one-pair",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "A", Quote: "B"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
			},
			want: PairPrice{
				Pair:      provider.Pair{Base: "A", Quote: "B"},
				Price:     10,
				Bid:       5,
				Ask:       15,
				Volume24h: 10,
				Time:      time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "invalid-conversion",
			prices: []PairPrice{
				{Pair: provider.Pair{Base: "A", Quote: "B"}},
				{Pair: provider.Pair{Base: "X", Quote: "Y"}},
			},
			want:    PairPrice{},
			wantErr: true,
		},
		{
			name: "ac/bc",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "A", Quote: "C"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "B", Quote: "C"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
			},
			want: PairPrice{
				Pair:      provider.Pair{Base: "A", Quote: "C"},
				Price:     float64(10) / 20,
				Bid:       float64(5) / 10,
				Ask:       float64(15) / 30,
				Volume24h: 10,
				Time:      time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "ac/bc-divByZero",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "A", Quote: "C"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "B", Quote: "C"},
					Price:     0,
					Bid:       0,
					Ask:       0,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
			},
			want:    PairPrice{},
			wantErr: true,
		},
		{
			name: "ca/cb",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "C", Quote: "A"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "C", Quote: "B"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
			},
			want: PairPrice{
				Pair:      provider.Pair{Base: "A", Quote: "C"},
				Price:     float64(20) / 10,
				Bid:       float64(10) / 5,
				Ask:       float64(30) / 15,
				Volume24h: 10,
				Time:      time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "ca/cb-divByZero",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "C", Quote: "A"},
					Price:     0,
					Bid:       0,
					Ask:       0,
					Volume24h: 0,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "C", Quote: "B"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
			},
			want:    PairPrice{},
			wantErr: true,
		},
		{
			name: "ac/cb",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "A", Quote: "C"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "C", Quote: "B"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
			},
			want: PairPrice{
				Pair:      provider.Pair{Base: "A", Quote: "C"},
				Price:     float64(20) * 10,
				Bid:       float64(10) * 5,
				Ask:       float64(30) * 15,
				Volume24h: 10,
				Time:      time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "ca/bc",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "C", Quote: "A"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "B", Quote: "C"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
			},
			want: PairPrice{
				Pair:      provider.Pair{Base: "A", Quote: "C"},
				Price:     float64(1) / 20 / 10,
				Bid:       float64(1) / 10 / 5,
				Ask:       float64(1) / 30 / 15,
				Volume24h: 10,
				Time:      time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "ca/bc-divByZero1",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "C", Quote: "A"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "B", Quote: "C"},
					Price:     0,
					Bid:       0,
					Ask:       0,
					Volume24h: 0,
					Time:      time.Unix(10, 0),
				},
			},
			want:    PairPrice{},
			wantErr: true,
		},
		{
			name: "ca/bc-divByZero2",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "C", Quote: "A"},
					Price:     0,
					Bid:       0,
					Ask:       0,
					Volume24h: 0,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "B", Quote: "C"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
			},
			want:    PairPrice{},
			wantErr: true,
		},
		{
			name: "lowest-timestamp",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "A", Quote: "B"},
					Price:     1,
					Bid:       1,
					Ask:       1,
					Volume24h: 0,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "B", Quote: "C"},
					Price:     1,
					Bid:       1,
					Ask:       1,
					Volume24h: 0,
					Time:      time.Unix(5, 0),
				},
				{
					Pair:      provider.Pair{Base: "C", Quote: "D"},
					Price:     1,
					Bid:       1,
					Ask:       1,
					Volume24h: 0,
					Time:      time.Unix(15, 0),
				},
			},
			want: PairPrice{
				Pair:      provider.Pair{Base: "A", Quote: "D"},
				Price:     1,
				Bid:       1,
				Ask:       1,
				Volume24h: 0,
				Time:      time.Unix(5, 0),
			},
			wantErr: false,
		},
		{
			name: "five-pairs",
			prices: []PairPrice{
				{
					Pair:      provider.Pair{Base: "ETH", Quote: "BTC"}, // -> ETH/BTC
					Price:     0.050,
					Bid:       0.040,
					Ask:       0.060,
					Volume24h: 10,
					Time:      time.Unix(10, 0),
				},
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"}, // -> ETH/USD
					Price:     10000.000,
					Bid:       9000.000,
					Ask:       11000.000,
					Volume24h: 15,
					Time:      time.Unix(9, 0),
				},
				{
					Pair:      provider.Pair{Base: "EUR", Quote: "USD"}, // -> ETH/EUR
					Price:     1.250,
					Bid:       1.200,
					Ask:       1.300,
					Volume24h: 20,
					Time:      time.Unix(11, 0),
				},
				{
					Pair:      provider.Pair{Base: "EUR", Quote: "CAD"}, // -> ETH/CAD
					Price:     1.250,
					Bid:       1.200,
					Ask:       1.300,
					Volume24h: 25,
					Time:      time.Unix(8, 0),
				},
				{
					Pair:      provider.Pair{Base: "GPB", Quote: "ETH"}, // -> GPB/CAD
					Price:     0.005,
					Bid:       0.004,
					Ask:       0.006,
					Volume24h: 30,
					Time:      time.Unix(13, 0),
				},
			},
			want: PairPrice{
				Pair:      provider.Pair{Base: "GPB", Quote: "CAD"},
				Price:     float64(1) / (((float64(0.050) * 10000.000) / 1.250) * 1.250) / 0.005,
				Bid:       float64(1) / (((float64(0.040) * 9000.000) / 1.200) * 1.200) / 0.004,
				Ask:       float64(1) / (((float64(0.060) * 11000.000) / 1.300) * 1.300) / 0.006,
				Volume24h: 0,
				Time:      time.Unix(8, 0),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crossRate(tt.prices)

			if err != nil {
				if tt.wantErr {
					assert.Error(t, err)
				}

				return
			}

			assert.InDelta(t, tt.want.Price, got.Price, 0.000000001)
			assert.InDelta(t, tt.want.Bid, got.Bid, 0.000000001)
			assert.InDelta(t, tt.want.Ask, got.Ask, 0.000000001)
			assert.Equal(t, tt.want.Time.Unix(), got.Time.Unix())
			assert.Nil(t, err)
		})
	}
}
