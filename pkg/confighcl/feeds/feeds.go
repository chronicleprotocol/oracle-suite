package feeds

import (
	"errors"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type Feeds struct {
	Addresses []string `hcl:"addresses"`
}

var ErrInvalidEthereumAddress = errors.New("invalid ethereum address")

func (f *Feeds) ConfigureAddresses() ([]ethereum.Address, error) {
	var addrs []ethereum.Address
	for _, addr := range f.Addresses {
		if !ethereum.IsHexAddress(addr) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidEthereumAddress, addr)
		}
		addrs = append(addrs, ethereum.HexToAddress(addr))
	}
	return addrs, nil
}
