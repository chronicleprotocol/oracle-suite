package webapi

import (
	"context"
	"sync"
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
)

// AddressBook provides a list of addresses to which the messages should be
// sent.
type AddressBook interface {
	Consumers(ctx context.Context) ([]string, error)
}

// MultiAddressBook is an implementation of AddressBook that merges the
// addresses from multiple AddressBook instances.
type MultiAddressBook struct {
	books []AddressBook
}

// NewMultiAddressBook creates a new instance of MultiAddressBook.
func NewMultiAddressBook(books ...AddressBook) *MultiAddressBook {
	return &MultiAddressBook{
		books: books,
	}
}

// Consumers implements the AddressBook interface.
func (m *MultiAddressBook) Consumers(ctx context.Context) ([]string, error) {
	var addresses []string
	for _, book := range m.books {
		toMerge, err := book.Consumers(ctx)
		if err != nil {
			return nil, err
		}
		for _, addr1 := range toMerge {
			found := false
			for _, addr2 := range addresses {
				if addr1 == addr2 {
					found = true
					break
				}
			}
			if !found {
				addresses = append(addresses, addr1)
			}
		}
	}
	return addresses, nil
}

// StaticAddressBook is an implementation of AddressBook that returns a static
// list of addresses.
type StaticAddressBook struct {
	addresses []string
}

// NewStaticAddressBook creates a new instance of StaticAddressBook.
func NewStaticAddressBook(addresses []string) *StaticAddressBook {
	return &StaticAddressBook{
		addresses: addresses,
	}
}

// Consumers implements the AddressBook interface.
func (c *StaticAddressBook) Consumers(_ context.Context) ([]string, error) {
	return c.addresses, nil
}

// EthereumAddressBook is an AddressBook implementation that uses an Ethereum
// contract to store the list of addresses.
type EthereumAddressBook struct {
	mu sync.Mutex

	client    rpc.RPC       // Ethereum client
	address   types.Address // Address of the contract.
	cache     []string      // Cached list of addresses.
	cacheTime time.Time     // Time when the cache was last updated.
	cacheTTL  time.Duration // How long the cache should be valid.
}

// NewEthereumAddressBook creates a new instance of EthereumAddressBook.
// The cacheTTL parameter specifies how long the list of addresses should be
// cached before it is fetched again from the Ethereum contract.
func NewEthereumAddressBook(r rpc.RPC, addr types.Address, cacheTTL time.Duration) *EthereumAddressBook {
	return &EthereumAddressBook{
		client:   r,
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
	cd, err := consumersMethod.EncodeArgs()
	if err != nil {
		return nil, err
	}
	res, err := c.client.Call(ctx, types.Call{
		To:    &c.address,
		Input: cd,
	}, types.LatestBlockNumber)
	if err != nil {
		return nil, err
	}
	var addrs []string
	err = consumersMethod.DecodeValues(res, &addrs)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}

var consumersMethod = abi.MustParseMethod("function list() returns (string[])")
