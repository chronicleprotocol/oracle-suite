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

// Exchange URL
const bitThumpBaseURL = "https://global-openapi.bithumb.pro"
const bitThumpURL = "%s/openapi/v1/spot/ticker?symbol=%s"

type bitThumbPriceResponse struct {
	Low    stringAsFloat64 `json:"l"`
	High   stringAsFloat64 `json:"h"`
	Last   stringAsFloat64 `json:"c"`
	Symbol string          `json:"s"`
	Volume stringAsFloat64 `json:"v"`
}
type bitThumbResponse struct {
	Data      []bitThumbPriceResponse `json:"data"`
	Code      string                  `json:"code"`
	Msg       string                  `json:"msg"`
	Timestamp intAsUnixTimestampMs    `json:"timestamp"`
}

// Bithumb origin handler
type BitThump struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (c *BitThump) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(pair.Base), strings.ToUpper(pair.Quote))
}

func (c *BitThump) getURL(pair Pair) string {
	return buildOriginURL(bitThumpURL, c.BaseURL, bitThumpBaseURL, c.localPairName(pair))
}

func (c BitThump) Pool() query.WorkerPool {
	return c.WorkerPool
}

func (c BitThump) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&c, pairs)
}

func (c *BitThump) callOne(pair Pair) (*Price, error) {
	var err error
	req := &query.HTTPRequest{
		URL: c.getURL(pair),
	}

	// make query
	res := c.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp bitThumbResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Bithumb response: %w", err)
	}
	if resp.Code != "0" || resp.Msg != "success" || len(resp.Data) != 1 {
		return nil, fmt.Errorf("invalid Bithumb response: %s", res.Body)
	}
	priceResp := resp.Data[0]
	// building Price
	return &Price{
		Pair:      pair,
		Price:     priceResp.Last.val(),
		Volume24h: priceResp.Volume.val(),
		Timestamp: resp.Timestamp.val(),
	}, nil
}
