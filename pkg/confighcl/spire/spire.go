package spire

import (
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

type Spire struct {
	RPCListenAddr string   `hcl:"rpc_listen_addr"`
	Pairs         []string `hcl:"pairs"`
}

type AgentDependencies struct {
	Signer     ethereum.Signer
	Transport  transport.Transport
	PriceStore *store.PriceStore
	Feeds      []ethereum.Address
	Logger     log.Logger
}

type ClientDependencies struct {
	Signer ethereum.Signer
}

type PriceStoreDependencies struct {
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *Spire) ConfigureAgent(d AgentDependencies) (*spire.Agent, error) {
	agent, err := spire.NewAgent(spire.AgentConfig{
		PriceStore: d.PriceStore,
		Transport:  d.Transport,
		Signer:     d.Signer,
		Address:    c.RPCListenAddr,
		Logger:     d.Logger,
	})
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (c *Spire) ConfigureClient(d ClientDependencies) (*spire.Client, error) {
	return spire.NewClient(spire.ClientConfig{
		Signer:  d.Signer,
		Address: c.RPCListenAddr,
	})
}

func (c *Spire) ConfigurePriceStore(d PriceStoreDependencies) (*store.PriceStore, error) {
	return store.New(store.Config{
		Storage:   store.NewMemoryStorage(),
		Signer:    d.Signer,
		Transport: d.Transport,
		Pairs:     c.Pairs,
		Logger:    d.Logger,
	})
}
