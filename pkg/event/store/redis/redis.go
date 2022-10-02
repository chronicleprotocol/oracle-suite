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

package redis

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

var ErrMemoryLimitExceed = errors.New("memory limit exceeded")

const txRetryAttempts = 3        // Maximum number of attempts to retry a transaction.
const memUsageTimeQuantum = 3600 // The length of the time window for which memory usage information is stored.

// Storage provides storage mechanism for store.EventStore.
// It uses a Redis database to store events.
type Storage struct {
	mu sync.Mutex

	client   redis.UniversalClient
	ttl      time.Duration
	memLimit int64
}

// Config is the configuration for the Storage.
type Config struct {
	// MemoryLimit specifies a maximum memory limit for a single Oracle.
	MemoryLimit int64
	// TTL specifies how long messages should be kept in storage.
	TTL time.Duration
	// Address specifies Redis server address as "host:port".
	Address string
	// Username specifies Redis username for the ACL.
	Username string
	// Password specifies Redis server password.
	Password string
	// DB is the Redis database number.
	DB int
	// TLS specifies whether to use TLS for Redis connection.
	TLS bool
	// TLSServerName specifies the server name used to verify
	// the hostname on the returned certificates from the server.
	TLSServerName string
	// TLSCertFile specifies the path to the client certificate file.
	TLSCertFile string
	// TLSKeyFile specifies the path to the client key file.
	TLSKeyFile string
	// TLSRootCAFile specifies the path to the CA certificate file.
	TLSRootCAFile string
	// TLSInsecureSkipVerify specifies whether to skip server certificate verification.
	TLSInsecureSkipVerify bool
	// Cluster specifies whether the Redis server is a cluster.
	Cluster bool
	// ClusterAddrs specifies the Redis cluster addresses as "host:port".
	ClusterAddrs []string
}

// New returns a new instance of Redis.
func New(cfg Config) (*Storage, error) {
	var client redis.UniversalClient
	var tlsConfig *tls.Config

	if cfg.TLS {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		if cfg.TLSServerName != "" {
			tlsConfig.ServerName = cfg.TLSServerName
		}
		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		if cfg.TLSRootCAFile != "" {
			caCert, err := os.ReadFile(cfg.TLSRootCAFile)
			if err != nil {
				return nil, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
		tlsConfig.InsecureSkipVerify = cfg.TLSInsecureSkipVerify
	}

	if cfg.Cluster {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:         cfg.ClusterAddrs,
			Username:      cfg.Username,
			Password:      cfg.Password,
			TLSConfig:     tlsConfig,
			RouteRandomly: true,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:      cfg.Address,
			Username:  cfg.Username,
			Password:  cfg.Password,
			DB:        cfg.DB,
			TLSConfig: tlsConfig,
		})
	}

	return &Storage{
		client:   client,
		ttl:      cfg.TTL,
		memLimit: cfg.MemoryLimit,
	}, nil
}

// Ping checks if the Redis server is available.
func (r *Storage) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return err
	}
	if rds, ok := r.client.(*redis.ClusterClient); ok {
		if err := rds.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
			return shard.Ping(ctx).Err()
		}); err != nil {
			return err
		}
	}
	return nil
}

