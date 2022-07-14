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

const poloniexBaseURL = "https://poloniex.com"
const poloniexURL = "%s/public?command=returnTicker"

type poloniexResponse struct {
	Last       stringAsFloat64 `json:"Last"`
	HidPrice   stringAsFloat64 `json:"highestBid"`
	LowestAsk  stringAsFloat64 `json:"lowestAsk"`
	BaseVolume stringAsFloat64 `json:"baseVolume"`
	IsFrozen   string          `json:"isFrozen"`
}

// Poloniex origin handler
type Poloniex struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (p *Poloniex) localPairName(pair Pair) string {
	const (
		REP   = "REP"
		REPV2 = "REPV2"
	)

	if pair.Quote == REP {
		pair.Quote = REPV2
	}

	if pair.Base == REP {
		pair.Base = REPV2
	}

	return fmt.Sprintf("%s_%s", pair.Quote, pair.Base)
}

func (p Poloniex) Pool() query.WorkerPool {
	return p.WorkerPool
}

func (p Poloniex) PullPrices(pairs []Pair) []FetchResult {
	var err error
	req := &query.HTTPRequest{
		URL: buildOriginURL(poloniexURL, p.BaseURL, poloniexBaseURL),
	}

	// make query
	res := p.Pool().Query(req)
	if res == nil {
		return fetchResultListWithErrors(pairs, ErrEmptyOriginResponse)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, res.Error)
	}

	// parse JSON
	var resp map[string]poloniexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse Poloniex response: %w", err))
	}

	// prepare result
	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		if r, ok := resp[p.localPairName(pair)]; !ok {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else {
			if r.IsFrozen == "0" {
				results = append(results, FetchResult{
					Price: Price{
						Pair:      pair,
						Price:     r.Last.val(),
						Bid:       r.HidPrice.val(),
						Ask:       r.LowestAsk.val(),
						Volume24h: r.BaseVolume.val(),
						Timestamp: time.Now(),
					},
				})
			} else {
				results = append(results, FetchResult{
					Price: Price{Pair: pair},
					Error: fmt.Errorf("pair is indicated as a frozen"),
				})
			}
		}
	}

	return results
}
