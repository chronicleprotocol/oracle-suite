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
var eventAPIFactory = func(cfg api.Config) (*api.EventAPI, error) {
	return api.New(cfg)
}

type EventAPI struct {
	ListenAddr string  `json:"listenAddr"`
	Storage    storage `json:"storage"`
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
	TTL         int    `json:"ttl"`
	Address     string `json:"address"`
	Password    string `json:"password"`
	DB          int    `json:"db"`
	MemoryLimit int64  `json:"memoryLimit"`
}

type Dependencies struct {
	EventStore *store.EventStore
	Transport  transport.Transport
	Logger     log.Logger
}

type DatastoreDependencies struct {
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *EventAPI) Configure(d Dependencies) (*api.EventAPI, error) {
	return eventAPIFactory(api.Config{
		EventStore: d.EventStore,
		Address:    c.ListenAddr,
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
		r, err := redis.New(redis.Config{
			TTL:         time.Duration(ttl) * time.Second,
			Address:     c.Storage.Redis.Address,
			Password:    c.Storage.Redis.Password,
			DB:          c.Storage.Redis.DB,
			MemoryLimit: c.Storage.Redis.MemoryLimit,
		})
		if err != nil {
			return nil, fmt.Errorf(`eventapi config: unable to connect to the Redis server: %w`, err)
		}
		return r, nil
	default:
		return nil, fmt.Errorf(`eventapi config: storage type must be "memory", "redis" or empty to use default one`)
	}
}
