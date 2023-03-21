package spire

import (
	"fmt"

	"github.com/defiweb/go-eth/types"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

type AgentDependencies struct {
	Keys       ethereumConfig.KeyRegistry
	Transport  transport.Transport
	PriceStore *store.PriceStore
	Feeds      []types.Address
	Logger     log.Logger
}

type ClientDependencies struct {
	KeyRegistry ethereumConfig.KeyRegistry
}

type PriceStoreDependencies struct {
	Transport transport.Transport
	Feeds     []types.Address
	Logger    log.Logger
}

type ConfigSpire struct {
	// RPCListenAddr is an address to listen for RPC requests.
	RPCListenAddr string `hcl:"rpc_listen_addr"`

	// Pairs is a list of pairs to store in the price store.
	Pairs []string `hcl:"pairs"`

	// EthereumKey is a name of an Ethereum key to use for signing
	// prices.
	EthereumKey string `hcl:"ethereum_key,optional"`
}

func (c *ConfigSpire) ConfigureAgent(d AgentDependencies) (*spire.Agent, error) {
	signer, ok := d.Keys[c.EthereumKey]
	if !ok {
		return nil, fmt.Errorf("spire config: ethereum key %q not found", c.EthereumKey)
	}
	agent, err := spire.NewAgent(spire.AgentConfig{
		PriceStore: d.PriceStore,
		Transport:  d.Transport,
		Signer:     signer,
		Address:    c.RPCListenAddr,
		Logger:     d.Logger,
	})
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (c *ConfigSpire) ConfigureClient(d ClientDependencies) (*spire.Client, error) {
	signer := d.KeyRegistry[c.EthereumKey] // Signer may be nil.
	return spire.NewClient(spire.ClientConfig{
		Signer:  signer,
		Address: c.RPCListenAddr,
	})
}

func (c *ConfigSpire) ConfigurePriceStore(d PriceStoreDependencies) (*store.PriceStore, error) {
	return store.New(store.Config{
		Storage:   store.NewMemoryStorage(),
		Transport: d.Transport,
		Pairs:     c.Pairs,
		Logger:    d.Logger,
	})
}
