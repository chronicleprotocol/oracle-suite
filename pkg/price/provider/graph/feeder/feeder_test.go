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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/nodes"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/origins"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

type mockHandler struct {
	// mockedPrices is a list of mocked prices to be returned by Fetch method
	mockedPrices map[origins.Pair]origins.Price
	// fetchPairs is a list of pairs provided to Fetch method on last call
	fetchPairs []origins.Pair
	// delay is a simulated delay during price fetching
	delay time.Duration
	// updateTimestamp determines if a price's timestamp should be updated on
	// every fetch
	updateTimestamp bool
}

func (m *mockHandler) Fetch(pairs []origins.Pair) []origins.FetchResult {
	m.fetchPairs = pairs
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	var fr []origins.FetchResult
	for _, pair := range pairs {
		p := m.mockedPrices[pair]
		if m.updateTimestamp {
			p.Timestamp = time.Now()
		}
		fr = append(fr, origins.FetchResult{
			Price: p,
			Error: nil,
		})
	}
	return fr
}

func originsSetMock(prices map[string][]origins.Price) *origins.Set {
	handlers := map[string]origins.Handler{}
	for origin, prices := range prices {
		pricesMap := map[origins.Pair]origins.Price{}
		for _, price := range prices {
			pricesMap[price.Pair] = price
		}
		handlers[origin] = &mockHandler{mockedPrices: pricesMap}
	}
	return origins.NewSet(handlers)
}

func TestFeeder_Feed_EmptyGraph(t *testing.T) {
	f := NewFeeder(originsSetMock(nil), null.New())

	// Feed method shouldn't panic
	warns := f.Feed(nil, time.Now())

	assert.Len(t, warns.List, 0)
}

func TestFeeder_Feed_NoFeedableNodes(t *testing.T) {
	g := nodes.NewMedianAggregatorNode(provider.Pair{Base: "A", Quote: "B"}, 1)
	f := NewFeeder(originsSetMock(nil), null.New())

	// Feed method shouldn't panic
	warns := f.Feed([]nodes.Node{nodes.Node(g)}, time.Now())

	assert.Len(t, warns.List, 0)
}

func TestFeeder_Feed_OneOriginNode(t *testing.T) {
	s := originsSetMock(map[string][]origins.Price{
		"test": {
			origins.Price{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     10,
				Bid:       9,
				Ask:       11,
				Volume24h: 10,
				Timestamp: time.Unix(10000, 0),
			},
		},
	})

	g := nodes.NewMedianAggregatorNode(provider.Pair{Base: "A", Quote: "B"}, 1)
	o := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test",
		Pair:   provider.Pair{Base: "A", Quote: "B"},
	}, 0, 0)

	g.AddChild(o)
	f := NewFeeder(s, null.New())
	warns := f.Feed([]nodes.Node{g}, time.Now())

	assert.Len(t, warns.List, 0)
	assert.Equal(t, provider.Pair{Base: "A", Quote: "B"}, o.Price().Pair)
	assert.Equal(t, 10.0, o.Price().Price)
	assert.Equal(t, 9.0, o.Price().Bid)
	assert.Equal(t, 11.0, o.Price().Ask)
	assert.Equal(t, 10.0, o.Price().Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o.Price().Time)
}

