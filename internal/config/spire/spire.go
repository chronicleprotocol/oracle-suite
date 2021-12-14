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
	"context"

	"github.com/makerdao/oracle-suite/pkg/datastore"
	datastoreMemory "github.com/makerdao/oracle-suite/pkg/datastore/memory"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/spire"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

//nolint
var spireAgentFactory = func(ctx context.Context, cfg spire.AgentConfig) (*spire.Agent, error) {
	return spire.NewAgent(ctx, cfg)
}

//nolint
var spireClientFactory = func(ctx context.Context, cfg spire.ClientConfig) (*spire.Client, error) {
	return spire.NewClient(ctx, cfg)
}

var datastoreFactory = func(ctx context.Context, cfg datastoreMemory.Config) (datastore.Datastore, error) {
	return datastoreMemory.NewDatastore(ctx, cfg)
}

type Spire struct {
	RPC            RPC      `json:"rpc"`
	Pairs          []string `json:"pairs"`
	TransportToUse string   `json:"transport"`
}

type RPC struct {
	Address string `json:"address"`
}

type AgentDependencies struct {
	Context   context.Context
	Signer    ethereum.Signer
	Transport transport.Transport
	Datastore datastore.Datastore
	Feeds     []ethereum.Address
	Logger    log.Logger
}

type ClientDependencies struct {
	Context context.Context
	Signer  ethereum.Signer
}

type DatastoreDependencies struct {
	Context   context.Context
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *Spire) ConfigureAgent(d AgentDependencies) (*spire.Agent, error) {
	agent, err := spireAgentFactory(d.Context, spire.AgentConfig{
		Datastore: d.Datastore,
		Transport: d.Transport,
		Signer:    d.Signer,
		Address:   c.RPC.Address,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (c *Spire) ConfigureClient(d ClientDependencies) (*spire.Client, error) {
	return spireClientFactory(d.Context, spire.ClientConfig{
		Signer:  d.Signer,
		Address: c.RPC.Address,
	})
}

func (c *Spire) ConfigureDatastore(d DatastoreDependencies) (datastore.Datastore, error) {
	cfg := datastoreMemory.Config{
		Signer:    d.Signer,
		Transport: d.Transport,
		Pairs:     make(map[string]*datastoreMemory.Pair),
		Logger:    d.Logger,
	}
	for _, name := range c.Pairs {
		cfg.Pairs[name] = &datastoreMemory.Pair{Feeds: d.Feeds}
	}
	return datastoreFactory(d.Context, cfg)
}

const TransportLibP2P = "libp2p"
const TransportLibSSB = "ssb"
const DefaultTransport = TransportLibP2P
