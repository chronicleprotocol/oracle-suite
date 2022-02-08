package redis

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Redis struct {
	mu sync.Mutex

	client *redis.Client
	ttl    time.Duration
}

type Config struct {
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
func New(cfg Config) *Redis {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Address,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		ttl: cfg.TTL,
	}
}

// Add implements the store.Storage interface.
func (r *Redis) Add(ctx context.Context, author []byte, evt *messages.Event) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx := r.client.TxPipeline()
	defer func() {
		_, txErr := tx.Exec(ctx)
		if err == nil {
			err = txErr
		}
	}()
	key := redisMessageKey(evt.Type, evt.Index, author, evt.ID)
	val, err := evt.MarshallBinary()
	if err != nil {
		return err
	}
	getRes := r.client.Get(ctx, key)
	if getRes.Err() == redis.Nil {
		tx.Set(ctx, key, val, 0)
		tx.ExpireAt(ctx, key, evt.EventDate.Add(r.ttl))
	} else {
		currEvt := &messages.Event{}
		err = currEvt.UnmarshallBinary([]byte(getRes.Val()))
		if err != nil {
			return err
		}
		if currEvt.MessageDate.Before(evt.MessageDate) {
			tx.Set(ctx, key, val, 0)
			tx.ExpireAt(ctx, key, evt.EventDate.Add(r.ttl))
		}
	}
	return err
}

// Get implements the store.Storage interface.
func (r *Redis) Get(ctx context.Context, typ string, idx []byte) (evts []*messages.Event, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var keys []string
	var cursor uint64
	key := redisWildcardKey(typ, idx)
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

func redisMessageKey(typ string, index []byte, author []byte, id []byte) string {
	return fmt.Sprintf("%x:%x", hashIndex(typ, index), hashUnique(author, id))
}

func redisWildcardKey(typ string, index []byte) string {
	return fmt.Sprintf("%x:*", hashIndex(typ, index))
}

func hashUnique(author []byte, id []byte) [sha256.Size]byte {
	return sha256.Sum256(append(author, id...))
}

func hashIndex(typ string, index []byte) [sha256.Size]byte {
	return sha256.Sum256(append([]byte(typ), index...))
}
