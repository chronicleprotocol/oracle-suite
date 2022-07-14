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
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

const ftxBaseURL = "https://ftx.com"
const ftxURL = "%s/api/markets"

type Ftx struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (f Ftx) Pool() query.WorkerPool {
	return f.WorkerPool
}

func (f Ftx) PullPrices(pairs []Pair) []FetchResult {
	req := &query.HTTPRequest{
		URL: buildOriginURL(ftxURL, f.BaseURL, ftxBaseURL),
	}
	res := f.Pool().Query(req)
	if errorResponses := validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return f.parseResponse(pairs, res)
}

type ftxResponse struct {
	Results []ftxTicker `json:"result"`
	Success bool        `json:"success"`
}

type ftxTicker struct {
	Ask            float64 `json:"ask"`
	BaseCurrency   string  `json:"baseCurrency"`
	Bid            float64 `json:"bid"`
	Change1H       float64 `json:"change1h"`
	Change24H      float64 `json:"change24h"`
	ChangeBod      float64 `json:"changeBod"`
	Enabled        bool    `json:"enabled"`
	Last           float64 `json:"last"`
	MinProvideSize float64 `json:"minProvideSize"`
	Name           string  `json:"name"`
	PostOnly       bool    `json:"postOnly"`
	Price          float64 `json:"price"`
	PriceIncrement float64 `json:"priceIncrement"`
	QuoteCurrency  string  `json:"quoteCurrency"`
	QuoteVolume24H float64 `json:"quoteVolume24h"`
	Restricted     bool    `json:"restricted"`
	SizeIncrement  float64 `json:"sizeIncrement"`
	Type           string  `json:"type"`
	Underlying     string  `json:"underlying"`
	VolumeUsd24H   float64 `json:"volumeUsd24h"`
}

func (f *Ftx) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	results := make([]FetchResult, 0)
	var resp ftxResponse
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse response: %w", err))
	}
	if !resp.Success {
		return fetchResultListWithErrors(pairs, ErrInvalidResponseStatus)
	}

	tickers := make(map[string]ftxTicker)
	for _, t := range resp.Results {
		tickers[t.Name] = t
	}

	for _, pair := range pairs {
		if t, is := tickers[f.localPairName(pair)]; !is {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     t.Last,
					Bid:       t.Bid,
					Ask:       t.Ask,
					Volume24h: t.QuoteVolume24H,
					Timestamp: time.Now(),
				},
			})
		}
	}
	return results
}

func (f *Ftx) localPairName(pair Pair) string {
	return fmt.Sprintf("%s/%s", pair.Base, pair.Quote)
}
