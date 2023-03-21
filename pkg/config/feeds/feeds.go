package feeds

import (
	"fmt"

	"github.com/defiweb/go-eth/types"
)

// ConfigFeeds is a list of feed addresses.
type ConfigFeeds []string

func (f ConfigFeeds) Addresses() ([]types.Address, error) {
	var addrs []types.Address
	for _, hexAddr := range f {
		addr, err := types.AddressFromHex(hexAddr)
		if err != nil {
			return nil, fmt.Errorf("feeds config: invalid address: %s", hexAddr)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}
