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

package gofer

import (
	"testing"
	"time"

	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/nodes"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/origins"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_buildGraphs_ValidConfig(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"B/C": {
				Method: "median",
				Sources: [][]Source{
					{
						{Origin: "bc1", Pair: "B/C"},
					},
					{
						{Origin: "bc2", Pair: "B/C"},
					},
				},
				Params: []byte(`{"minimumSuccessfulSources": 3}`),
			},
			"A/C": {
				Method: "median",
				Sources: [][]Source{
					{
						{Origin: "ab1", Pair: "A/B"},
						{Origin: "bc1", Pair: "B/C"},
					},
					{
						{Origin: "ab2", Pair: "A/B"},
						{Origin: ".", Pair: "B/C"},
					},
				},
				Params: []byte(`{"minimumSuccessfulSources": 3}`),
			},
		},
	}

	c, err2 := config.buildGraphs()
	assert.Nil(t, err2)

	// List of pairs used in config file:
	ab := gofer.Pair{Base: "A", Quote: "B"}
	bc := gofer.Pair{Base: "B", Quote: "C"}
	ac := gofer.Pair{Base: "A", Quote: "C"}

	// Check if all three pairs was loaded correctly:
	assert.Contains(t, c, bc)
	assert.Contains(t, c, ac)
	assert.IsType(t, &nodes.MedianAggregatorNode{}, c[bc])
	assert.IsType(t, &nodes.MedianAggregatorNode{}, c[ac])

	// --- Tests for B/C pair ---
	assert.Len(t, c[bc].Children(), 2)
	// Sources have only one pair so we expect OriginNodes instead of
	// the IndirectAggregatorNode:
	assert.IsType(t, &nodes.OriginNode{}, c[bc].Children()[0])
	assert.IsType(t, &nodes.OriginNode{}, c[bc].Children()[1])
	// Check if pairs was assigned correctly to nodes:
	assert.Equal(t, "bc1", c[bc].Children()[0].(*nodes.OriginNode).OriginPair().Origin)
	assert.Equal(t, "bc2", c[bc].Children()[1].(*nodes.OriginNode).OriginPair().Origin)
	assert.Equal(t, bc, c[bc].Children()[0].(*nodes.OriginNode).OriginPair().Pair)
	assert.Equal(t, bc, c[bc].Children()[1].(*nodes.OriginNode).OriginPair().Pair)

	// --- Tests for A/C pair ---
	assert.Len(t, c[ac].Children(), 2)
	// Sources have more than one pair so now we expect the
	// IndirectAggregatorNode.
	assert.IsType(t, &nodes.IndirectAggregatorNode{}, c[ac].Children()[0])
	assert.IsType(t, &nodes.IndirectAggregatorNode{}, c[ac].Children()[1])
	// Check if pairs was assigned correctly to nodes:
	assert.Equal(t, ac, c[ac].Children()[0].(*nodes.IndirectAggregatorNode).Pair())
	assert.Equal(t, ac, c[ac].Children()[1].(*nodes.IndirectAggregatorNode).Pair())
	assert.Equal(t, ab, c[ac].Children()[0].(*nodes.IndirectAggregatorNode).Children()[0].(*nodes.OriginNode).OriginPair().Pair)
	assert.Equal(t, bc, c[ac].Children()[0].(*nodes.IndirectAggregatorNode).Children()[1].(*nodes.OriginNode).OriginPair().Pair)
	// In a second source, there is a reference to another root node. We should
	// use previously created instance instead creating new one:
	assert.Same(t, c[bc], c[ac].Children()[1].(*nodes.IndirectAggregatorNode).Children()[1])
}

func TestConfig_buildGraphs_CyclicConfig(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"A/B": {
				Method: "median",
				Sources: [][]Source{
					{
						{Origin: "ab1", Pair: "A/B"},
					},
					{
						{Origin: "ab2", Pair: "A/B"},
					},
					{
						{Origin: "ac1", Pair: "A/B"},
						{Origin: ".", Pair: "B/C"},
					},
				},
				Params: []byte(`{"minimumSuccessfulSources": 3}`),
			},
			"A/C": {
				Method: "median",
				Sources: [][]Source{
					{
						{Origin: "ab1", Pair: "A/B"},
						{Origin: ".", Pair: "B/C"},
					},
				},
				Params: []byte(`{"minimumSuccessfulSources": 3}`),
			},
			"B/C": {
				Method: "median",
				Sources: [][]Source{
					{
						{Origin: "ab1", Pair: "A/B"},
						{Origin: ".", Pair: "A/C"},
					},
				},
				Params: []byte(`{"minimumSuccessfulSources": 3}`),
			},
		},
	}

	_, err2 := config.buildGraphs()
	assert.Error(t, err2)
}