func TestFeeder_Feed_ManyOriginNodes(t *testing.T) {
	s := originsSetMock(map[string][]origins.Price{
		"test": {
			origins.Price{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     10,
				Bid:       9,
				Ask:       11,
				Volume24h: 10,
				Timestamp: time.Unix(10000, 0),
			},
			origins.Price{
				Pair:      origins.Pair{Base: "C", Quote: "D"},
				Price:     20,
				Bid:       19,
				Ask:       21,
				Volume24h: 20,
				Timestamp: time.Unix(20000, 0),
			},
		},
		"test2": {
			origins.Price{
				Pair:      origins.Pair{Base: "E", Quote: "F"},
				Price:     30,
				Bid:       39,
				Ask:       31,
				Volume24h: 30,
				Timestamp: time.Unix(30000, 0),
			},
		},
	})

	g := nodes.NewMedianAggregatorNode(provider.Pair{Base: "A", Quote: "B"}, 1)
	o1 := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test",
		Pair:   provider.Pair{Base: "A", Quote: "B"},
	}, 0, 0)
	o2 := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test",
		Pair:   provider.Pair{Base: "C", Quote: "D"},
	}, 0, 0)
	o3 := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test2",
		Pair:   provider.Pair{Base: "E", Quote: "F"},
	}, 0, 0)
	o4 := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test2",
		Pair:   provider.Pair{Base: "E", Quote: "F"},
	}, 0, 0)

	// The last o4 origin is intentionally same as an o3 origin. Also an o3
	// origin was added two times as a child for the g node. The feeder should
	// ask for E/F pair only once.

	g.AddChild(o1)
	g.AddChild(o2)
	g.AddChild(o3)
	g.AddChild(o3) // intentionally
	g.AddChild(o4)

	f := NewFeeder(s, null.New())
	warns := f.Feed([]nodes.Node{g}, time.Now())

	assert.Len(t, warns.List, 0)

	assert.Equal(t, provider.Pair{Base: "A", Quote: "B"}, o1.Price().Pair)
	assert.Equal(t, 10.0, o1.Price().Price)
	assert.Equal(t, 9.0, o1.Price().Bid)
	assert.Equal(t, 11.0, o1.Price().Ask)
	assert.Equal(t, 10.0, o1.Price().Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o1.Price().Time)

	assert.Equal(t, provider.Pair{Base: "C", Quote: "D"}, o2.Price().Pair)
	assert.Equal(t, 20.0, o2.Price().Price)
	assert.Equal(t, 19.0, o2.Price().Bid)
	assert.Equal(t, 21.0, o2.Price().Ask)
	assert.Equal(t, 20.0, o2.Price().Volume24h)
	assert.Equal(t, time.Unix(20000, 0), o2.Price().Time)

	assert.Equal(t, provider.Pair{Base: "E", Quote: "F"}, o3.Price().Pair)
	assert.Equal(t, 30.0, o3.Price().Price)
	assert.Equal(t, 39.0, o3.Price().Bid)
	assert.Equal(t, 31.0, o3.Price().Ask)
	assert.Equal(t, 30.0, o3.Price().Volume24h)
	assert.Equal(t, time.Unix(30000, 0), o3.Price().Time)

	assert.Equal(t, provider.Pair{Base: "E", Quote: "F"}, o4.Price().Pair)
	assert.Equal(t, 30.0, o4.Price().Price)
	assert.Equal(t, 39.0, o4.Price().Bid)
	assert.Equal(t, 31.0, o4.Price().Ask)
	assert.Equal(t, 30.0, o4.Price().Volume24h)
	assert.Equal(t, time.Unix(30000, 0), o4.Price().Time)

	// Check if pairs were properly grouped per origins and check if the E/F pair
	// appeared only once:
	testPairs := s.Handlers()["test"].(*mockHandler).fetchPairs
	test2Pairs := s.Handlers()["test2"].(*mockHandler).fetchPairs
	assert.ElementsMatch(t, []origins.Pair{{Base: "A", Quote: "B"}, {Base: "C", Quote: "D"}}, testPairs)
	assert.ElementsMatch(t, []origins.Pair{{Base: "E", Quote: "F"}}, test2Pairs)
}

