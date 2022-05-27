package starknet

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/chronicleprotocol/oracle-suite/internal/util/retry"
)

const retryAttempts = 3               // The maximum number of attempts to call EthClient in case of an error.
const retryInterval = 5 * time.Second // The delay between retry attempts.

type Scope string

const (
	ScopeTXNHash            Scope = "TXN_HASH"
	ScopeFullTXNs           Scope = "FULL_TXNS"
	ScopeFullTXNAndReceipts Scope = "FULL_TXN_AND_RECEIPTS"
)

type Client struct {
	rpcClient *rpc.Client
}

func NewClient(endpoint string) (*Client, error) {
	rpcClient, err := rpc.DialContext(context.Background(), endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{rpcClient: rpcClient}, nil
}

func (s *Client) BlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber uint64
	err := s.call(ctx, &blockNumber, "starknet_blockNumber")
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}

func (s *Client) GetBlockByNumber(ctx context.Context, blockNumber uint64, scope Scope) (*Block, error) {
	var block *Block
	err := s.call(ctx, &block, "starknet_getBlockByNumber", blockNumber, scope)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (s *Client) call(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	return retry.Retry(ctx, func() error {
		return s.rpcClient.CallContext(ctx, result, method, args...)
	}, retryAttempts, retryInterval)
}
