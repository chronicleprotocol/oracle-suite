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
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

var ErrMemoryLimitExceed = errors.New("memory limit exceeded")

const retryAttempts = 3               // The maximum number of attempts to call EthClient in case of an error.
const retryInterval = 1 * time.Second // The delay between retry attempts.
const memUsageTimeQuantum = 3600      // The length of the time window for which memory usage information is stored.

// Redis provides storage mechanism for store.EventStore.
// It uses a Redis database to store events.
type Redis struct {
	mu sync.Mutex

	client   *redis.Client
	ttl      time.Duration
	memLimit int64
}

// Config contains configuration parameters for Redis.
type Config struct {
	// MemoryLimit specifies a maximum memory limit for a single Oracle.
	MemoryLimit int64
	// TTL specifies how long messages should be kept in storage.
	TTL time.Duration
	// Address specifies Redis server address as "host:port".
	Address string
	// Password specifies Redis server password.
	Password string
	// DB is the Redis database number.
	DB int
}

// New returns a new instance of Redis.
func New(cfg Config) (*Redis, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	// go-redis default timeout is 5 seconds, so using background context should be ok
	res := cli.Ping(context.Background())
	if res.Err() != nil {
		return nil, res.Err()
	}
	return &Redis{
		client:   cli,
		ttl:      cfg.TTL,
		memLimit: cfg.MemoryLimit,
	}, nil
}

// Add implements the store.Storage interface.
func (r *Redis) Add(ctx context.Context, author []byte, evt *messages.Event) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var added bool
	var err error
	err = retry(func() error {
		added, err = r.add(ctx, author, evt)
		return err
	})
	return added, err
}

// Get implements the store.Storage interface.
func (r *Redis) Get(ctx context.Context, typ string, idx []byte) ([]*messages.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var err error
	var evts []*messages.Event
	err = retry(func() error {
		evts, err = r.get(ctx, typ, idx)
		return err
	})
	return evts, err
}

func (r *Redis) add(ctx context.Context, author []byte, evt *messages.Event) (bool, error) {
	key := evtKey(evt.Type, evt.Index, author, evt.ID)
	val, err := evt.MarshallBinary()
	if err != nil {
		return false, err
	}
	mem, err := r.getAvailMem(ctx, r.client, author)
	if err != nil {
		return false, err
	}
	if r.memLimit > 0 && int64(len(val)) > mem {
		return false, ErrMemoryLimitExceed
	}
	var added bool
	err = r.client.Watch(ctx, func(tx *redis.Tx) error {
		prevVal, err := r.client.Get(ctx, key).Result()
		switch err {
		case nil:
			// If an event with the same ID exists, replace it if it is older.
			prevEvt := &messages.Event{}
			err = prevEvt.UnmarshallBinary([]byte(prevVal))
			if err != nil {
				return err
			}
			if prevEvt.MessageDate.Before(evt.MessageDate) {
				err = r.incrMemUsage(ctx, tx, author, len(val)-len(prevVal), evt.EventDate)
				if err != nil {
					return err
				}
				tx.Set(ctx, key, val, 0)
				tx.ExpireAt(ctx, key, evt.EventDate.Add(r.ttl))
			}
		case redis.Nil:
			// If an event with that ID does not exist, add it.
			err = r.incrMemUsage(ctx, tx, author, len(val), evt.EventDate)
			if err != nil {
				return err
			}
			tx.Set(ctx, key, val, 0)
			tx.ExpireAt(ctx, key, evt.EventDate.Add(r.ttl))
			added = true
		default:
			return err
		}
		return nil
	}, key)
	return added, err
}

func (r *Redis) get(ctx context.Context, typ string, idx []byte) ([]*messages.Event, error) {
	var evts []*messages.Event
	err := r.scan(ctx, wildcardEvtKey(typ, idx), r.client, func(keys []string) error {
		vals, err := r.client.MGet(ctx, keys...).Result()
		if err != nil {
			return err
		}
		for _, val := range vals {
			b, ok := val.(string)
			if !ok {
				continue
			}
			evt := &messages.Event{}
			err = evt.UnmarshallBinary([]byte(b))
			if err != nil {
				continue
			}
			evts = append(evts, evt)
		}
		return nil
	})
	return evts, err
}

func (r *Redis) incrMemUsage(ctx context.Context, c redis.Cmdable, author []byte, mem int, evtDate time.Time) error {
	if r.memLimit == 0 {
		return nil
	}
	var err error
	key := memUsageKey(author, evtDate)
	err = c.IncrBy(ctx, key, int64(mem)).Err()
	if err != nil {
		return err
	}
	q := int64(memUsageTimeQuantum)
	t := (evtDate.Unix()/q)*q + q
	err = c.ExpireAt(ctx, key, time.Unix(t, 0).Add(r.ttl)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Redis) getAvailMem(ctx context.Context, c redis.Cmdable, author []byte) (int64, error) {
	if r.memLimit == 0 {
		return 0, nil
	}
	var size int64
	err := r.scan(ctx, wildcardMemUsageKey(author), c, func(keys []string) error {
		vals, err := c.MGet(ctx, keys...).Result()
		if err != nil {
			return err
		}
		for _, val := range vals {
			s, ok := val.(string)
			if !ok {
				continue
			}
			i, err := strconv.Atoi(s)
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

func (r *Redis) scan(ctx context.Context, pattern string, c redis.Cmdable, fn func(keys []string) error) error {
	var err error
	var keys []string
	var cursor uint64
	for {
		keys, cursor, err = c.Scan(ctx, cursor, pattern, 0).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			err = fn(keys)
			if err != nil {
				return err
			}
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

func evtKey(typ string, index []byte, author []byte, id []byte) string {
	return fmt.Sprintf("evt:%x:%x", hashIndex(typ, index), hashUnique(author, id))
}

func wildcardEvtKey(typ string, index []byte) string {
	return fmt.Sprintf("evt:%x:*", hashIndex(typ, index))
}

func memUsageKey(author []byte, eventDate time.Time) string {
	return fmt.Sprintf("mem:%x:%x", author, eventDate.Unix()/memUsageTimeQuantum)
}

func wildcardMemUsageKey(author []byte) string {
	return fmt.Sprintf("mem:%x:*", author)
}

func hashUnique(author []byte, id []byte) [sha256.Size]byte {
	return sha256.Sum256(append(author, id...))
}

func hashIndex(typ string, index []byte) [sha256.Size]byte {
	return sha256.Sum256(append([]byte(typ), index...))
}

// retry runs the f function until it returns nil. Maximum number of retries
// and delay between them are defined in the retryAttempts and retryInterval
// constants.
//
// If error is ErrMemoryLimitExceed, it will be returned immediately.
func retry(f func() error) (err error) {
	for i := 0; i < retryAttempts; i++ {
		if i > 0 {
			time.Sleep(retryInterval)
		}
		err = f()
		if errors.Is(err, ErrMemoryLimitExceed) {
			return err
		}
		if err == nil {
			return nil
		}
	}
	return err
}
