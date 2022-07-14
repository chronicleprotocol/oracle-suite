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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

const uniswapV3BaseURL = "https://api.thegraph.com"
const uniswapV3URL = "%s/subgraphs/name/uniswap/uniswap-v3"

type uniswapV3Response struct {
	Data struct {
		Pools []uniswapV3PairResponse
	}
}

type uniswapV3TokenResponse struct {
	Symbol string `json:"symbol"`
}

type uniswapV3PairResponse struct {
	ID      string                 `json:"id"`
	Price0  stringAsFloat64        `json:"token0Price"`
	Price1  stringAsFloat64        `json:"token1Price"`
	Volume0 stringAsFloat64        `json:"volumeToken0"`
	Volume1 stringAsFloat64        `json:"volumeToken1"`
	Token0  uniswapV3TokenResponse `json:"token0"`
	Token1  uniswapV3TokenResponse `json:"token1"`
}

type UniswapV3 struct {
	WorkerPool        query.WorkerPool
	BaseURL           string
	ContractAddresses ContractAddresses
}

func (u UniswapV3) Pool() query.WorkerPool {
	return u.WorkerPool
}

func (u UniswapV3) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&u, pairs)
}

func (u *UniswapV3) callOne(pair Pair) (*Price, error) {
	var err error

	contract, _, ok := u.ContractAddresses.ByPair(pair)
	if !ok {
		return nil, fmt.Errorf("failed to find contract address for pair: %s", pair.String())
	}

	pairsJSON, _ := json.Marshal(contract)
	gql := `
		query($id:String) {
			pools(where:{id: $id}) {
				id
				token0Price
				token1Price
				volumeToken0
				volumeToken1
				token0 { symbol }
				token1 { symbol }
			}
		}
	`
	body := fmt.Sprintf(
		`{"query":"%s","variables":{"id":%s}}`,
		strings.ReplaceAll(strings.ReplaceAll(gql, "\n", " "), "\t", ""),
		pairsJSON,
	)

	req := &query.HTTPRequest{
		URL:    buildOriginURL(uniswapV3URL, u.BaseURL, uniswapV3BaseURL),
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
	}

	// make query
	res := u.WorkerPool.Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parse JSON
	var resp uniswapV3Response
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UniswapV3 response: %w", err)
	}

	// convert response from a slice to a map
	respMap := map[string]uniswapV3PairResponse{}
	for _, pairResp := range resp.Data.Pools {
		respMap[pairResp.Token0.Symbol+"/"+pairResp.Token1.Symbol] = pairResp
	}

	b := pair.Base
	q := pair.Quote

	pair0 := b + "/" + q
	pair1 := q + "/" + b

	if r, ok := respMap[pair0]; ok {
		return &Price{
			Pair:      pair,
			Price:     r.Price1.val(),
			Bid:       r.Price1.val(),
			Ask:       r.Price1.val(),
			Volume24h: r.Volume0.val(),
			Timestamp: time.Now(),
		}, nil
	} else if r, ok := respMap[pair1]; ok {
		return &Price{
			Pair:      pair,
			Price:     r.Price0.val(),
			Bid:       r.Price0.val(),
			Ask:       r.Price0.val(),
			Volume24h: r.Volume1.val(),
			Timestamp: time.Now(),
		}, nil
	}
	return nil, ErrMissingResponseForPair
}
