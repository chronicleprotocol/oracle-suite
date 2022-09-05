//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package spire

import (
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var spireAgentFactory = func(cfg spire.AgentConfig) (*spire.Agent, error) {
	return spire.NewAgent(cfg)
}

//nolint
var spireClientFactory = func(cfg spire.ClientConfig) (*spire.Client, error) {
	return spire.NewClient(cfg)
}

//nolint
var priceStoreFactory = func(cfg store.Config) (*store.PriceStore, error) {
	return store.New(cfg)
}

type Spire struct {
	RPC           RPC      `yaml:"rpc"` // Old configuration format, to remove in the future.
	RPCListenAddr string   `yaml:"rpcListenAddr"`
	Pairs         []string `yaml:"pairs"`
}

type RPC struct {
	Address string `yaml:"address"`
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
	listenAddr := c.RPC.Address
	if len(c.RPCListenAddr) != 0 {
		listenAddr = c.RPCListenAddr
	}
	agent, err := spireAgentFactory(spire.AgentConfig{
		PriceStore: d.PriceStore,
		Transport:  d.Transport,
		Signer:     d.Signer,
		Address:    listenAddr,
		Logger:     d.Logger,
	})
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (c *Spire) ConfigureClient(d ClientDependencies) (*spire.Client, error) {
	listenAddr := c.RPC.Address
	if len(c.RPCListenAddr) != 0 {
		listenAddr = c.RPCListenAddr
	}
	return spireClientFactory(spire.ClientConfig{
		Signer:  d.Signer,
		Address: listenAddr,
	})
}

func (c *Spire) ConfigurePriceStore(d PriceStoreDependencies) (*store.PriceStore, error) {
	cfg := store.Config{
		Storage:   store.NewMemoryStorage(),
		Signer:    d.Signer,
		Transport: d.Transport,
		Pairs:     c.Pairs,
		Logger:    d.Logger,
	}
	return priceStoreFactory(cfg)
}
