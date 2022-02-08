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
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/api"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store/memory"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store/redis"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

const week = 3600 * 24 * 7

//nolint
var eventAPIFactory = func(ctx context.Context, cfg api.Config) (*api.EventAPI, error) {
	return api.New(ctx, cfg)
}

type EventAPI struct {
	Address string  `json:"address"`
	Storage storage `json:"storage"`
}

type storage struct {
	Type   string        `json:"type"`
	Memory storageMemory `json:"memory"`
	Redis  storageRedis  `json:"redis"`
}

type storageMemory struct {
	TTL int `json:"ttl"`
}

type storageRedis struct {
	TTL      int    `json:"ttl"`
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type Dependencies struct {
	Context    context.Context
	EventStore *store.EventStore
	Transport  transport.Transport
	Logger     log.Logger
}

type DatastoreDependencies struct {
	Context   context.Context
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *EventAPI) Configure(d Dependencies) (*api.EventAPI, error) {
	return eventAPIFactory(d.Context, api.Config{
		EventStore: d.EventStore,
		Address:    c.Address,
		Logger:     d.Logger,
	})
}

func (c *EventAPI) ConfigureStorage() (store.Storage, error) {
	switch c.Storage.Type {
	case "memory", "":
		ttl := week
		if c.Storage.Memory.TTL > 0 {
			ttl = c.Storage.Memory.TTL
		}
		return memory.New(time.Second * time.Duration(ttl)), nil
	case "redis":
		ttl := week
		if c.Storage.Redis.TTL > 0 {
			ttl = c.Storage.Redis.TTL
		}
		return redis.New(redis.Config{
			TTL:      time.Duration(ttl) * time.Second,
			Address:  c.Storage.Redis.Address,
			Password: c.Storage.Redis.Password,
			DB:       c.Storage.Redis.DB,
		}), nil
	default:
		return nil, fmt.Errorf(`eventapi config: storage type must be "memory", "redis" or empty to use default one`)
	}
}
