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

package starknet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Sequencer struct {
	endpoint   string
	httpClient http.Client
}

func NewSequencer(endpoint string, httpClient http.Client) *Sequencer {
	return &Sequencer{endpoint: endpoint, httpClient: httpClient}
}

func (s *Sequencer) GetPendingBlock(ctx context.Context) (*Block, error) {
	return s.getBlock(ctx, "pending")
}

func (s *Sequencer) GetLatestBlock(ctx context.Context) (*Block, error) {
	return s.getBlock(ctx, "null")
}

func (s *Sequencer) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*Block, error) {
	return s.getBlock(ctx, fmt.Sprintf("%d", blockNumber))
}

func (s *Sequencer) getBlock(ctx context.Context, blockNumber string) (*Block, error) {
	url := fmt.Sprintf("%s/feeder_gateway/get_block?blockNumber=%s", s.endpoint, blockNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var block *Block
	err = json.Unmarshal(body, &block)
	if err != nil {
		return nil, err
	}
	return block, nil
}
