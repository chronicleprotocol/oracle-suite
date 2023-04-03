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

package priceprovider

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/origins"
)

// averageFromBlocks is a list of blocks distances from the latest blocks from
// which prices will be averaged.
var averageFromBlocks = []int64{0, 10, 20}

func parseParamsSymbolAliases(params map[string]any) origins.SymbolAliases {
	aliasesMap, ok := params["symbol_aliases"].(map[string]any)
	if !ok {
		return nil
	}
	aliases := make(origins.SymbolAliases, len(aliasesMap))
	for k, v := range aliasesMap {
		aliases[k], _ = v.(string)
	}
	return aliases
}

func parseParamsContracts(params map[string]any) origins.ContractAddresses {
	addrsMap, ok := params["contracts"].(map[string]any)
	if !ok {
		return nil
	}
	addrs := make(origins.ContractAddresses, len(addrsMap))
	for k, v := range addrsMap {
		addrs[k], _ = v.(string)
	}
	return addrs
}

func parseSingleParam(params map[string]any, name string) string {
	if apiKey, ok := params[name].(string); ok {
		return apiKey
	}
	return ""
}

//nolint:funlen,gocyclo,whitespace
func NewHandler(
	origin string,
	wp query.WorkerPool,
	clients ethereum.ClientRegistry,
	params map[string]any,
) (origins.Handler, error) {

	aliases := parseParamsSymbolAliases(params)
	switch origin {
	case "balancer":
		contracts := parseParamsContracts(params)
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Balancer{
			WorkerPool:        wp,
			BaseURL:           baseURL,
			ContractAddresses: contracts,
		}, aliases), nil
	case "binance":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Binance{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "bitfinex":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Bitfinex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "bitstamp":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Bitstamp{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "bitthumb", "bithumb":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.BitThump{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "bittrex":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Bittrex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "coinbase", "coinbasepro":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.CoinbasePro{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "cryptocompare":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.CryptoCompare{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "coinmarketcap":
		baseURL := parseSingleParam(params, "url")
		apiKey := parseSingleParam(params, "api_key")
		return origins.NewBaseExchangeHandler(
			origins.CoinMarketCap{WorkerPool: wp, BaseURL: baseURL, APIKey: apiKey},
			aliases,
		), nil
	case "ddex":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Ddex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "folgory":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Folgory{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "fx":
		baseURL := parseSingleParam(params, "url")
		apiKey := parseSingleParam(params, "api_key")
		return origins.NewBaseExchangeHandler(
			origins.Fx{WorkerPool: wp, BaseURL: baseURL, APIKey: apiKey},
			aliases,
		), nil
	case "gateio":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Gateio{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "gemini":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Gemini{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "hitbtc":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Hitbtc{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "huobi":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Huobi{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "kraken":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Kraken{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "kucoin":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Kucoin{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "loopring":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Loopring{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "okex":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Okex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "okx":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Okx{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "openexchangerates":
		apiKey := parseSingleParam(params, "api_key")
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(
			origins.OpenExchangeRates{WorkerPool: wp, BaseURL: baseURL, APIKey: apiKey},
			aliases,
		), nil
	case "poloniex":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Poloniex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "sushiswap":
		baseURL := parseSingleParam(params, "url")
		contracts := parseParamsContracts(params)
		return origins.NewBaseExchangeHandler(origins.Sushiswap{
			WorkerPool:        wp,
			BaseURL:           baseURL,
			ContractAddresses: contracts,
		}, aliases), nil
	case "curve", "curvefinance":
		contracts := parseParamsContracts(params)
		clientName := parseSingleParam(params, "ethereum_client")
		client, ok := clients[clientName]
		if !ok {
			return nil, fmt.Errorf("ethereum client %s not found", clientName)
		}
		h, err := origins.NewCurveFinance(geth.NewClient(client), contracts, averageFromBlocks) //nolint:staticcheck
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "balancerV2":
		contracts := parseParamsContracts(params)
		clientName := parseSingleParam(params, "ethereum_client")
		client, ok := clients[clientName]
		if !ok {
			return nil, fmt.Errorf("ethereum client %s not found", clientName)
		}
		h, err := origins.NewBalancerV2(geth.NewClient(client), contracts, averageFromBlocks) //nolint:staticcheck
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "wsteth":
		contracts := parseParamsContracts(params)
		clientName := parseSingleParam(params, "ethereum_client")
		client, ok := clients[clientName]
		if !ok {
			return nil, fmt.Errorf("ethereum client %s not found", clientName)
		}
		h, err := origins.NewWrappedStakedETH(geth.NewClient(client), contracts, averageFromBlocks) //nolint:staticcheck
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "rocketpool":
		contracts := parseParamsContracts(params)
		clientName := parseSingleParam(params, "ethereum_client")
		client, ok := clients[clientName]
		if !ok {
			return nil, fmt.Errorf("ethereum client %s not found", clientName)
		}
		h, err := origins.NewRocketPool(geth.NewClient(client), contracts, averageFromBlocks) //nolint:staticcheck
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "uniswap", "uniswapV2":
		baseURL := parseSingleParam(params, "url")
		contracts := parseParamsContracts(params)
		return origins.NewBaseExchangeHandler(origins.Uniswap{
			WorkerPool:        wp,
			BaseURL:           baseURL,
			ContractAddresses: contracts,
		}, aliases), nil
	case "uniswapV3":
		baseURL := parseSingleParam(params, "url")
		contracts := parseParamsContracts(params)
		return origins.NewBaseExchangeHandler(origins.UniswapV3{
			WorkerPool:        wp,
			BaseURL:           baseURL,
			ContractAddresses: contracts,
		}, aliases), nil
	case "upbit":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.Upbit{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "ishares":
		baseURL := parseSingleParam(params, "url")
		return origins.NewBaseExchangeHandler(origins.IShares{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	}

	return nil, origins.ErrUnknownOrigin
}
