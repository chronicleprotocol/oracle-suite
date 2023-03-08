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

const krakenBaseURL = "https://api.kraken.com"
const krakenURL = "%s/0/public/Ticker?pair=%s"

type Kraken struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (k Kraken) Pool() query.WorkerPool {
	return k.WorkerPool
}
func (k Kraken) PullPrices(pairs []Pair) []FetchResult {
	req := &query.HTTPRequest{
		URL: buildOriginURL(krakenURL, k.BaseURL, krakenBaseURL, k.localPairName(pairs...)),
	}
	res := k.Pool().Query(req)
	if errorResponses := validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return k.parseResponse(pairs, res)
}

type krakenResponse struct {
	Errors []string `json:"error"`
	Result map[string]krakenPairResponse
}

type krakenPairResponse struct {
	Price  firstStringFromSliceAsFloat64 `json:"c"`
	Volume firstStringFromSliceAsFloat64 `json:"v"`
	Ask    firstStringFromSliceAsFloat64 `json:"a"`
	Bid    firstStringFromSliceAsFloat64 `json:"b"`
}

func (k *Kraken) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	var resp krakenResponse
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse response: %w", err))
	}
	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		if t, is := resp.Result[k.localPairName(pair)]; !is {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     t.Price.val(),
					Ask:       t.Ask.val(),
					Bid:       t.Bid.val(),
					Volume24h: t.Volume.val(),
					Timestamp: time.Now(),
				},
			})
		}
	}
	return results
}

func (k *Kraken) localPairName(pairs ...Pair) string {
	var l []string
	for _, pair := range pairs {
		l = append(l, pair.String())
	}
	return strings.Join(l, ",")
}
