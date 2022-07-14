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

package origins

import (
	"encoding/json"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

const kyberBaseURL = "https://api.kyber.network"
const kyberURL = "%s/change24h"

type Kyber struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (k Kyber) Pool() query.WorkerPool {
	return k.WorkerPool
}

func (k Kyber) PullPrices(pairs []Pair) []FetchResult {
	req := &query.HTTPRequest{
		URL: buildOriginURL(kyberURL, k.BaseURL, kyberBaseURL),
	}
	res := k.Pool().Query(req)
	if errorResponses := validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return k.parseResponse(pairs, res)
}

type kyberTicker struct {
	Timestamp    intAsUnixTimestampMs `json:"timestamp"`
	TokenName    string               `json:"token_name"`
	TokenSymbol  string               `json:"token_symbol"`
	TokenDecimal int                  `json:"token_decimal"`
	TokenAddress string               `json:"token_address"`
	RateEthNow   float64              `json:"rate_eth_now"`
	ChangeEth24H float64              `json:"change_eth_24h"`
	ChangeUsd24H float64              `json:"change_usd_24h"`
	RateUsdNow   float64              `json:"rate_usd_now"`
}

func (k *Kyber) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	results := make([]FetchResult, 0)
	var tickers map[string]kyberTicker
	err := json.Unmarshal(res.Body, &tickers)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse response: %w", err))
	}

	for _, pair := range pairs {
		//nolint:gocritic
		if t, is := tickers[pair.Quote+"_"+pair.Base]; !is {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else if t.TokenSymbol != pair.Base {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrInvalidPrice,
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     t.RateEthNow,
					Timestamp: t.Timestamp.val(),
				},
			})
		}
	}
	return results
}
