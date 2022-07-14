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
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

// Exchange URL
const coinMarketCapBaseURL = "https://pro-api.coinmarketcap.com"
const coinMarketCapURL = "%s/v1/cryptocurrency/quotes/latest?id=%s"

type quoteResponse struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume_24h"`
}

type coinMarketCapPairResponse struct {
	Quote map[string]quoteResponse
}

type coinMarketCapResponse struct {
	Status struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}
	Data map[string]coinMarketCapPairResponse
}

// Exchange handler
type CoinMarketCap struct {
	WorkerPool query.WorkerPool
	BaseURL    string
	APIKey     string
}

// GetURL implementation
func (c *CoinMarketCap) getURL(uriPairs []string) string {
	return buildOriginURL(coinMarketCapURL, c.BaseURL, coinMarketCapBaseURL, strings.Join(uriPairs, ","))
}

// LocalPairName implementation
func (c *CoinMarketCap) localPairName(pair Pair) string {
	switch pair {
	case Pair{Base: "BTC", Quote: "USD"}:
		return "1"
	case Pair{Base: "USDT", Quote: "USD"}:
		return "825"
	case Pair{Base: "POLY", Quote: "USD"}:
		return "2496"
	default:
		return pair.Base
	}
}

func (c CoinMarketCap) Pool() query.WorkerPool {
	return c.WorkerPool
}

func (c CoinMarketCap) PullPrices(pairs []Pair) []FetchResult {
	var uriPairs []string
	for _, pair := range pairs {
		uriPairs = append(uriPairs, c.localPairName(pair))
	}

	var err error
	req := &query.HTTPRequest{
		URL: c.getURL(uriPairs),
		Headers: map[string]string{
			"CMC_PRO_API_KEY": c.APIKey,
			"Accept":          "application/json",
		},
	}
	// make query
	res := c.Pool().Query(req)
	if res == nil {
		return fetchResultListWithErrors(pairs, ErrEmptyOriginResponse)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, res.Error)
	}

	// parsing JSON
	var resp coinMarketCapResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse loopring response: %w", err))
	}
	if resp.Status.ErrorCode != 0 || resp.Status.ErrorMessage != "" {
		//nolint:lll
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to get data from coinmarketcap: %s %s", resp.Status.ErrorMessage, res.Body))
	}
	if resp.Data == nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("empty `data` field for loopring response: %s", res.Body))
	}

	var results []FetchResult
	for _, pair := range pairs {
		results = append(results, c.pickPairDetails(resp, pair))
	}
	return results
}

func (c *CoinMarketCap) pickPairDetails(response coinMarketCapResponse, pair Pair) FetchResult {
	pairName := c.localPairName(pair)
	pairResp, ok := response.Data[pairName]
	if !ok {
		return fetchResultWithError(pair, fmt.Errorf("no %s pair exist in loopring response", pair))
	}
	quoteRes, ok := pairResp.Quote[pair.Quote]
	if !ok {
		return fetchResultWithError(pair, fmt.Errorf("failed to get quote response for %s", pairName))
	}
	// building Price
	return fetchResult(Price{
		Pair:      pair,
		Price:     quoteRes.Price,
		Volume24h: quoteRes.Volume,
		Timestamp: time.Now(),
	})
}
