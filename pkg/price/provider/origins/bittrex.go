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

const bittrexBaseURL = "https://api.bittrex.com"
const bittrexURL = "%s/api/v1.1/public/getticker?market=%s"

type bittrexResponse struct {
	Success bool                  `json:"success"`
	Result  bittrexSymbolResponse `json:"result"`
}

type bittrexSymbolResponse struct {
	Ask  float64 `json:"Ask"`
	Bid  float64 `json:"Bid"`
	Last float64 `json:"Last"`
}

// Bittrex origin handler
type Bittrex struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (b *Bittrex) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", pair.Quote, pair.Base)
}

func (b Bittrex) Pool() query.WorkerPool {
	return b.WorkerPool
}

func (b Bittrex) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&b, pairs)
}

func (b *Bittrex) callOne(pair Pair) (*Price, error) {
	var err error
	req := &query.HTTPRequest{
		URL: buildOriginURL(bittrexURL, b.BaseURL, bittrexBaseURL, b.localPairName(pair)),
	}

	// make query
	res := b.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parse JSON
	var resp bittrexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Bittrex response: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("wrong response from Bittrex %v", resp)
	}

	return &Price{
		Pair:      pair,
		Price:     resp.Result.Last,
		Bid:       resp.Result.Bid,
		Ask:       resp.Result.Ask,
		Timestamp: time.Now(),
	}, nil
}
