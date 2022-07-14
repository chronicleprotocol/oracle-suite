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

package eventpublisher

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
)

func TestEventPublisher_Configure_Teleport(t *testing.T) {
	prevEventPublisherFactory := eventPublisherFactory
	defer func() { eventPublisherFactory = prevEventPublisherFactory }()

	sig := geth.NewSigner(nil)
	tra := local.New([]byte("test"), 0, nil)
	_ = tra.Start(context.Background())
	log := null.New()

	config := EventPublisher{Listeners: listeners{TeleportEVM: []teleportEVMListener{{
		Ethereum:    ethereumConfig.Ethereum{RPC: "https://example.com/"},
		Interval:    1,
		BlocksDelta: []int{10, 60},
		BlocksLimit: 10,
		Addresses:   []common.Address{common.HexToAddress("0x07a35a1d4b751a818d93aa38e615c0df23064881")},
	}}}}

	eventPublisherFactory = func(cfg publisher.Config) (*publisher.EventPublisher, error) {
		assert.Equal(t, tra, cfg.Transport)
		assert.NotNil(t, cfg.Signers)
		assert.Equal(t, log, cfg.Logger)
		assert.Len(t, cfg.Listeners, 1)
		assert.Len(t, cfg.Signers, 1)
		return &publisher.EventPublisher{}, nil
	}

	ep, err := config.Configure(Dependencies{
		Signer:    sig,
		Transport: tra,
		Logger:    log,
	})
	require.NoError(t, err)
	require.NotNil(t, ep)
}

func Test_ethClients_configure(t *testing.T) {
	c := &ethClients{}

	c1, err := c.configure(ethereumConfig.Ethereum{RPC: "https://example.com/"}, null.New())
	require.NoError(t, err)
	c2, err := c.configure(ethereumConfig.Ethereum{RPC: "https://example.com/"}, null.New())
	require.NoError(t, err)
	c3, err := c.configure(ethereumConfig.Ethereum{RPC: "https://example.com/", MaxBlocksBehind: 10}, null.New())
	require.NoError(t, err)

	assert.Same(t, c1, c2)
	assert.NotSame(t, c1, c3)
}
