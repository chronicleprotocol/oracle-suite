package origin

import (
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type ContractAddresses map[string]string

type TickGenericWeb3Options struct {
	Protocol          string
	Clients           []*ethereum.Client
	ContractAddresses []ContractAddresses
}

type TickGenericWeb3 struct {
}
