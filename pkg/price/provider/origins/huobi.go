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

// Huobi URL
const huobiBaseURL = "https://api.huobi.pro"
const huobiURL = "%s/market/tickers"

type huobiResponse struct {
	Symbol string  `json:"symbol"`
	Volume float64 `json:"vol"`
	Bid    float64 `json:"bid"`
	Ask    float64 `json:"ask"`
}

// Huobi origin handler
type Huobi struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (h *Huobi) localPairName(pair Pair) string {
	return strings.ToLower(pair.Base + pair.Quote)
}

func (h Huobi) Pool() query.WorkerPool {
	return h.WorkerPool
}

func (h Huobi) PullPrices(pairs []Pair) []FetchResult {
	frs, err := h.fetch(pairs)
	if err != nil {
		return fetchResultListWithErrors(pairs, err)
	}
	return frs
}

func (h *Huobi) fetch(pairs []Pair) ([]FetchResult, error) {
	var err error
	req := &query.HTTPRequest{
		URL: buildOriginURL(huobiURL, h.BaseURL, huobiBaseURL),
	}

	res := h.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	var resp struct {
		Status    string          `json:"status"`
		Timestamp int64           `json:"ts"`
		Data      []huobiResponse `json:"data"`
	}

	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse huobi response: %w", err)
	}
	if resp.Status == "error" {
		return nil, fmt.Errorf("error response from huobi origin %s", res.Body)
	}

	respMap := map[string]huobiResponse{}
	for _, t := range resp.Data {
		respMap[t.Symbol] = t
	}

	ts := time.Unix(resp.Timestamp/1000, 0)
	frs := make([]FetchResult, len(pairs))
	for i, p := range pairs {
		if t, has := respMap[h.localPairName(p)]; has {
			frs[i] = fetchResult(Price{
				Pair:      p,
				Price:     (t.Ask + t.Bid) / 2,
				Ask:       t.Ask,
				Bid:       t.Bid,
				Volume24h: t.Volume,
				Timestamp: ts,
			})
		} else {
			frs[i] = fetchResultWithError(
				p,
				fmt.Errorf("failed to find symbol %s in huobi response", p),
			)
		}
	}

	return frs, nil
}
