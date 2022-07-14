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

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

const upbitBaseURL = "https://api.upbit.com"
const upbitURL = "%s/v1/ticker?markets=%s"

type Upbit struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (o Upbit) Pool() query.WorkerPool {
	return o.WorkerPool
}
func (o Upbit) PullPrices(pairs []Pair) []FetchResult {
	req := &query.HTTPRequest{
		URL: buildOriginURL(upbitURL, o.BaseURL, upbitBaseURL, o.localPairName(pairs...)),
	}
	res := o.Pool().Query(req)
	if errorResponses := validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return o.parseResponse(pairs, res)
}

type upbitTicker struct {
	Market             string               `json:"market"`
	TradeDate          string               `json:"trade_date"`
	TradeTime          string               `json:"trade_time"`
	TradeTimestamp     int64                `json:"trade_timestamp"`
	OpeningPrice       float64              `json:"opening_price"`
	HighPrice          float64              `json:"high_price"`
	LowPrice           float64              `json:"low_price"`
	TradePrice         float64              `json:"trade_price"`
	PrevClosingPrice   float64              `json:"prev_closing_price"`
	Change             string               `json:"change"`
	ChangePrice        float64              `json:"change_price"`
	ChangeRate         float64              `json:"change_rate"`
	SignedChangePrice  float64              `json:"signed_change_price"`
	SignedChangeRate   float64              `json:"signed_change_rate"`
	TradeVolume        float64              `json:"trade_volume"`
	AccTradePrice      float64              `json:"acc_trade_price"`
	AccTradePrice24H   float64              `json:"acc_trade_price_24h"`
	AccTradeVolume     float64              `json:"acc_trade_volume"`
	AccTradeVolume24H  float64              `json:"acc_trade_volume_24h"`
	Highest52WeekPrice float64              `json:"highest_52_week_price"`
	Highest52WeekDate  string               `json:"highest_52_week_date"`
	Lowest52WeekPrice  float64              `json:"lowest_52_week_price"`
	Lowest52WeekDate   string               `json:"lowest_52_week_date"`
	Timestamp          intAsUnixTimestampMs `json:"timestamp"`
}

func (o *Upbit) localPairName(pairs ...Pair) string {
	var l []string
	for _, pair := range pairs {
		l = append(l, fmt.Sprintf("%s-%s", pair.Quote, pair.Base))
	}
	return strings.Join(l, ",")
}

func (o *Upbit) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	var resp []upbitTicker
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse response: %w", err))
	}

	tickers := make(map[string]upbitTicker)
	for _, t := range resp {
		tickers[t.Market] = t
	}

	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		if t, is := tickers[o.localPairName(pair)]; !is {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     t.TradePrice,
					Volume24h: t.AccTradeVolume24H,
					Timestamp: t.Timestamp.val(),
				},
			})
		}
	}
	return results
}
