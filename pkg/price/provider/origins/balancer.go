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

const balancerBaseURL = "https://api.thegraph.com"
const balancerURL = "%s/subgraphs/name/balancer-labs/balancer"

type balancerResponse struct {
	Data struct {
		TokenPrices []balancerPairResponse `json:"tokenPrices"`
	}
}

type balancerPairResponse struct {
	Symbol string          `json:"symbol"`
	Price  stringAsFloat64 `json:"price"`
	Volume stringAsFloat64 `json:"poolLiquidity"`
}

type Balancer struct {
	WorkerPool        query.WorkerPool
	BaseURL           string
	ContractAddresses ContractAddresses
}

func (s Balancer) Pool() query.WorkerPool {
	return s.WorkerPool
}

func (s *Balancer) pairsToContractAddress(pair Pair) (string, error) {
	// We're checking for reverse pairs because the same contract is used to
	// trade in both directions.
	contract, _, ok := s.ContractAddresses.ByPair(pair)
	if !ok {
		return "", fmt.Errorf("failed to find Balancer contract address for pair: %s", pair.String())
	}
	return contract, nil
}

func (s Balancer) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

func (s *Balancer) callOne(pair Pair) (*Price, error) {
	var err error

	contract, err := s.pairsToContractAddress(pair)
	if err != nil {
		return nil, err
	}
	pairsJSON, _ := json.Marshal(contract)
	gql := `
		query($id:String) {
			tokenPrices(where: { id: $id }){
				symbol price poolLiquidity
			}
		}
	`
	body := fmt.Sprintf(
		`{"query":"%s","variables":{"id":%s}}`,
		strings.ReplaceAll(strings.ReplaceAll(gql, "\n", " "), "\t", ""),
		pairsJSON,
	)

	req := &query.HTTPRequest{
		URL:    buildOriginURL(balancerURL, s.BaseURL, balancerBaseURL),
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
	}

	// make query
	res := s.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parse JSON
	var resp balancerResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Balancer response: %w", err)
	}

	pairPrice := resp.Data.TokenPrices[0]
	if pairPrice.Symbol != pair.Base {
		return nil, ErrMissingResponseForPair
	}

	return &Price{
		Pair:      pair,
		Price:     pairPrice.Price.val(),
		Volume24h: pairPrice.Volume.val(),
		Timestamp: time.Now(),
	}, nil
}
