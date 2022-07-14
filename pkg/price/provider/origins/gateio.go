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

// Gateio URL
const gateioBaseURL = "https://fx-api.gateio.ws"
const gateioSinglePairURL = "%s/api/v4/spot/tickers?currency_pair=%s"
const gateioURL = "%s/api/v4/spot/tickers"

type gateioResponse struct {
	Pair   string `json:"currency_pair"`
	Volume string `json:"quote_volume"`
	Price  string `json:"last"`
	Ask    string `json:"lowest_ask"`
	Bid    string `json:"highest_bid"`
}

// Gateio exchange handler
type Gateio struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (g *Gateio) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

func (g *Gateio) localPairName(pair Pair) string {
	return fmt.Sprintf("%s_%s", g.renameSymbol(pair.Base), g.renameSymbol(pair.Quote))
}

func (g Gateio) Pool() query.WorkerPool {
	return g.WorkerPool
}

func (g Gateio) PullPrices(pairs []Pair) []FetchResult {
	crs, err := g.fetch(pairs)
	if err != nil {
		return fetchResultListWithErrors(pairs, err)
	}
	return crs
}

func (g *Gateio) fetch(pairs []Pair) ([]FetchResult, error) {
	var url string
	if len(pairs) == 1 {
		url = buildOriginURL(gateioSinglePairURL, g.BaseURL, gateioBaseURL, g.localPairName(pairs[0]))
	} else {
		url = buildOriginURL(gateioURL, g.BaseURL, gateioBaseURL)
	}

	req := &query.HTTPRequest{URL: url}

	// make query
	res := g.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resps []gateioResponse
	err := json.Unmarshal(res.Body, &resps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gateio response: %w", err)
	}
	if len(resps) < 1 {
		return nil, fmt.Errorf("wrong gateio response: %s", res.Body)
	}

	respMap := map[string]gateioResponse{}
	for _, resp := range resps {
		respMap[resp.Pair] = resp
	}

	frs := make([]FetchResult, len(pairs))
	for i, pair := range pairs {
		symbol := g.localPairName(pair)
		if resp, has := respMap[symbol]; has {
			price, err := g.newPrice(pair, resp)
			if err != nil {
				frs[i] = fetchResultWithError(
					pair,
					fmt.Errorf("failed to create price point from gateio response: %w: %s", err, res.Body),
				)
			} else {
				frs[i] = fetchResult(price)
			}
		} else {
			frs[i] = fetchResultWithError(
				pair,
				fmt.Errorf("failed to find symbol %s in gateio response: %s", pair, res.Body),
			)
		}
	}
	return frs, nil
}

func (g *Gateio) newPrice(pair Pair, resp gateioResponse) (Price, error) {
	// Check pair name
	if resp.Pair != g.localPairName(pair) {
		return Price{}, fmt.Errorf("wrong gateio pair returned: %s", resp.Pair)
	}

	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return Price{}, fmt.Errorf("failed to parse price from gateio exchange")
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return Price{}, fmt.Errorf("failed to parse volume from gateio exchange")
	}
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return Price{}, fmt.Errorf("failed to parse ask from gateio exchange")
	}
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return Price{}, fmt.Errorf("failed to parse bid from gateio exchange")
	}

	// building Price
	return Price{
		Pair:      pair,
		Price:     price,
		Ask:       ask,
		Bid:       bid,
		Volume24h: volume,
		Timestamp: time.Now(),
	}, nil
}
