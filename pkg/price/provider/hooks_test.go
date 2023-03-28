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

package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"

	"github.com/stretchr/testify/assert"
)

func TestPostPriceHook(t *testing.T) {
	prices := make(map[Pair]*Price)

	pair, err := NewPair("RETH/ETH")
	assert.NoError(t, err)

	prices[pair] = &Price{
		Price: 1.0,
		Prices: []*Price{
			{
				Parameters: map[string]string{
					"origin": "rocketpool",
				},
				Price: 0.99,
			},
		},
	}

	contract := "0x0011223344556677889900112233445566778899"
	params := make(map[string]interface{})
	params["circuit_contract"] = contract
	params["ethereum_client"] = "default"
	pairParams := NewHookParams()
	pairParams["RETH/ETH"] = params

	cli := &ethereumMocks.RPC{}
	readMethodID := []byte{87, 222, 38, 164}
	divisorMethodID := []byte{31, 45, 197, 239}

	ctx := context.Background()
	cli.On("Call", ctx, types.Call{To: types.MustAddressFromHexPtr(contract), Input: readMethodID}, types.LatestBlockNumber).Return(
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 39, 16},
		nil,
	)
	cli.On("Call", ctx, types.Call{To: types.MustAddressFromHexPtr(contract), Input: divisorMethodID}, types.LatestBlockNumber).Return(
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 134, 160},
		nil,
	)

	clients := ethereum.ClientRegistry{
		"default": cli,
	}
	hook, err := NewPostPriceHook(context.Background(), clients, pairParams)
	assert.NoError(t, err)

	// Moderate price deviation, expect success
	err = hook.Check(prices)
	assert.NoError(t, err)
	assert.True(t, prices[pair].Error == "")

	// Extreme price deviation, expect error
	prices[pair].Prices[0].Price = 2.0
	err = hook.Check(prices)
	assert.NoError(t, err)
	assert.True(t, prices[pair].Error != "")
	t.Log(prices[pair].Error)

	// Price pair with no hooks defined succeeds
	prices = make(map[Pair]*Price)
	pair, err = NewPair("WAKA/WAKA")
	assert.NoError(t, err)

	prices[pair] = &Price{Price: 1.0}
	err = hook.Check(prices)
	assert.NoError(t, err)
	assert.True(t, prices[pair].Error == "")
}

