package data

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/anypb"

	"cacheserver/pkg/cache"
)

// namespacedCache implements the namespaced.Cache interface using chain cache.
type namespacedCache struct {
	chain *cache.ChainCache[any]
	log   *log.Helper
}

// Set stores a value in the cache.
func (c *namespacedCache) Set(ctx context.Context, key string, value *anypb.Any) error {
	// Serialize protobuf to JSON for storage
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.chain.Set(ctx, key, string(data))
}

// SetWithTTL stores a value in the cache with a TTL.
func (c *namespacedCache) SetWithTTL(ctx context.Context, key string, value *anypb.Any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.chain.SetWithTTL(ctx, key, string(data), ttl)
}

// Get retrieves a value from the cache.
func (c *namespacedCache) Get(ctx context.Context, key string) (*anypb.Any, error) {
	result, err := c.chain.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	value := &anypb.Any{}
	if str, ok := result.(string); ok {
		if err := json.Unmarshal([]byte(str), value); err != nil {
			return nil, err
		}
	}
	return value, nil
}

// GetWithTTL retrieves a value and its TTL from the cache.
func (c *namespacedCache) GetWithTTL(ctx context.Context, key string) (*anypb.Any, time.Duration, error) {
	result, ttl, err := c.chain.GetWithTTL(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	value := &anypb.Any{}
	if str, ok := result.(string); ok {
		if err := json.Unmarshal([]byte(str), value); err != nil {
			return nil, 0, err
		}
	}
	return value, ttl, nil
}

// Del removes a value from the cache.
func (c *namespacedCache) Del(ctx context.Context, key string) error {
	return c.chain.Del(ctx, key)
}
