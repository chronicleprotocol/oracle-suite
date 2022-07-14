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

// Bitstamp URL
const bitstampBaseURL = "https://www.bitstamp.net"
const bitstampURL = "%s/api/v2/ticker/%s"

type bitstampResponse struct {
	Ask       string `json:"ask"`
	Volume    string `json:"volume"`
	Price     string `json:"last"`
	Bid       string `json:"bid"`
	Timestamp string `json:"timestamp"`
}

// Bitstamp origin handler
type Bitstamp struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (b *Bitstamp) renameSymbol(symbol string) string {
	return symbol
}

func (b *Bitstamp) localPairName(pair Pair) string {
	return strings.ToLower(b.renameSymbol(pair.Base) + b.renameSymbol(pair.Quote))
}

func (b *Bitstamp) getURL(pair Pair) string {
	return buildOriginURL(bitstampURL, b.BaseURL, bitstampBaseURL, b.localPairName(pair))
}

func (b Bitstamp) Pool() query.WorkerPool {
	return b.WorkerPool
}

func (b Bitstamp) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&b, pairs)
}

func (b *Bitstamp) callOne(pair Pair) (*Price, error) {
	var err error
	req := &query.HTTPRequest{
		URL: b.getURL(pair),
	}

	// make query
	res := b.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp bitstampResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bitstamp response: %w", err)
	}

	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from bitstamp origin %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from bitstamp origin %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from bitstamp origin %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from bitstamp origin %s", res.Body)
	}
	// Parsing timestamp from string
	timestamp, err := strconv.ParseInt(resp.Timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp from bitstamp origin %s", res.Body)
	}
	// building Price
	return &Price{
		Pair:      pair,
		Price:     price,
		Volume24h: volume,
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Unix(timestamp, 0),
	}, nil
}
