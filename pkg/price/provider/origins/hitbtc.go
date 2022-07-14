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

// Hitbtc URL
const hitbtcBaseURL = "https://api.hitbtc.com"
const hitbtcURL = "%s/api/2/public/ticker?symbols=%s"

type hitbtcResponse struct {
	Symbol    string    `json:"symbol"`
	Ask       string    `json:"ask"`
	Volume    string    `json:"volume"`
	Price     string    `json:"last"`
	Bid       string    `json:"bid"`
	Timestamp time.Time `json:"timestamp"`
}

// Hitbtc exchange handler
type Hitbtc struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (h *Hitbtc) localPairName(pair Pair) string {
	return strings.ToUpper(pair.Base + pair.Quote)
}

func (h *Hitbtc) getURL(pairs []Pair) string {
	pairsStr := make([]string, len(pairs))
	for i, pair := range pairs {
		pairsStr[i] = h.localPairName(pair)
	}
	return buildOriginURL(hitbtcURL, h.BaseURL, hitbtcBaseURL, strings.Join(pairsStr, ","))
}

func (h Hitbtc) Pool() query.WorkerPool {
	return h.WorkerPool
}

func (h Hitbtc) PullPrices(pairs []Pair) []FetchResult {
	crs, err := h.fetch(pairs)
	if err != nil {
		return fetchResultListWithErrors(pairs, err)
	}
	return crs
}

func (h *Hitbtc) fetch(pairs []Pair) ([]FetchResult, error) {
	req := &query.HTTPRequest{
		URL: h.getURL(pairs),
	}

	// make query
	res := h.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resps []hitbtcResponse
	err := json.Unmarshal(res.Body, &resps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hitbtc response: %w", err)
	}

	respMap := map[string]hitbtcResponse{}
	for _, resp := range resps {
		respMap[resp.Symbol] = resp
	}

	crs := make([]FetchResult, len(pairs))
	for i, pair := range pairs {
		symbol := h.localPairName(pair)
		if resp, has := respMap[symbol]; has {
			p, err := h.newPrice(pair, resp)
			if err != nil {
				crs[i] = fetchResultWithError(
					pair,
					fmt.Errorf("failed to create price point from hitbtc response: %w: %s", err, res.Body),
				)
			} else {
				crs[i] = fetchResult(p)
			}
		} else {
			crs[i] = fetchResultWithError(
				pair,
				fmt.Errorf("failed to find symbol %s in hitbtc response: %s", pair, res.Body),
			)
		}
	}
	return crs, nil
}

func (h *Hitbtc) newPrice(pair Pair, resp hitbtcResponse) (Price, error) {
	// Parsing price from string.
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return Price{}, fmt.Errorf("failed to parse price from hitbtc exchange")
	}
	// Parsing ask from string.
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return Price{}, fmt.Errorf("failed to parse ask from hitbtc exchange")
	}
	// Parsing volume from string.
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return Price{}, fmt.Errorf("failed to parse volume from hitbtc exchange")
	}
	// Parsing bid from string.
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return Price{}, fmt.Errorf("failed to parse bid from hitbtc exchange")
	}
	// Building Price.
	return Price{
		Pair:      pair,
		Price:     price,
		Ask:       ask,
		Bid:       bid,
		Volume24h: volume,
		Timestamp: resp.Timestamp,
	}, nil
}
