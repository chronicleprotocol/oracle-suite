package tor

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type Consumers interface {
	// Consumers returns a list of consumer addresses.
	Consumers(ctx context.Context) ([]string, error)
}

type ContractConsumers struct {
	mu sync.Mutex

	client    ethereum.Client
	address   ethereum.Address
	cache     []string
	cacheTime time.Time
	cacheTTL  time.Duration
}

func NewContractConsumers(client ethereum.Client, address ethereum.Address, cacheTTL time.Duration) *ContractConsumers {
	return &ContractConsumers{
		client:   client,
		address:  address,
		cacheTTL: cacheTTL,
	}
}

func (c *ContractConsumers) Consumers(ctx context.Context) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil || c.cacheTime.Add(c.cacheTTL).Before(time.Now()) {
		addrs, err := c.fetchConsumers(ctx)
		if err != nil {
			return nil, err
		}
		c.cache = addrs
		c.cacheTime = time.Now()
	}
	return c.cache, nil
}

func (c *ContractConsumers) fetchConsumers(ctx context.Context) ([]string, error) {
	cd, err := consumersABI.Pack("getConsumers")
	if err != nil {
		return nil, err
	}
	res, err := c.client.Call(ctx, ethereum.Call{
		Address: c.address,
		Data:    cd,
	})
	if err != nil {
		return nil, err
	}
	ret, err := consumersABI.Unpack("getConsumers", res)
	if err != nil {
		return nil, err
	}
	return ret[0].([]string), nil
}

const consumersJSONABI = `
[
  {
    "inputs": [],
    "name": "getConsumers",
    "outputs": [
      {
        "internalType": "string[]",
        "name": "",
        "type": "string[]"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]
`

var consumersABI abi.ABI

func init() {
	var err error
	consumersABI, err = abi.JSON(strings.NewReader(consumersJSONABI))
	if err != nil {
		panic(err.Error())
	}
}
