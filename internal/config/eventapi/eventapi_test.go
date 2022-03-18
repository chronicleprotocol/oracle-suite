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
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/event/api"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store/memory"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store/redis"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
)

func TestEventAPI_Configure(t *testing.T) {
	prevEventAPIFactory := eventAPIFactory
	defer func() { eventAPIFactory = prevEventAPIFactory }()

	tra := local.New([]byte("test"), 0, nil)
	log := null.New()
	evs := &store.EventStore{}

	config := EventAPI{
		ListenAddr: "127.0.0.1:0",
	}

	eventAPIFactory = func(cfg api.Config) (*api.EventAPI, error) {
		assert.Equal(t, evs, cfg.EventStore)
		assert.Equal(t, config.ListenAddr, cfg.Address)
		assert.Equal(t, log, cfg.Logger)
		return &api.EventAPI{}, nil
	}

	a, err := config.Configure(Dependencies{
		EventStore: evs,
		Transport:  tra,
		Logger:     log,
	})
	require.NoError(t, err)
	assert.NotNil(t, a)
}

func TestEventAPI_ConfigureStorage_memory(t *testing.T) {
	config := EventAPI{
		Storage: storage{
			Type: "memory",
		},
	}
	sto, err := config.ConfigureStorage()
	require.NoError(t, err)
	require.IsType(t, &memory.Memory{}, sto)
}

func TestEventAPI_ConfigureStorage_redis(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	pass := os.Getenv("TEST_REDIS_PASS")
	db, _ := strconv.Atoi(os.Getenv("TEST_REDIS_DB"))
	if len(addr) == 0 {
		t.Skip()
		return
	}
	config := EventAPI{
		Storage: storage{
			Type: "redis",
			Redis: storageRedis{
				TTL:      60,
				Address:  addr,
				Password: pass,
				DB:       db,
			},
		},
	}
	sto, err := config.ConfigureStorage()
	require.NoError(t, err)
	require.IsType(t, &redis.Redis{}, sto)
}

func TestEventAPI_ConfigureStorage_invalidType(t *testing.T) {
	config := EventAPI{
		Storage: storage{
			Type: "invalid",
		},
	}
	_, err := config.ConfigureStorage()
	require.Error(t, err)
}

func TestEventAPI_ConfigureStorage_defaultType(t *testing.T) {
	config := EventAPI{}
	sto, err := config.ConfigureStorage()
	require.NoError(t, err)
	require.IsType(t, &memory.Memory{}, sto)
}
