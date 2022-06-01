package starknet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Sequencer struct {
	endpoint string
	client   http.Client
}

func NewSequencer(endpoint string, client http.Client) *Sequencer {
	return &Sequencer{endpoint: endpoint, client: client}
}

func (s *Sequencer) GetBlockByNumber(ctx context.Context, blockNumber *uint64) (*Block, error) {
	var url string
	if blockNumber == nil {
		url = fmt.Sprintf("%s/feeder_gateway/get_block?blockNumber=null", s.endpoint)
	} else {
		url = fmt.Sprintf("%s/feeder_gateway/get_block?blockNumber=%d", s.endpoint, *blockNumber)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	res, err := s.client.Do(req)
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
