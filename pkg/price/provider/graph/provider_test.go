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

package graph

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/feeder"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/nodes"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/origins"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

var (
	testGraph  map[provider.Pair]nodes.Aggregator
	testFeeder *feeder.Feeder
	testPairs  = map[string]provider.Pair{
		"A/B": {Base: "A", Quote: "B"},
		"X/Y": {Base: "X", Quote: "Y"},
	}
	testTime   = time.Now()
	testModels = map[string]*provider.Model{
		"A/B": {
			Type:       "median",
			Parameters: map[string]string{},
			Pair:       testPairs["A/B"],
			Models: []*provider.Model{
				{
					Type:       "origin",
					Parameters: map[string]string{"origin": "a"},
					Pair:       testPairs["A/B"],
					Models:     nil,
				},
				{
					Type:       "median",
					Parameters: map[string]string{},
					Pair:       testPairs["A/B"],
					Models: []*provider.Model{
						{
							Type:       "origin",
							Parameters: map[string]string{"origin": "a"},
							Pair:       testPairs["A/B"],
							Models:     nil,
						},
						{
							Type:       "origin",
							Parameters: map[string]string{"origin": "b"},
							Pair:       testPairs["A/B"],
							Models:     nil,
						},
					},
				},
			},
		},
		"X/Y": {
			Type:       "median",
			Parameters: map[string]string{},
			Pair:       testPairs["X/Y"],
			Models: []*provider.Model{
				{
					Type:       "origin",
					Parameters: map[string]string{"origin": "x"},
					Pair:       testPairs["X/Y"],
					Models:     nil,
				},
				{
					Type:       "origin",
					Parameters: map[string]string{"origin": "y"},
					Pair:       testPairs["X/Y"],
					Models:     nil,
				},
			},
		},
	}
	testPrices = map[string]*provider.Price{
		"A/B": {
			Type: "aggregator",
			Parameters: map[string]string{
				"method":                   "median",
				"minimumSuccessfulSources": "0",
			},
			Pair:      provider.Pair{Base: "A", Quote: "B"},
			Price:     10,
			Bid:       9,
			Ask:       11,
			Volume24h: 0,
			Time:      testTime,
			Prices: []*provider.Price{
				{
					Type: "origin",
					Parameters: map[string]string{
						"origin": "a",
					},
					Pair:      provider.Pair{Base: "A", Quote: "B"},
					Price:     10,
					Bid:       9,
					Ask:       11,
					Volume24h: 20,
					Time:      testTime,
				},
				{
					Type: "aggregator",
					Parameters: map[string]string{
						"method":                   "median",
						"minimumSuccessfulSources": "0",
					},
					Pair:      provider.Pair{Base: "A", Quote: "B"},
					Price:     10,
					Bid:       9,
					Ask:       11,
					Volume24h: 0,
					Time:      testTime,
					Prices: []*provider.Price{
						{
							Type: "origin",
							Parameters: map[string]string{
								"origin": "a",
							},
							Pair:      provider.Pair{Base: "A", Quote: "B"},
							Price:     10,
							Bid:       9,
							Ask:       11,
							Volume24h: 20,
							Time:      testTime,
						},
						{
							Type: "origin",
							Parameters: map[string]string{
								"origin": "b",
							},
							Pair:      provider.Pair{Base: "A", Quote: "B"},
							Price:     10,
							Bid:       9,
							Ask:       11,
							Volume24h: 20,
							Time:      testTime,
						},
					},
				},
			},
		},
		"X/Y": {
			Type: "aggregator",
			Parameters: map[string]string{
				"method":                   "median",
				"minimumSuccessfulSources": "0",
			},
			Pair:      provider.Pair{Base: "X", Quote: "Y"},
			Price:     10,
			Bid:       9,
			Ask:       11,
			Volume24h: 0,
			Time:      testTime,
			Prices: []*provider.Price{
				{
					Type: "origin",
					Parameters: map[string]string{
						"origin": "x",
					},
					Pair:      provider.Pair{Base: "X", Quote: "Y"},
					Price:     10,
					Bid:       9,
					Ask:       11,
					Volume24h: 20,
					Time:      testTime,
				},
				{
					Type: "origin",
					Parameters: map[string]string{
						"origin": "y",
					},
					Pair:      provider.Pair{Base: "X", Quote: "Y"},
					Price:     10,
					Bid:       9,
					Ask:       11,
					Volume24h: 20,
					Time:      testTime,
				},
			},
		},
	}
)

