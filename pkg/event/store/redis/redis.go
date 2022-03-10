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

const memUsageTimeQuantum = 3600

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
func (r *Redis) Add(ctx context.Context, author []byte, evt *messages.Event) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	val, err := evt.MarshallBinary()
	if err != nil {
		return err
	}
	availMem, err := r.getAvailMem(ctx, author)
	if err != nil {
		return err
	}
	if r.memLimit > 0 && availMem < int64(len(val)) {
		return ErrMemoryLimitExceed
	}
	key := redisMessageKey(evt.Type, evt.Index, author, evt.ID)
	tx := r.client.TxPipeline()
	defer func() {
		_, txErr := tx.Exec(ctx)
		if err == nil {
			err = txErr
		}
	}()
	getRes := r.client.Get(ctx, key)
	switch getRes.Err() {
	case nil:
		// If an event with the same ID exists, replace it if it is older.
		currEvt := &messages.Event{}
		err = currEvt.UnmarshallBinary([]byte(getRes.Val()))
		if err != nil {
			return err
		}
		if currEvt.MessageDate.Before(evt.MessageDate) {
			err = r.incrMemUsage(ctx, author, len(val)-len(getRes.Val()), evt.EventDate)
			if err != nil {
				return err
			}
			tx.Set(ctx, key, val, 0)
			tx.ExpireAt(ctx, key, evt.EventDate.Add(r.ttl))
		}
	case redis.Nil:
		// If an event with that ID does not exist, add it.
		err = r.incrMemUsage(ctx, author, len(val), evt.EventDate)
		if err != nil {
			return err
		}
		tx.Set(ctx, key, val, 0)
		tx.ExpireAt(ctx, key, evt.EventDate.Add(r.ttl))
	default:
		return getRes.Err()
	}
	return nil
}

// Get implements the store.Storage interface.
func (r *Redis) Get(ctx context.Context, typ string, idx []byte) (evts []*messages.Event, err error) {
	var keys []string
	var cursor uint64
	key := redisWildcardMessageKey(typ, idx)
	for {
		scanRes := r.client.Scan(ctx, cursor, key, 0)
		keys, cursor, err = scanRes.Result()
		if err != nil {
			return nil, err
		}
		if len(keys) == 0 {
			break
		}
		getRes := r.client.MGet(ctx, keys...)
		if getRes.Err() != nil {
			return nil, getRes.Err()
		}
		for _, val := range getRes.Val() {
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
		if cursor == 0 {
			break
		}
	}
	return evts, nil
}

func (r *Redis) incrMemUsage(ctx context.Context, author []byte, eventSize int, eventDate time.Time) error {
	k := redisMemUsageKey(author, eventDate)
	t := time.Unix(eventDate.Unix()/memUsageTimeQuantum*memUsageTimeQuantum, 0)
	p := r.client.Pipeline()
	incrRes := p.IncrBy(ctx, k, int64(eventSize))
	if incrRes.Err() != nil {
		return incrRes.Err()
	}
	expireRes := p.ExpireAt(ctx, k, t.Add(r.ttl+(time.Second*memUsageTimeQuantum)))
	if expireRes.Err() != nil {
		return expireRes.Err()
	}
	_, err := p.Exec(ctx)
	return err
}

func (r *Redis) getAvailMem(ctx context.Context, author []byte) (int64, error) {
	var err error
	var size int
	var keys []string
	var cursor uint64
	key := redisWildcardMemUsageKey(author)
	for {
		scanRes := r.client.Scan(ctx, cursor, key, 0)
		keys, cursor, err = scanRes.Result()
		if err != nil {
			return 0, err
		}
		if len(keys) == 0 {
			break
		}
		getRes := r.client.MGet(ctx, keys...)
		if getRes.Err() != nil {
			return 0, getRes.Err()
		}
		for _, val := range getRes.Val() {
			s, ok := val.(string)
			if !ok {
				continue
			}
			i, err := strconv.Atoi(s)
			if err != nil {
				continue
			}
			size += i
		}
		if cursor == 0 {
			break
		}
	}
	return r.memLimit - int64(size), nil
}

func redisMemUsageKey(author []byte, eventDate time.Time) string {
	return fmt.Sprintf("%x:%x", author, eventDate.Unix()/memUsageTimeQuantum)
}

func redisWildcardMemUsageKey(author []byte) string {
	return fmt.Sprintf("%x:*", author)
}

func redisMessageKey(typ string, index []byte, author []byte, id []byte) string {
	return fmt.Sprintf("%x:%x", hashIndex(typ, index), hashUnique(author, id))
}

func redisWildcardMessageKey(typ string, index []byte) string {
	return fmt.Sprintf("%x:*", hashIndex(typ, index))
}

func hashUnique(author []byte, id []byte) [sha256.Size]byte {
	return sha256.Sum256(append(author, id...))
}

func hashIndex(typ string, index []byte) [sha256.Size]byte {
	return sha256.Sum256(append([]byte(typ), index...))
}
