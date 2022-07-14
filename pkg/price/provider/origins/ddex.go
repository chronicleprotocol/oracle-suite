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

type Ddex struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

const ddexTickersBaseURL = "https://api.ddex.io"
const ddexTickersURL = "%s/v4/markets/tickers"

func (d Ddex) Pool() query.WorkerPool {
	return d.WorkerPool
}

func (d Ddex) PullPrices(pairs []Pair) []FetchResult {
	req := &query.HTTPRequest{
		URL: buildOriginURL(ddexTickersURL, d.BaseURL, ddexTickersBaseURL),
	}
	res := d.Pool().Query(req)
	if errorResponses := validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return d.parseResponse(pairs, res)
}

func (d *Ddex) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
}

type ddexTicker struct {
	Ask      stringAsFloat64      `json:"ask"`
	Bid      stringAsFloat64      `json:"bid"`
	High     stringAsFloat64      `json:"high"`
	Low      stringAsFloat64      `json:"low"`
	MarketID string               `json:"marketId"`
	Price    stringAsFloat64      `json:"price"`
	UpdateAt intAsUnixTimestampMs `json:"updateAt"`
	Volume   stringAsFloat64      `json:"volume"`
}
type ddexTickersResponse struct {
	Desc   string `json:"desc"`
	Status int    `json:"status"`
	Data   struct {
		Tickers []ddexTicker `json:"tickers"`
	} `json:"data"`
}

func (d *Ddex) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	results := make([]FetchResult, 0)
	var resp ddexTickersResponse
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse response: %w", err))
	}
	if resp.Status != 0 {
		return fetchResultListWithErrors(pairs, ErrInvalidResponseStatus)
	}

	tickers := make(map[string]ddexTicker)
	for _, t := range resp.Data.Tickers {
		tickers[t.MarketID] = t
	}

	for _, pair := range pairs {
		if t, is := tickers[d.localPairName(pair)]; !is {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     t.Price.val(),
					Bid:       t.Bid.val(),
					Ask:       t.Ask.val(),
					Volume24h: t.Volume.val(),
					Timestamp: t.UpdateAt.val(),
				},
			})
		}
	}
	return results
}
