package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"cacheserver/pkg/cache/store"
)

// RedisStore is a store for Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedis creates a new store to Redis instance(s).
func NewRedis(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}

// Get returns data stored from a given key.
func (s *RedisStore) Get(ctx context.Context, key any) (any, error) {
	obj, err := s.client.Get(ctx, key.(string)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, store.ErrKeyNotFound
	}
	return obj, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL.
func (s *RedisStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	obj, err := s.client.Get(ctx, key.(string)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, 0, store.ErrKeyNotFound
	}
	if err != nil {
		return nil, 0, err
	}

	ttl, err := s.client.TTL(ctx, key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	return obj, ttl, nil
}

// Set defines data in Redis for given key identifier.
func (s *RedisStore) Set(ctx context.Context, key any, value any) error {
	return s.client.Set(ctx, key.(string), value, 0).Err()
}

// SetWithTTL defines data in Redis for given key identifier with TTL.
func (s *RedisStore) SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) error {
	return s.client.Set(ctx, key.(string), value, ttl).Err()
}

// Del removes data from Redis for given key identifier.
func (s *RedisStore) Del(ctx context.Context, key any) error {
	_, err := s.client.Del(ctx, key.(string)).Result()
	return err
}

// Clear resets all data in the store.
func (s *RedisStore) Clear(ctx context.Context) error {
	return s.client.FlushAll(ctx).Err()
}

// Wait waits for all operations to complete.
func (s *RedisStore) Wait(_ context.Context) {}
