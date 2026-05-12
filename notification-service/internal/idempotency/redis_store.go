package idempotency

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore() *RedisStore {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisStore{client: client}
}

func (s *RedisStore) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	key := fmt.Sprintf("notification:processed:%s", eventID)

	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (s *RedisStore) MarkProcessed(ctx context.Context, eventID string) error {
	key := fmt.Sprintf("notification:processed:%s", eventID)

	return s.client.Set(ctx, key, "processed", 24*time.Hour).Err()
}
