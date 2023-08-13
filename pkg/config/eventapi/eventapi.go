//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"

	"github.com/chronicleprotocol/oracle-suite/pkg/event/api"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store/redis"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

const week uint32 = 3600 * 24 * 7

type Dependencies struct {
	EventStore *store.EventStore
	Transport  transport.Service
	Logger     log.Logger
}

type DatastoreDependencies struct {
	Transport transport.Service
	Feeds     []types.Address
	Logger    log.Logger
}

type Config struct {
	// ListenAddr is the address on which the event API will listen.
	ListenAddr string `hcl:"listen_addr"`

	// Memory is the configuration for the in-memory storage. Cannot be
	// used together with storage_redis configuration.
	Memory *storageMemory `hcl:"storage_memory,block,optional"`

	// Redis is the configuration for the Redis storage. Cannot be used
	// together with storage_memory configuration.
	Redis *storageRedis `hcl:"storage_redis,block,optional"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured services:
	eventAPI *api.EventAPI
	storage  store.Storage
}

type storageMemory struct {
	// TTL is the time to live for the events in the in-memory storage in
	// seconds. Defaults to 604800 (one week).
	TTL uint32 `hcl:"ttl,optional"`
}

type storageRedis struct {
	// TTL is the time to live for the events in the Redis storage in seconds.
	// Defaults to 604800 (one week).
	TTL uint32 `hcl:"ttl,optional"`

	// Address is the redis server address provided as the combination of IP
	// address or host and port number, e.g. `0.0.0.0:8080`.
	Address string `hcl:"addr"`

	// Username is the username for the ACL.
	Username string `hcl:"user,optional"`

	// Password is the password for the ACL.
	Password string `hcl:"pass,optional"`

	// DB is the database number. Ignored in cluster mode.
	DB int `hcl:"db,optional"`

	// MemoryLimit is a limit of data per feed in bytes. If 0 or not specified,
	// no limit is applied.
	MemoryLimit int64 `hcl:"memory_limit,optional"`

	// TLS enables TLS for the connection to the Redis server.
	TLS bool `hcl:"tls,optional"`

	// TLSServerName is the server name used to verify the hostname on the
	// returned certificates from the server. Ignored if empty
	TLSServerName string `hcl:"tls_server_name,optional"`

	// TLSCertFile is the path to PEM encoded certificate file.
	TLSCertFile string `hcl:"tls_cert_file,optional"`

	// TLSKeyFile is the path to PEM encoded private key file.
	TLSKeyFile string `hcl:"tls_key_file,optional"`

	// TLSRootCAFile is the path to PEM encoded root certificate file.
	TLSRootCAFile string `hcl:"tls_root_ca_file,optional"`

	// TLSInsecureSkipVerify disables TLS certificate verification.
	TLSInsecureSkipVerify bool `hcl:"tls_insecure_skip_verify,optional"`

	// Cluster enables cluster mode.
	Cluster bool `hcl:"cluster,optional"`

	// ClusterAddrs is a list of cluster node addresses provided as the
	// combination of IP address or host and port number, e.g. `0.0.0.0:8080`.
	ClusterAddrs []string `hcl:"cluster_addrs,optional"`

	// HCL fields:
	Range hcl.Range `hcl:",range"`
}

func (c *Config) EventAPI(d Dependencies) (*api.EventAPI, error) {
	if c.eventAPI != nil {
		return c.eventAPI, nil
	}
	eventAPI, err := api.New(api.Config{
		EventStore: d.EventStore,
		Address:    c.ListenAddr,
		Logger:     d.Logger,
	})
	if err != nil {
		return nil, err
	}
	c.eventAPI = eventAPI
	return eventAPI, nil
}

func (c *Config) Storage() (store.Storage, error) {
	if c.storage != nil {
		return c.storage, nil
	}
	if c.Memory != nil && c.Redis != nil {
		return nil, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   `"storage_memory" and "storage_redis" storage types are mutually exclusive`,
			Subject:  c.Range.Ptr(),
		}}
	}
	switch {
	case c.Memory != nil:
		ttl := week
		if c.Memory.TTL > 0 {
			ttl = c.Memory.TTL
		}
		c.storage = store.NewMemoryStorage(time.Second * time.Duration(ttl))
		return c.storage, nil
	case c.Redis != nil:
		ttl := week
		if c.Redis.TTL > 0 {
			ttl = c.Redis.TTL
		}
		r, err := redis.New(redis.Config{
			TTL:                   time.Second * time.Duration(ttl),
			Address:               c.Redis.Address,
			Username:              c.Redis.Username,
			Password:              c.Redis.Password,
			DB:                    c.Redis.DB,
			MemoryLimit:           c.Redis.MemoryLimit,
			TLS:                   c.Redis.TLS,
			TLSServerName:         c.Redis.TLSServerName,
			TLSCertFile:           c.Redis.TLSCertFile,
			TLSKeyFile:            c.Redis.TLSKeyFile,
			TLSRootCAFile:         c.Redis.TLSRootCAFile,
			TLSInsecureSkipVerify: c.Redis.TLSInsecureSkipVerify,
			Cluster:               c.Redis.Cluster,
			ClusterAddrs:          c.Redis.ClusterAddrs,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf(`Unable to create a Redis storage: %s`, err),
				Subject:  c.Redis.Range.Ptr(),
			}
		}
		if err := r.Ping(context.Background()); err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf(`Unable to ping the Redis storage: %s`, err),
				Subject:  c.Redis.Range.Ptr(),
			}
		}
		c.storage = r
		return c.storage, nil
	default:
		return nil, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   `One of "storage_memory" or "storage_redis" storage types must be specified`,
			Subject:  c.Range.Ptr(),
		}}
	}
}
