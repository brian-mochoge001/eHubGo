package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	IsAvailable() bool
}

type RedisStore struct {
	client *redis.Client
	active bool
}

func NewRedisStore(url string) *RedisStore {
	if url == "" {
		url = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Printf("[Cache] Failed to parse Redis URL, cache disabled: %v", err)
		return &RedisStore{active: false}
	}

	client := redis.NewClient(opt)
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[Cache] Failed to connect to Redis, cache disabled: %v", err)
		return &RedisStore{active: false}
	}

	log.Printf("[Cache] Redis connected successfully")
	return &RedisStore{
		client: client,
		active: true,
	}
}

func (s *RedisStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !s.active {
		return nil
	}
	return s.client.Set(ctx, key, value, expiration).Err()
}

func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	if !s.active {
		return "", redis.Nil // simulate cache miss
	}
	return s.client.Get(ctx, key).Result()
}

func (s *RedisStore) Delete(ctx context.Context, key string) error {
	if !s.active {
		return nil
	}
	return s.client.Del(ctx, key).Err()
}

func (s *RedisStore) IsAvailable() bool {
	return s.active
}
