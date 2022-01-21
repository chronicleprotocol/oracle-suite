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

package eventapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/event/api"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
)

func TestEventAPI_Configure(t *testing.T) {
	prevEventAPIFactory := eventAPIFactory
	defer func() { eventAPIFactory = prevEventAPIFactory }()

	ctx := context.Background()
	tra := local.New(ctx, 0, nil)
	log := null.New()
	evs := &store.EventStore{}

	config := EventAPI{
		Address: "127.0.0.1:0",
	}

	eventAPIFactory = func(ctx context.Context, cfg api.Config) (*api.EventAPI, error) {
		assert.NotNil(t, ctx)
		assert.Equal(t, evs, cfg.EventStore)
		assert.Equal(t, config.Address, cfg.Address)
		assert.Equal(t, log, cfg.Logger)

		return &api.EventAPI{}, nil
	}

	a, err := config.Configure(Dependencies{
		Context:    ctx,
		EventStore: evs,
		Transport:  tra,
		Logger:     log,
	})
	require.NoError(t, err)
	assert.NotNil(t, a)
}
