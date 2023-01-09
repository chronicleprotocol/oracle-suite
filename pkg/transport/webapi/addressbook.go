package webapi

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

// AddressBook provides a list of addresses to which the messages should be
// sent.
type AddressBook interface {
	Consumers(ctx context.Context) ([]string, error)
}

// EthereumAddressBook is an AddressBook implementation that uses an Ethereum
// contract to store the list of addresses.
type EthereumAddressBook struct {
	mu sync.Mutex

	client    ethereum.Client  // Ethereum client
	address   ethereum.Address // Address of the contract.
	cache     []string         // Cached list of addresses.
	cacheTime time.Time        // Time when the cache was last updated.
	cacheTTL  time.Duration    // How long the cache should be valid.
}

// NewEthereumAddressBook creates a new instance of EthereumAddressBook.
// The cacheTTL parameter specifies how long the list of addresses should be
// cached before it is fetched again from the Ethereum contract.
func NewEthereumAddressBook(c ethereum.Client, addr ethereum.Address, cacheTTL time.Duration) *EthereumAddressBook {
	return &EthereumAddressBook{
		client:   c,
		address:  addr,
		cacheTTL: cacheTTL,
	}
}

// Consumers implements the AddressBook interface.
func (c *EthereumAddressBook) Consumers(ctx context.Context) ([]string, error) {
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

func (c *EthereumAddressBook) fetchConsumers(ctx context.Context) ([]string, error) {
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
	// Addresses on a smart contract may omit protocol scheme, so we add it
	// here.
	var addrs []string
	for n, addr := range ret[0].([]string) {
		if !strings.Contains(addr, "://") {
			// Data transmitted over the WebAPI protocol is signed, hence
			// there is no need to use HTTPS.
			addrs[n] = "http://" + addr
		}
	}
	return addrs, nil
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
