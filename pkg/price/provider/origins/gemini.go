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
	"strconv"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

// Gemini URL
const geminiBaseURL = "https://api.gemini.com"
const geminiURL = "%s/v1/pubticker/%s"

type geminiResponse struct {
	Price string `json:"last"`
	Ask   string `json:"ask"`
	Bid   string `json:"bid"`
}

// Gemini origin handler
type Gemini struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (g *Gemini) localPairName(pair Pair) string {
	return strings.ToLower(pair.Base + pair.Quote)
}

func (g *Gemini) getURL(pair Pair) string {
	return buildOriginURL(geminiURL, g.BaseURL, geminiBaseURL, g.localPairName(pair))
}

func (g Gemini) Pool() query.WorkerPool {
	return g.WorkerPool
}

func (g Gemini) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&g, pairs)
}

func (g *Gemini) callOne(pair Pair) (*Price, error) {
	var err error
	req := &query.HTTPRequest{
		URL: g.getURL(pair),
	}

	// make query
	res := g.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp geminiResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gemini response: %w", err)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from gemini origin %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from gemini origin %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from gemini origin %s", res.Body)
	}
	// building Price
	return &Price{
		Pair:      pair,
		Price:     price,
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Now(),
	}, nil
}
