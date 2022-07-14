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

const folgoryBaseURL = "https://folgory.com"
const folgoryURL = "%s/api/v1"

type Folgory struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (o Folgory) Pool() query.WorkerPool {
	return o.WorkerPool
}

func (o Folgory) PullPrices(pairs []Pair) []FetchResult {
	req := &query.HTTPRequest{
		URL: buildOriginURL(folgoryURL, o.BaseURL, folgoryBaseURL),
	}
	res := o.Pool().Query(req)
	if errorResponses := validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return o.parseResponse(pairs, res)
}

type folgoryTicker struct {
	Symbol string          `json:"symbol"`
	Price  stringAsFloat64 `json:"last"`
	Volume stringAsFloat64 `json:"volume"`
}

func (o *Folgory) localPairName(pair Pair) string {
	return fmt.Sprintf("%s/%s", pair.Base, pair.Quote)
}

func (o *Folgory) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	results := make([]FetchResult, 0)
	var resp []folgoryTicker
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse response: %w", err))
	}

	tickers := make(map[string]folgoryTicker)
	for _, t := range resp {
		tickers[t.Symbol] = t
	}

	for _, pair := range pairs {
		if t, is := tickers[o.localPairName(pair)]; !is {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: fmt.Errorf("no response for %s", pair.String()),
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     t.Price.val(),
					Volume24h: t.Volume.val(),
					Timestamp: time.Now(),
				},
			})
		}
	}
	return results
}