// Add implements the store.Storage interface.
func (r *Storage) Add(ctx context.Context, author []byte, evt *messages.Event) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := evtKey(evt.Type, evt.Index, author, evt.ID)
	val, err := evt.MarshallBinary()
	if err != nil {
		return false, err
	}
	// Check if the memory limit is exceeded.
	mem, err := r.getAvailMem(ctx, author)
	if err != nil {
		return false, err
	}
	if r.memLimit > 0 && int64(len(val)) > mem {
		return false, ErrMemoryLimitExceed
	}
	var isNew bool
	// If the key already exists, we need to decide whether to overwrite it.
	// We do this by comparing the timestamps of the new and existing events.
	// If the new event is older than the existing one, we do not overwrite it.
	//
	// To be able to do this, we must first get an existing event, but this may
	// cause a race condition - the key can be updated between reading it and
	// updating it. To avoid this, we use a Redis transaction. The following
	// transaction watches to see if the value of the key has been changed
	// during the transaction, if so, the transaction is canceled. In this
	// case, the redisWatch function will try to retry the transaction several times.
	//
	// The transaction is also passed to the incrMemUsage method, so that
	// memory usage is updated only if the transaction is successful.
	err = r.redisWatch(ctx, func(tx *redis.Tx) error {
		prevVal, err := r.client.Get(ctx, key).Result()
		switch err {
		case nil: // No error, the key exists.
			prevEvt := &messages.Event{}
			if err := prevEvt.UnmarshallBinary([]byte(prevVal)); err != nil {
				return err
			}
			if prevEvt.MessageDate.Before(evt.MessageDate) {
				if err := r.incrMemUsage(ctx, tx, author, len(val)-len(prevVal), evt.EventDate); err != nil {
					return err
				}
				tx.Set(ctx, key, val, 0)
				tx.ExpireAt(ctx, key, evt.EventDate.Add(r.ttl))
			}
		case redis.Nil: // The key does not exist.
			if err := r.incrMemUsage(ctx, tx, author, len(val), evt.EventDate); err != nil {
				return err
			}
			tx.Set(ctx, key, val, 0)
			tx.ExpireAt(ctx, key, evt.EventDate.Add(r.ttl))
			isNew = true
		default:
			return err
		}
		return nil
	}, key)
	return isNew, err
}

// Get implements the store.Storage interface.
func (r *Storage) Get(ctx context.Context, typ string, idx []byte) ([]*messages.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var evts []*messages.Event
	err := r.redisScan(ctx, wildcardEvtKey(typ, idx), func(keys []string) error {
		vals, err := r.redisMGet(ctx, keys...)
		if err != nil {
			return err
		}
		for _, val := range vals {
			evt := &messages.Event{}
			if err := evt.UnmarshallBinary([]byte(val)); err != nil {
				continue
			}
			evts = append(evts, evt)
		}
		return nil
	})
	return evts, err
}