func TestConfig_buildGraphs_NoSources(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"A/B": {
				Method:  "median",
				Sources: [][]Source{},
			},
		},
	}

	_, err2 := config.buildGraphs()
	assert.Nil(t, err2)
}

func TestConfig_buildGraphs_InvalidPairName(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"A_B": { // the "A_B" name is incorrect
				Method:  "median",
				Sources: [][]Source{},
			},
		},
	}

	_, err2 := config.buildGraphs()
	assert.Error(t, err2)
}

func TestConfig_buildGraphs_ReferenceToMissingPair(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"A/B": {
				Method: "median",
				Sources: [][]Source{
					{
						{Origin: ".", Pair: "X/Y"},
					},
				},
			},
		},
	}

	_, err2 := config.buildGraphs()
	assert.Error(t, err2)
}

func TestConfig_buildGraphs_ReferenceToSelf(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"A/B": {
				Method: "median",
				Sources: [][]Source{
					{
						{Origin: ".", Pair: "A/B"},
					},
				},
			},
		},
	}

	_, err2 := config.buildGraphs()
	assert.Error(t, err2)
}

func TestConfig_buildGraphs_DefaultTTL(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"A/B": {
				Method: "median",
				Sources: [][]Source{
					{
						{Origin: "ab", Pair: "A/B"},
					},
				},
			},
		},
	}

	p, _ := gofer.NewPair("A/B")
	g, _ := config.buildGraphs()

	assert.Equal(t, 300*time.Second, g[p].Children()[0].(*nodes.OriginNode).MaxTTL())
	assert.Equal(t, 60*time.Second, g[p].Children()[0].(*nodes.OriginNode).MinTTL())
}

func TestConfig_buildGraphs_OriginTTL(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"A/B": {
				Method: "median",
				TTL:    90, // should be ignored
				Sources: [][]Source{
					{
						{Origin: "ab", Pair: "A/B", TTL: 120},
					},
				},
			},
		},
	}

	p, _ := gofer.NewPair("A/B")
	g, _ := config.buildGraphs()

	assert.Equal(t, 360*time.Second, g[p].Children()[0].(*nodes.OriginNode).MaxTTL())
	assert.Equal(t, 120*time.Second, g[p].Children()[0].(*nodes.OriginNode).MinTTL())
}

func TestConfig_buildGraphs_MedianTTL(t *testing.T) {
	config := Gofer{
		Origins: nil,
		PriceModels: map[string]PriceModel{
			"A/B": {
				Method: "median",
				TTL:    120,
				Sources: [][]Source{
					{
						{Origin: "ab", Pair: "A/B"},
					},
				},
			},
		},
	}

	p, _ := gofer.NewPair("A/B")
	g, _ := config.buildGraphs()

	assert.Equal(t, 360*time.Second, g[p].Children()[0].(*nodes.OriginNode).MaxTTL())
	assert.Equal(t, 120*time.Second, g[p].Children()[0].(*nodes.OriginNode).MinTTL())
}

func TestConfig_buildGraphs_UpdatedOriginURL(t *testing.T) {
	url := "http://localhost:8080"

	config := Gofer{
		Origins: map[string]Origin{
			"ab": {
				Type:   "binance",
				Name:   "ab",
				URL:    url,
				Params: []byte(`{}`),
			},
		},
		PriceModels: map[string]PriceModel{
			"A/B": {
				Method: "median",
				TTL:    120,
				Sources: [][]Source{
					{
						{Origin: "ab", Pair: "A/B"},
					},
				},
			},
		},
	}

	o, err := config.buildOrigins(&ethereumMocks.Client{})
	require.NoError(t, err)
	require.NotNil(t, o)

	h := o.Handlers()
	require.NotNil(t, h)
	require.NotNil(t, h["ab"])

	handler := h["ab"].(*origins.BaseExchangeHandler)
	require.NotNil(t, handler.ExchangeHandler)

	bin := handler.ExchangeHandler.(origins.Binance)
	require.NotNil(t, bin)
	require.Equal(t, url, bin.BaseURL)
}
