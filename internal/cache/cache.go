package cache

import (
	"context"
	"github.com/go-redis/redis/v9"
	"time"
)

type Cache interface {
	Set(key, value string, expired time.Duration) error
	Get(key string) (string, error)
}

type cache struct {
	rdb *redis.Client
}

func NewCatch(rdb *redis.Client) Cache {
	return cache{
		rdb: rdb,
	}
}

func (c cache) Set(key, value string, expired time.Duration) error {
	return c.rdb.Set(context.Background(), key, value, expired).Err()
}

func (c cache) Get(key string) (string, error) {
	return c.rdb.Get(context.Background(), key).Result()
}
