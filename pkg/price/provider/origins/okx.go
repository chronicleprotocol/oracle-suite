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

const okxBaseURL = "https://www.okx.com"
const okxURL = "%s/api/v5/market/ticker?instId=%s"

type okxResponseData struct {
	InstrumentID  string                  `json:"instId"`
	Last          stringAsFloat64         `json:"last"`
	BestAsk       stringAsFloat64         `json:"askPx"`
	BestBid       stringAsFloat64         `json:"bidPx"`
	BaseVolume24H stringAsFloat64         `json:"vol24h"`
	Timestamp     stringAsUnixTimestampMs `json:"ts"`
}

type okxResponse struct {
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
	Data []okxResponseData `json:"data"`
}

// Okx origin handler
type Okx struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (o Okx) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s-SWAP", pair.Base, pair.Quote)
}

func (o Okx) Pool() query.WorkerPool {
	return o.WorkerPool
}

func (o Okx) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&o, pairs)
}

func (o Okx) callOne(pair Pair) (*Price, error) {
	var err error
	req := &query.HTTPRequest{
		URL: buildOriginURL(okxURL, o.BaseURL, okxBaseURL, o.localPairName(pair)),
	}

	// make query
	res := o.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parse JSON
	var resp okxResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Okx response: %w", err)
	}

	if resp.Code != "0" {
		return nil, fmt.Errorf("okx response code is invalid: %s", resp.Code)
	}

	if len(resp.Data) != 1 {
		return nil, ErrMissingResponseForPair
	}

	data := resp.Data[0]

	return &Price{
		Pair:      pair,
		Price:     data.Last.val(),
		Volume24h: data.BaseVolume24H.val(),
		Timestamp: data.Timestamp.val(),
		Ask:       data.BestAsk.val(),
		Bid:       data.BestBid.val(),
	}, nil
}