func TestFindPrices(t *testing.T) {
	var prices Price

	j := `{
  "type": "aggregator",
  "base": "RETH",
  "quote": "USD",
  "price": 1240.7604730837622,
  "bid": 0,
  "ask": 0,
  "vol24h": 0,
  "ts": "2022-11-23T18:16:19Z",
  "Parameters": {
    "method": "median",
    "minimumSuccessfulSources": "1"
  },
  "prices": [
    {
      "type": "aggregator",
      "base": "RETH",
      "quote": "USD",
      "price": 1240.7604730837622,
      "bid": 0,
      "ask": 0,
      "vol24h": 0,
      "ts": "2022-11-23T18:16:19Z",
      "Parameters": {
        "method": "indirect"
      },
      "prices": [
        {
          "type": "aggregator",
          "base": "RETH",
          "quote": "ETH",
          "price": 1.0749148255322616,
          "bid": 0,
          "ask": 0,
          "vol24h": 0,
          "ts": "2022-11-23T18:16:21.188533Z",
          "Parameters": {
            "method": "median",
            "minimumSuccessfulSources": "3"
          },
          "prices": [
            {
              "type": "origin",
              "base": "RETH",
              "quote": "ETH",
              "price": 1.048148949276213,
              "bid": 0,
              "ask": 0,
              "vol24h": 0,
              "ts": "2022-11-23T18:16:21.194933Z",
              "Parameters": {
                "origin": "rocketpool"
              }
            },
            {
              "type": "origin",
              "base": "RETH",
              "quote": "ETH",
              "price": 1.0755430566179713,
              "bid": 0,
              "ask": 0,
              "vol24h": 0,
              "ts": "2022-11-23T18:16:21.291785Z",
              "Parameters": {
                "origin": "balancerV2"
              }
            },
            {
              "type": "aggregator",
              "base": "RETH",
              "quote": "ETH",
              "price": 1.0749148255322616,
              "bid": 0,
              "ask": 0,
              "vol24h": 0,
              "ts": "2022-11-23T18:16:21.188533Z",
              "Parameters": {
                "method": "indirect"
              },
              "prices": [
                {
                  "type": "origin",
                  "base": "RETH",
                  "quote": "WSTETH",
                  "price": 0.9956479520503729,
                  "bid": 0,
                  "ask": 0,
                  "vol24h": 0,
                  "ts": "2022-11-23T18:16:21.284767Z",
                  "Parameters": {
                    "origin": "curve"
                  }
                },
                {
                  "type": "aggregator",
                  "base": "WSTETH",
                  "quote": "ETH",
                  "price": 1.0796133546186195,
                  "bid": 0,
                  "ask": 0,
                  "vol24h": 0,
                  "ts": "2022-11-23T18:16:21.188533Z",
                  "Parameters": {
                    "method": "median",
                    "minimumSuccessfulSources": "1"
                  },
                  "prices": [
                    {
                      "type": "aggregator",
                      "base": "WSTETH",
                      "quote": "ETH",
                      "price": 1.0796133546186195,
                      "bid": 0,
                      "ask": 0,
                      "vol24h": 0,
                      "ts": "2022-11-23T18:16:21.188533Z",
                      "Parameters": {
                        "method": "indirect"
                      },
                      "prices": [
                        {
                          "type": "origin",
                          "base": "WSTETH",
                          "quote": "STETH",
                          "price": 1.0975476180869606,
                          "bid": 0,
                          "ask": 0,
                          "vol24h": 0,
                          "ts": "2022-11-23T18:16:21.188533Z",
                          "Parameters": {
                            "origin": "wsteth"
                          }
                        },
                        {
                          "type": "aggregator",
                          "base": "STETH",
                          "quote": "ETH",
                          "price": 0.9836596944198187,
                          "bid": 0,
                          "ask": 0,
                          "vol24h": 0,
                          "ts": "2022-11-23T18:16:21.284778Z",
                          "Parameters": {
                            "method": "median",
                            "minimumSuccessfulSources": "2"
                          },
                          "prices": [
                            {
                              "type": "origin",
                              "base": "STETH",
                              "quote": "ETH",
                              "price": 0.9833888745671291,
                              "bid": 0,
                              "ask": 0,
                              "vol24h": 0,
                              "ts": "2022-11-23T18:16:21.284778Z",
                              "Parameters": {
                                "origin": "curve"
                              }
                            },
                            {
                              "type": "origin",
                              "base": "STETH",
                              "quote": "ETH",
                              "price": 0.9839305142725083,
                              "bid": 0,
                              "ask": 0,
                              "vol24h": 0,
                              "ts": "2022-11-23T18:16:21.291791Z",
                              "Parameters": {
                                "origin": "balancerV2"
                              }
                            }
                          ]
                        }
                      ]
                    }
                  ]
                }
              ]
            }
          ]
        },
        {
          "type": "aggregator",
          "base": "ETH",
          "quote": "USD",
          "price": 1154.2872454748956,
          "bid": 1154.2822454748957,
          "ask": 1154.58,
          "vol24h": 0,
          "ts": "2022-11-23T18:16:19Z",
          "Parameters": {
            "method": "median",
            "minimumSuccessfulSources": "4"
          },
          "prices": [
            {
              "type": "origin",
              "base": "ETH",
              "quote": "USD",
              "price": 1153.7,
              "bid": 1153.3,
              "ask": 1153.9,
              "vol24h": 12684.12230558,
              "ts": "2022-11-23T18:16:19Z",
              "Parameters": {
                "origin": "bitstamp"
              }
            },
            {
              "type": "origin",
              "base": "ETH",
              "quote": "USD",
              "price": 1154.82,
              "bid": 1154.62,
              "ask": 1154.81,
              "vol24h": 442138.66656182,
              "ts": "2022-11-23T18:16:20.850441Z",
              "Parameters": {
                "origin": "coinbasepro"
              }
            },
            {
              "type": "origin",
              "base": "ETH",
              "quote": "USD",
              "price": 1153.47,
              "bid": 1153.89,
              "ask": 1155.04,
              "vol24h": 0,
              "ts": "2022-11-23T18:16:21.091778Z",
              "Parameters": {
                "origin": "gemini"
              }
            },
            {
              "type": "origin",
              "base": "ETH",
              "quote": "USD",
              "price": 1154.35,
              "bid": 1154.34,
              "ask": 1154.35,
              "vol24h": 40307.1735873,
              "ts": "2022-11-23T18:16:20.926406Z",
              "Parameters": {
                "origin": "kraken"
              }
            },
            {
              "type": "origin",
              "base": "ETH",
              "quote": "USD",
              "price": 1154.2244909497913,
              "bid": 1154.2244909497913,
              "ask": 1154.2244909497913,
              "vol24h": 138351857.4053526,
              "ts": "2022-11-23T18:16:20.915368Z",
              "Parameters": {
                "origin": "uniswapV3"
              }
            },
            {
              "type": "aggregator",
              "base": "ETH",
              "quote": "USD",
              "price": 1154.897624,
              "bid": 1154.9187476,
              "ask": 1155.2048316900002,
              "vol24h": 0,
              "ts": "2022-11-23T18:16:20Z",
              "Parameters": {
                "method": "indirect"
              },
              "prices": [
                {
                  "type": "origin",
                  "base": "ETH",
                  "quote": "BTC",
                  "price": 0.070412,
                  "bid": 0.070412,
                  "ask": 0.070413,
                  "vol24h": 90676.2152,
                  "ts": "2022-11-23T18:16:21Z",
                  "Parameters": {
                    "origin": "binance"
                  }
                },
                {
                  "type": "aggregator",
                  "base": "BTC",
                  "quote": "USD",
                  "price": 16402,
                  "bid": 16402.3,
                  "ask": 16406.13,
                  "vol24h": 0,
                  "ts": "2022-11-23T18:16:20Z",
                  "Parameters": {
                    "method": "median",
                    "minimumSuccessfulSources": "3"
                  },
                  "prices": [
                    {
                      "type": "origin",
                      "base": "BTC",
                      "quote": "USD",
                      "price": 16405.01,
                      "bid": 16403.02,
                      "ask": 16406.13,
                      "vol24h": 7683.190338,
                      "ts": "2022-11-23T18:16:20Z",
                      "Parameters": {
                        "origin": "binance_us"
                      }
                    },
                    {
                      "type": "origin",
                      "base": "BTC",
                      "quote": "USD",
                      "price": 16402,
                      "bid": 16397,
                      "ask": 16403,
                      "vol24h": 3166.14749227,
                      "ts": "2022-11-23T18:16:21Z",
                      "Parameters": {
                        "origin": "bitstamp"
                      }
                    },
                    {
                      "type": "origin",
                      "base": "BTC",
                      "quote": "USD",
                      "price": 16403.56,
                      "bid": 16403.98,
                      "ask": 16406.28,
                      "vol24h": 35088.83987094,
                      "ts": "2022-11-23T18:16:20.900259Z",
                      "Parameters": {
                        "origin": "coinbasepro"
                      }
                    },
                    {
                      "type": "origin",
                      "base": "BTC",
                      "quote": "USD",
                      "price": 16393.72,
                      "bid": 16396.57,
                      "ask": 16411.78,
                      "vol24h": 0,
                      "ts": "2022-11-23T18:16:21.003681Z",
                      "Parameters": {
                        "origin": "gemini"
                      }
                    },
                    {
                      "type": "origin",
                      "base": "BTC",
                      "quote": "USD",
                      "price": 16402,
                      "bid": 16402.3,
                      "ask": 16404.1,
                      "vol24h": 2535.22147078,
                      "ts": "2022-11-23T18:16:20.926404Z",
                      "Parameters": {
                        "origin": "kraken"
                      }
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}`

	err := json.Unmarshal([]byte(j), &prices)
	assert.NoError(t, err)
	refPrice := findPrice(prices.Prices,
		func(p *Price) bool {
			return p.Parameters["origin"] == "rocketpool"
		},
	)
	assert.True(t, refPrice.Price == 1.048148949276213)
}
