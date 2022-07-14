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

// Coinbase URL
const openExchangeRatesBaseURL = "https://openexchangerates.org"
const openExchangeRatesURL = "%s/api/latest.json?app_id=%s&base=%s&symbols=%s"

type openExchangeRatesResponse struct {
	Timestamp intAsUnixTimestamp `json:"timestamp"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
}

// OpenExchangeRates origin handler
type OpenExchangeRates struct {
	WorkerPool query.WorkerPool
	BaseURL    string
	APIKey     string
}

func (o *OpenExchangeRates) getURL(pair Pair) string {
	return buildOriginURL(openExchangeRatesURL, o.BaseURL, openExchangeRatesBaseURL, o.APIKey, pair.Base, pair.Quote)
}

func (o OpenExchangeRates) Pool() query.WorkerPool {
	return o.WorkerPool
}
func (o OpenExchangeRates) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&o, pairs)
}

func (o *OpenExchangeRates) callOne(pair Pair) (*Price, error) {
	var err error
	req := &query.HTTPRequest{
		URL: o.getURL(pair),
	}

	// make query
	res := o.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp openExchangeRatesResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenExchangeRate response: %w", err)
	}
	price, ok := resp.Rates[pair.Quote]
	if !ok {
		return nil, ErrMissingResponseForPair
	}
	// building Price
	return &Price{
		Pair:      pair,
		Price:     price,
		Timestamp: resp.Timestamp.val(),
	}, nil
}