type testExchange struct{}

func (f *testExchange) Fetch(pairs []origins.Pair) []origins.FetchResult {
	var r []origins.FetchResult
	for _, p := range pairs {
		r = append(r, origins.FetchResult{
			Price: origins.Price{
				Pair:      p,
				Price:     10,
				Bid:       9,
				Ask:       11,
				Volume24h: 20,
				Timestamp: testTime,
			},
			Error: nil,
		})
	}
	return r
}

func init() {
	ab := testPairs["A/B"]
	xy := testPairs["X/Y"]
	exp := 3600 * time.Second

	abGraph := nodes.NewMedianAggregatorNode(ab, 0)
	abc1 := nodes.NewOriginNode(nodes.OriginPair{Origin: "a", Pair: ab}, exp, exp)
	abc2 := nodes.NewOriginNode(nodes.OriginPair{Origin: "b", Pair: ab}, exp, exp)
	abc3 := nodes.NewMedianAggregatorNode(ab, 0)
	abGraph.AddChild(abc1)
	abGraph.AddChild(abc3)
	abc3.AddChild(abc1)
	abc3.AddChild(abc2)

	xyGraph := nodes.NewMedianAggregatorNode(xy, 0)
	xyc1 := nodes.NewOriginNode(nodes.OriginPair{Origin: "x", Pair: xy}, exp, exp)
	xyc2 := nodes.NewOriginNode(nodes.OriginPair{Origin: "y", Pair: xy}, exp, exp)
	xyGraph.AddChild(xyc1)
	xyGraph.AddChild(xyc2)

	testGraph = map[provider.Pair]nodes.Aggregator{
		ab: abGraph,
		xy: xyGraph,
	}

	testFeeder = feeder.NewFeeder(origins.NewSet(map[string]origins.Handler{
		"a": &testExchange{},
		"b": &testExchange{},
		"x": &testExchange{},
		"y": &testExchange{},
	}), null.New())
}

func TestGofer_Models_SinglePair(t *testing.T) {
	g := NewProvider(testGraph, testFeeder)
	r, err := g.Models(testPairs["A/B"])

	assert.Equal(t, map[provider.Pair]*provider.Model{
		testPairs["A/B"]: testModels["A/B"],
	}, r)
	assert.NoError(t, err)
}

func TestGofer_Models_AllPairs(t *testing.T) {
	g := NewProvider(testGraph, testFeeder)
	r, err := g.Models()

	assert.Equal(t, map[provider.Pair]*provider.Model{
		testPairs["A/B"]: testModels["A/B"],
		testPairs["X/Y"]: testModels["X/Y"],
	}, r)
	assert.NoError(t, err)
}

func TestGofer_Models_MissingPair(t *testing.T) {
	g := NewProvider(testGraph, testFeeder)
	_, err := g.Models(provider.Pair{})

	assert.True(t, errors.As(err, &ErrPairNotFound{}))
}

func TestGofer_Price(t *testing.T) {
	g := NewProvider(testGraph, testFeeder)
	r, err := g.Price(testPairs["A/B"])

	assert.Equal(t, testPrices["A/B"], r)
	assert.NoError(t, err)
}

func TestGofer_Price_MissingPair(t *testing.T) {
	g := NewProvider(testGraph, testFeeder)
	_, err := g.Price(provider.Pair{})

	assert.True(t, errors.As(err, &ErrPairNotFound{}))
}

func TestGofer_Prices_SinglePair(t *testing.T) {
	g := NewProvider(testGraph, testFeeder)
	r, err := g.Prices(testPairs["A/B"])

	assert.Equal(t, map[provider.Pair]*provider.Price{
		testPairs["A/B"]: testPrices["A/B"],
	}, r)
	assert.NoError(t, err)
}

func TestGofer_Prices_AllPair(t *testing.T) {
	g := NewProvider(testGraph, testFeeder)
	r, err := g.Prices()

	assert.Equal(t, map[provider.Pair]*provider.Price{
		testPairs["A/B"]: testPrices["A/B"],
		testPairs["X/Y"]: testPrices["X/Y"],
	}, r)
	assert.NoError(t, err)
}

func TestGofer_Prices_MissingPair(t *testing.T) {
	g := NewProvider(testGraph, testFeeder)
	_, err := g.Prices(provider.Pair{})

	assert.True(t, errors.As(err, &ErrPairNotFound{}))
}
