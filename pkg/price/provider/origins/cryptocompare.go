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

const cryptoCompareMultiBaseURL = "https://min-api.cryptocompare.com"

// nolint:lll
const cryptoCompareMultiURL = "%s/data/pricemultifull?fsyms=%s&tsyms=%s&tryConversion=false&extraParams=gofer&relaxedValidation=true"

type cryptoCompareMultiResponse struct {
	Raw map[string]map[string]struct {
		Base  string  `json:"FROMSYMBOL"`
		Query string  `json:"TOSYMBOL"`
		Price float64 `json:"PRICE"`
		TS    int64   `json:"LASTUPDATE"`
		Vol24 float64 `json:"VOLUME24HOUR"`
	} `json:"RAW"`
}

type CryptoCompare struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (c CryptoCompare) Pool() query.WorkerPool {
	return c.WorkerPool
}

func (c CryptoCompare) PullPrices(pairs []Pair) []FetchResult {
	req := c.makeRequest(pairs)
	res := c.Pool().Query(req)
	if errorResponses := c.validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return c.parseResponse(pairs, res)
}

func (c *CryptoCompare) makeRequest(pairs []Pair) *query.HTTPRequest {
	var bList, qList []string

	for _, pair := range pairs {
		bList = append(bList, pair.Base)
	}
	for _, pair := range pairs {
		qList = append(qList, pair.Quote)
	}

	req := &query.HTTPRequest{
		URL: buildOriginURL(
			cryptoCompareMultiURL,
			c.BaseURL,
			cryptoCompareMultiBaseURL,
			strings.Join(bList, ","),
			strings.Join(qList, ","),
		),
	}
	return req
}

func (c *CryptoCompare) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	results := make([]FetchResult, 0)

	var resp cryptoCompareMultiResponse
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		for _, pair := range pairs {
			results = append(
				results,
				fetchResultWithError(pair, fmt.Errorf("failed to parse CryptoCompare response: %w", err)),
			)
		}
		return results
	}

	for _, pair := range pairs {
		if bObj, is := resp.Raw[pair.Base]; !is {
			results = append(
				results,
				fetchResultWithError(pair, fmt.Errorf("no response for %s base from CryptoCompare", pair.Base)),
			)
		} else {
			if qObj, is := bObj[pair.Quote]; !is {
				results = append(
					results,
					fetchResultWithError(pair, fmt.Errorf("no response for %s quote from CryptoCompare", pair.Quote)),
				)
			} else {
				results = append(results, FetchResult{
					Price: Price{
						Timestamp: time.Unix(qObj.TS, 0),
						Pair:      pair,
						Price:     qObj.Price,
						Volume24h: qObj.Vol24,
					},
					Error: nil,
				})
			}
		}
	}
	return results
}

func (c *CryptoCompare) validateResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	results := make([]FetchResult, 0)

	if res == nil {
		for _, pair := range pairs {
			results = append(results, fetchResultWithError(
				pair,
				fmt.Errorf("no response for %s from CryptoCompare", pair.String()),
			))
		}
		return results
	}
	if res.Error != nil {
		for _, pair := range pairs {
			results = append(results, fetchResultWithError(
				pair,
				res.Error,
			))
		}
		return results
	}

	return []FetchResult{}
}
