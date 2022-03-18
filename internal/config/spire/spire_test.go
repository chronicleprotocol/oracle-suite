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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	datastoreMemory "github.com/chronicleprotocol/oracle-suite/pkg/datastore/memory"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
)

func TestSpire_ConfigureAgent(t *testing.T) {
	prevSpireAgentFactory := spireAgentFactory
	defer func() {
		spireAgentFactory = prevSpireAgentFactory
	}()

	signer := &ethereumMocks.Signer{}
	transport := local.New([]byte("test"), 0, nil)
	feeds := []ethereum.Address{ethereum.HexToAddress("0x07a35a1d4b751a818d93aa38e615c0df23064881")}
	logger := null.New()
	ds := &datastoreMemory.Datastore{}

	config := Spire{
		RPC:   RPC{Address: "1.2.3.4:1234"},
		Pairs: []string{"AAABBB"},
	}

	spireAgentFactory = func(cfg spire.AgentConfig) (*spire.Agent, error) {
		assert.Equal(t, ds, cfg.Datastore)
		assert.Equal(t, transport, cfg.Transport)
		assert.Equal(t, signer, cfg.Signer)
		assert.Equal(t, "1.2.3.4:1234", cfg.Address)
		assert.Equal(t, logger, cfg.Logger)
		return &spire.Agent{}, nil
	}

	a, err := config.ConfigureAgent(AgentDependencies{
		Signer:    signer,
		Transport: transport,
		Datastore: ds,
		Feeds:     feeds,
		Logger:    logger,
	})
	require.NoError(t, err)
	require.NotNil(t, a)
}

func TestSpire_ConfigureClient(t *testing.T) {
	prevSpireClientFactory := spireClientFactory
	defer func() { spireClientFactory = prevSpireClientFactory }()

	signer := &ethereumMocks.Signer{}

	config := Spire{
		RPC:   RPC{Address: "1.2.3.4:1234"},
		Pairs: []string{"AAABBB"},
	}

	spireClientFactory = func(cfg spire.ClientConfig) (*spire.Client, error) {
		assert.Equal(t, signer, cfg.Signer)
		assert.Equal(t, "1.2.3.4:1234", cfg.Address)
		return &spire.Client{}, nil
	}

	c, err := config.ConfigureClient(ClientDependencies{
		Signer: signer,
	})
	require.NoError(t, err)
	require.NotNil(t, c)
}
