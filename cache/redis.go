package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)
type Store interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetJSON(ctx context.Context, key string, dest interface{}) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) <-chan *redis.Message
	Delete(ctx context.Context, key string) error
	IsAvailable() bool
}

func (s *RedisStore) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !s.active {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, expiration).Err()
}

func (s *RedisStore) GetJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	if !s.active {
		return false, nil
	}
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(data, dest)
	if err != nil {
		return false, err
	}
	return true, nil
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

func (s *RedisStore) Incr(ctx context.Context, key string) (int64, error) {
	if !s.active {
		return 0, nil
	}
	return s.client.Incr(ctx, key).Result()
}

func (s *RedisStore) Publish(ctx context.Context, channel string, message interface{}) error {
	if !s.active {
		return nil
	}
	return s.client.Publish(ctx, channel, message).Err()
}

func (s *RedisStore) Subscribe(ctx context.Context, channel string) <-chan *redis.Message {
	if !s.active {
		return nil
	}
	pubsub := s.client.Subscribe(ctx, channel)
	return pubsub.Channel()
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
