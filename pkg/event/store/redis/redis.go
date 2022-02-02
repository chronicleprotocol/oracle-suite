package redis

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Redis struct {
	ctx    context.Context
	client *redis.Client
	ttl    time.Duration
}

type Config struct {
	TTL      time.Duration
	Addr     string
	Password string
	DB       int
}

func NewRedis(ctx context.Context, cfg Config) *Redis {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		ttl: cfg.TTL,
	}
}

func (r Redis) Add(author []byte, msg *messages.Event) error {
	val, err := msg.MarshallBinary()
	if err != nil {
		return err
	}
	s := r.client.Set(
		r.ctx,
		fmt.Sprintf("%s:%s", hashIndex(msg.Type, msg.Index), hashUnique(author, msg.ID)),
		val,
		r.ttl,
	)
	return s.Err()
}

func (r Redis) Get(typ string, idx []byte) ([]*messages.Event, error) {
	var keys []string
	var cursor uint64
	var err error
	var msgs []*messages.Event
	match := fmt.Sprintf("%s:*", hashIndex(typ, idx))
	for {
		scanRes := r.client.Scan(r.ctx, cursor, match, 0)
		keys, cursor, err = scanRes.Result()
		if err != nil {
			return nil, err
		}
		getRes := r.client.MGet(r.ctx, keys...)
		if getRes.Err() != nil {
			return nil, getRes.Err()
		}
		for _, val := range getRes.Val() {
			b, ok := val.([]byte)
			if !ok {
				continue
			}
			msg := &messages.Event{}
			err = msg.UnmarshallBinary(b)
			if err != nil {
				continue
			}
			msgs = append(msgs, msg)
		}
		if cursor == 0 {
			break
		}
	}
	return msgs, nil
}

func hashUnique(author []byte, id []byte) [sha256.Size]byte {
	return sha256.Sum256(append(author, id...))
}

func hashIndex(typ string, index []byte) [sha256.Size]byte {
	return sha256.Sum256(append([]byte(typ), index...))
}