// getAvailMem returns the available memory for the given author.
//
// Finds all the memory usage keys for given author and sums them up. The exact
// mechanism of how the memory usage is stored is described in incrMemUsage.
func (r *Storage) getAvailMem(ctx context.Context, author []byte) (int64, error) {
	if r.memLimit == 0 {
		return 0, nil
	}
	var size int64
	err := r.redisScan(ctx, wildcardMemUsageKey(author), func(keys []string) error {
		vals, err := r.redisMGet(ctx, keys...)
		if err != nil {
			return err
		}
		for _, val := range vals {
			i, err := strconv.Atoi(val)
			if err != nil {
				continue
			}
			size += int64(i)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return r.memLimit - size, nil
}

// incrMemUsage increments the memory usage of the author by the given amount.
//
// Because memory usage needs to be updated after an event expire, the memory
// usage is stored in multiple keys with a TTL that is the same as the event
// TTL. This allows the memory usage to be updated after the event expires.
//
// To reduce number of keys, the keys with similar TTL are grouped together.
// It is done by rounding the TTL to the nearest multiple of memUsageTimeQuantum.
func (r *Storage) incrMemUsage(ctx context.Context, c redis.Cmdable, author []byte, mem int, evtDate time.Time) error {
	if r.memLimit == 0 || mem == 0 {
		return nil
	}
	key := memUsageKey(author, evtDate)
	if err := c.IncrBy(ctx, key, int64(mem)).Err(); err != nil {
		return err
	}
	q := int64(memUsageTimeQuantum)
	t := (evtDate.Unix()/q)*q + q
	if err := c.ExpireAt(ctx, key, time.Unix(t, 0).Add(r.ttl)).Err(); err != nil {
		return err
	}
	return nil
}

// redisWatch starts a transaction that watches the given keys and retries the
// transaction if it fails up txRetryAttempts times. The transaction fails
// if the watched keys are modified by another client.
//
// It is important, that all keys modified in the transaction must belong to
// the same slot! Otherwise, transaction silently fails.
func (r *Storage) redisWatch(ctx context.Context, fn func(tx *redis.Tx) error, keys ...string) error {
	for i := 0; i < txRetryAttempts; i++ {
		err := r.client.Watch(ctx, fn, keys...)
		if err == nil {
			return nil // Success.
		}
		if ctx.Err() != nil {
			return ctx.Err() // Context canceled.
		}
		if err == redis.TxFailedErr {
			continue // Optimistic lock lost. Retry.
		}
		return err // Return any other error.
	}
	return redis.TxFailedErr
}

// redisScan iterates over all keys matching the pattern and calls the callback.
// In cluster mode a scan is performed on each master node.
func (r *Storage) redisScan(ctx context.Context, pattern string, fn func(keys []string) error) error {
	if rds, ok := r.client.(*redis.ClusterClient); ok {
		if err := rds.ForEachMaster(ctx, func(ctx context.Context, c *redis.Client) error {
			return scanSingleNode(ctx, c, pattern, fn)
		}); err != nil {
			return err
		}
	} else {
		return scanSingleNode(ctx, r.client, pattern, fn)
	}
	return nil
}

// redisMGet returns the values for the given keys. The mget does not work in
// cluster mode when different keys belongs to different slots, for this
// reason, in cluster mode a pipeline is used.
func (r *Storage) redisMGet(ctx context.Context, keys ...string) ([]string, error) {
	// Cluster mode:
	if _, ok := r.client.(*redis.ClusterClient); ok {
		cmds, err := r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for _, key := range keys {
				if err := pipe.Get(ctx, key).Err(); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		var res []string
		for _, cmd := range cmds {
			if err := cmd.Err(); err != nil {
				return nil, err
			}
			res = append(res, cmd.(*redis.StringCmd).Val())
		}
		return res, nil
	}
	// Single node mode:
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	var res []string
	for _, val := range vals {
		s, ok := val.(string)
		if !ok {
			continue
		}
		res = append(res, s)
	}
	return res, nil
}

func scanSingleNode(ctx context.Context, c redis.Cmdable, pattern string, fn func(keys []string) error) error {
	var (
		err    error
		keys   []string
		cursor uint64
	)
	for {
		keys, cursor, err = c.Scan(ctx, cursor, pattern, 0).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err = fn(keys); err != nil {
				return err
			}
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Helpers for generating Redis keys:

func evtKey(typ string, idx []byte, author []byte, id []byte) string {
	return fmt.Sprintf("evt:%x:%x:{%x}", hashIdx(typ, idx), hashUnique(author, id), hashtag(author))
}

func wildcardEvtKey(typ string, idx []byte) string {
	return fmt.Sprintf("evt:%x:*", hashIdx(typ, idx))
}

func memUsageKey(author []byte, eventDate time.Time) string {
	return fmt.Sprintf("mem:%x:%x:{%x}", author, eventDate.Unix()/memUsageTimeQuantum, hashtag(author))
}

func wildcardMemUsageKey(author []byte) string {
	return fmt.Sprintf("mem:%x:*", author)
}

func hashUnique(author []byte, id []byte) [sha256.Size]byte {
	return sha256.Sum256(append(author, id...))
}

func hashIdx(typ string, idx []byte) [sha256.Size]byte {
	return sha256.Sum256(append([]byte(typ), idx...))
}

// hashtag calculates Redis hashtag used to ensure that keys from the same
// author have the same slot.
// https://redis.io/docs/reference/cluster-spec/
func hashtag(author []byte) []byte {
	h := make([]byte, 4)
	binary.BigEndian.PutUint32(h, crc32.ChecksumIEEE(author))
	return h
}
