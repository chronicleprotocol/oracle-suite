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
	"encoding/json"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/internal/query"
	pkgEthereum "github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/origins"
)

func parseParamsSymbolAliases(params json.RawMessage) (origins.SymbolAliases, error) {
	if params == nil {
		return nil, fmt.Errorf("invalid origin parameters")
	}

	var res struct {
		SymbolAliases origins.SymbolAliases `json:"symbolAliases"`
	}
	err := json.Unmarshal(params, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin symbol aliases from params: %w", err)
	}
	return res.SymbolAliases, nil
}

func parseParamsAPIKey(params json.RawMessage) (string, error) {
	if params == nil {
		return "", fmt.Errorf("invalid origin parameters")
	}

	var res struct {
		APIKey string `json:"apiKey"`
	}
	err := json.Unmarshal(params, &res)
	if err != nil {
		return "", fmt.Errorf("failed to marshal origin symbol aliases from params: %w", err)
	}
	return res.APIKey, nil
}

func parseParamsContracts(params json.RawMessage) (origins.ContractAddresses, error) {
	if params == nil {
		return nil, fmt.Errorf("invalid origin parameters")
	}

	var res struct {
		Contracts origins.ContractAddresses `json:"contracts"`
	}
	err := json.Unmarshal(params, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin symbol aliases from params: %w", err)
	}
	return res.Contracts, nil
}

//nolint:funlen,gocyclo
func NewHandler(
	origin string,
	wp query.WorkerPool,
	cli pkgEthereum.Client,
	baseURL string,
	params json.RawMessage) (origins.Handler, error) {

	aliases, err := parseParamsSymbolAliases(params)
	if err != nil {
		return nil, err
	}
	switch origin {
	case "balancer":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(origins.Balancer{
			WorkerPool:        wp,
			BaseURL:           baseURL,
			ContractAddresses: contracts,
		}, aliases), nil
	case "binance":
		return origins.NewBaseExchangeHandler(origins.Binance{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "bitfinex":
		return origins.NewBaseExchangeHandler(origins.Bitfinex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "bitstamp":
		return origins.NewBaseExchangeHandler(origins.Bitstamp{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "bitthumb", "bithumb":
		return origins.NewBaseExchangeHandler(origins.BitThump{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "bittrex":
		return origins.NewBaseExchangeHandler(origins.Bittrex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "coinbase", "coinbasepro":
		return origins.NewBaseExchangeHandler(origins.CoinbasePro{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "cryptocompare":
		return origins.NewBaseExchangeHandler(origins.CryptoCompare{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "coinmarketcap":
		apiKey, err := parseParamsAPIKey(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(
			origins.CoinMarketCap{WorkerPool: wp, BaseURL: baseURL, APIKey: apiKey},
			aliases,
		), nil
	case "ddex":
		return origins.NewBaseExchangeHandler(origins.Ddex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "folgory":
		return origins.NewBaseExchangeHandler(origins.Folgory{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "ftx":
		return origins.NewBaseExchangeHandler(origins.Ftx{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "fx":
		apiKey, err := parseParamsAPIKey(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(
			origins.Fx{WorkerPool: wp, BaseURL: baseURL, APIKey: apiKey},
			aliases,
		), nil
	case "gateio":
		return origins.NewBaseExchangeHandler(origins.Gateio{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "gemini":
		return origins.NewBaseExchangeHandler(origins.Gemini{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "hitbtc":
		return origins.NewBaseExchangeHandler(origins.Hitbtc{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "huobi":
		return origins.NewBaseExchangeHandler(origins.Huobi{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "kraken":
		return origins.NewBaseExchangeHandler(origins.Kraken{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "kucoin":
		return origins.NewBaseExchangeHandler(origins.Kucoin{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "kyber":
		return origins.NewBaseExchangeHandler(origins.Kyber{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "loopring":
		return origins.NewBaseExchangeHandler(origins.Loopring{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "okex":
		return origins.NewBaseExchangeHandler(origins.Okex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "openexchangerates":
		apiKey, err := parseParamsAPIKey(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(
			origins.OpenExchangeRates{WorkerPool: wp, BaseURL: baseURL, APIKey: apiKey},
			aliases,
		), nil
	case "poloniex":
		return origins.NewBaseExchangeHandler(origins.Poloniex{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	case "sushiswap":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(origins.Sushiswap{
			WorkerPool:        wp,
			BaseURL:           baseURL,
			ContractAddresses: contracts,
		}, aliases), nil
	case "curve", "curvefinance":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		h, err := origins.NewCurveFinance(cli, contracts)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "balancerV2":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		h, err := origins.NewBalancerV2(cli, contracts)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "wsteth":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		h, err := origins.NewWrappedStakedETH(cli, contracts)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "uniswap", "uniswapV2":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(origins.Uniswap{
			WorkerPool:        wp,
			BaseURL:           baseURL,
			ContractAddresses: contracts,
		}, aliases), nil
	case "uniswapV3":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(origins.UniswapV3{
			WorkerPool:        wp,
			BaseURL:           baseURL,
			ContractAddresses: contracts,
		}, aliases), nil
	case "upbit":
		return origins.NewBaseExchangeHandler(origins.Upbit{WorkerPool: wp, BaseURL: baseURL}, aliases), nil
	}

	return nil, origins.ErrUnknownOrigin
}
