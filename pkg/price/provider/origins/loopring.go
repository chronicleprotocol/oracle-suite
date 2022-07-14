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

// Loopring URL
const loopringBaseURL = "https://api3.loopring.io"
const loopringURL = "%s/api/v3/ticker?market=%s"

type loopringResponse struct {
	Tickers [][]string `json:"tickers"`
}

// Loopring origin handler
type Loopring struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (l *Loopring) localPairName(pairs ...Pair) string {
	var list []string
	for _, pair := range pairs {
		list = append(list, fmt.Sprintf("%s-%s", strings.ToUpper(pair.Base), strings.ToUpper(pair.Quote)))
	}
	return strings.Join(list, ",")
}

func (l Loopring) Pool() query.WorkerPool {
	return l.WorkerPool
}

func (l Loopring) PullPrices(pairs []Pair) []FetchResult {
	var err error
	req := &query.HTTPRequest{
		URL: buildOriginURL(loopringURL, l.BaseURL, loopringBaseURL, l.localPairName(pairs...)),
	}
	// make query
	res := l.Pool().Query(req)
	if res == nil {
		return fetchResultListWithErrors(pairs, ErrEmptyOriginResponse)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, res.Error)
	}
	// parsing JSON
	var resp loopringResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse loopring response: %w", err))
	}
	if len(resp.Tickers) != len(pairs) {
		return fetchResultListWithErrors(pairs, fmt.Errorf("wrong loopring response %s", res.Body))
	}

	var results []FetchResult
	for _, pair := range pairs {
		results = append(results, l.pickPairDetails(resp, pair))
	}
	return results
}

func (l *Loopring) pickPairDetails(response loopringResponse, pair Pair) FetchResult {
	pairName := l.localPairName(pair)
	var pairRes []string
	for _, pairResponse := range response.Tickers {
		if len(pairResponse) > 0 && pairResponse[0] == pairName {
			pairRes = pairResponse
			break
		}
	}
	if pairRes == nil {
		return fetchResultWithError(pair, fmt.Errorf("no %s pair exist in loopring response", pair))
	}
	if len(pairRes) < 10 {
		return fetchResultWithError(pair, fmt.Errorf("invalid pair response for pair %s: %v", pair, pairRes))
	}
	timestamp, err := strconv.ParseInt(pairRes[1], 10, 64)
	if err != nil {
		return fetchResultWithError(pair, fmt.Errorf("failed to parse timestamp for pair %s: %w", pair, err))
	}

	price, err := strconv.ParseFloat(pairRes[7], 64)
	if err != nil {
		return fetchResultWithError(pair, fmt.Errorf("failed to parse price for pair %s: %w", pair, err))
	}
	bid, err := strconv.ParseFloat(pairRes[9], 64)
	if err != nil {
		return fetchResultWithError(pair, fmt.Errorf("failed to parse bid for pair %s: %w", pair, err))
	}
	ask, err := strconv.ParseFloat(pairRes[10], 64)
	if err != nil {
		return fetchResultWithError(pair, fmt.Errorf("failed to parse ask for pair %s: %w", pair, err))
	}
	// building Price
	return fetchResult(Price{
		Pair:      pair,
		Price:     price,
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Unix(timestamp, 0),
	})
}