func TestFeeder_Feed_NestedOriginNode(t *testing.T) {
	s := originsSetMock(map[string][]origins.Price{
		"test": {
			origins.Price{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     10,
				Bid:       9,
				Ask:       11,
				Volume24h: 10,
				Timestamp: time.Unix(10000, 0),
			},
		},
	})

	g := nodes.NewMedianAggregatorNode(provider.Pair{Base: "A", Quote: "B"}, 1)
	i := nodes.NewIndirectAggregatorNode(provider.Pair{Base: "A", Quote: "B"})
	o := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test",
		Pair:   provider.Pair{Base: "A", Quote: "B"},
	}, 0, 0)

	g.AddChild(i)
	i.AddChild(o)

	f := NewFeeder(s, null.New())
	warns := f.Feed([]nodes.Node{g}, time.Now())

	assert.Len(t, warns.List, 0)
	assert.Equal(t, provider.Pair{Base: "A", Quote: "B"}, o.Price().Pair)
	assert.Equal(t, 10.0, o.Price().Price)
	assert.Equal(t, 9.0, o.Price().Bid)
	assert.Equal(t, 11.0, o.Price().Ask)
	assert.Equal(t, 10.0, o.Price().Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o.Price().Time)
}

func TestFeeder_Feed_BelowMinTTL(t *testing.T) {
	s := originsSetMock(map[string][]origins.Price{
		"test": {
			origins.Price{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     11,
				Bid:       10,
				Ask:       12,
				Volume24h: 11,
				Timestamp: time.Unix(10000, 0),
			},
		},
	})

	g := nodes.NewMedianAggregatorNode(provider.Pair{Base: "A", Quote: "B"}, 1)
	o := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test",
		Pair:   provider.Pair{Base: "A", Quote: "B"},
	}, 10*time.Second, 10*time.Second)

	_ = o.Ingest(nodes.OriginPrice{
		PairPrice: nodes.PairPrice{
			Pair:      provider.Pair{Base: "A", Quote: "B"},
			Price:     10,
			Bid:       9,
			Ask:       11,
			Volume24h: 10,
			Time:      time.Now().Add(-5 * time.Second),
		},
		Origin: "test",
		Error:  nil,
	})

	g.AddChild(o)

	f := NewFeeder(s, null.New())
	warns := f.Feed([]nodes.Node{g}, time.Now())

	// OriginNode shouldn't be updated because time diff is below MinTTL setting:
	assert.Len(t, warns.List, 0)
	assert.Equal(t, provider.Pair{Base: "A", Quote: "B"}, o.Price().Pair)
	assert.Equal(t, 10.0, o.Price().Price)
	assert.Equal(t, 9.0, o.Price().Bid)
	assert.Equal(t, 11.0, o.Price().Ask)
	assert.Equal(t, 10.0, o.Price().Volume24h)
}

func TestFeeder_Feed_BetweenTTLs(t *testing.T) {
	s := originsSetMock(map[string][]origins.Price{
		"test": {
			origins.Price{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     11,
				Bid:       10,
				Ask:       12,
				Volume24h: 11,
				Timestamp: time.Unix(10000, 0),
			},
		},
	})

	g := nodes.NewMedianAggregatorNode(provider.Pair{Base: "A", Quote: "B"}, 1)
	o := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test",
		Pair:   provider.Pair{Base: "A", Quote: "B"},
	}, 10*time.Second, 60*time.Second)

	_ = o.Ingest(nodes.OriginPrice{
		PairPrice: nodes.PairPrice{
			Pair:      provider.Pair{Base: "A", Quote: "B"},
			Price:     10,
			Bid:       9,
			Ask:       11,
			Volume24h: 10,
			Time:      time.Now().Add(-30 * time.Second),
		},
		Origin: "test",
		Error:  nil,
	})

	g.AddChild(o)

	f := NewFeeder(s, null.New())
	warns := f.Feed([]nodes.Node{g}, time.Now())

	// OriginNode should be updated because time diff is above MinTTL setting:
	assert.Len(t, warns.List, 0)
	assert.Equal(t, provider.Pair{Base: "A", Quote: "B"}, o.Price().Pair)
	assert.Equal(t, 11.0, o.Price().Price)
	assert.Equal(t, 10.0, o.Price().Bid)
	assert.Equal(t, 12.0, o.Price().Ask)
	assert.Equal(t, 11.0, o.Price().Volume24h)
}
