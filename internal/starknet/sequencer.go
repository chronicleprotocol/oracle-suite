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
