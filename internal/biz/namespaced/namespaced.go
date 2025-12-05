package namespaced

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	v1 "cacheserver/api/cacheserver/v1"
)

// NamespacedBiz defines the interface for handling namespaced cache requests.
type NamespacedBiz interface {
	Set(ctx context.Context, key string, value *anypb.Any, ttl *durationpb.Duration) (*emptypb.Empty, error)
	Del(ctx context.Context, key string) (*emptypb.Empty, error)
	Get(ctx context.Context, key string) (*v1.GetResponse, error)
}

// Cache defines the interface for cache operations.
type Cache interface {
	Set(ctx context.Context, key string, value *anypb.Any) error
	SetWithTTL(ctx context.Context, key string, value *anypb.Any, ttl time.Duration) error
	Get(ctx context.Context, key string) (*anypb.Any, error)
	GetWithTTL(ctx context.Context, key string) (*anypb.Any, time.Duration, error)
	Del(ctx context.Context, key string) error
}

// NamespacedKey represents a key with a namespace.
type NamespacedKey struct {
	Namespace string
	Key       string
}

// CacheKey returns the cache key for the NamespacedKey.
func (k NamespacedKey) CacheKey() string {
	return fmt.Sprintf("namespace:%s:%s", k.Namespace, k.Key)
}

// namespacedBiz is the implementation of NamespacedBiz.
type namespacedBiz struct {
	cache     Cache
	namespace string
}

// Ensure that *namespacedBiz implements the NamespacedBiz.
var _ NamespacedBiz = (*namespacedBiz)(nil)

// New creates and returns a new instance of *namespacedBiz.
func New(cache Cache, namespace string) NamespacedBiz {
	return &namespacedBiz{cache: cache, namespace: namespace}
}

// Set stores a value with the given key and time to live (TTL) in the namespaced cache.
func (b *namespacedBiz) Set(ctx context.Context, key string, value *anypb.Any, ttl *durationpb.Duration) (*emptypb.Empty, error) {
	cacheKey := NamespacedKey{b.namespace, key}.CacheKey()
	var err error
	if ttl != nil {
		err = b.cache.SetWithTTL(ctx, cacheKey, value, ttl.AsDuration())
	} else {
		err = b.cache.Set(ctx, cacheKey, value)
	}
	return &emptypb.Empty{}, err
}

// Del deletes a value from the namespaced cache by its key.
func (b *namespacedBiz) Del(ctx context.Context, key string) (*emptypb.Empty, error) {
	cacheKey := NamespacedKey{b.namespace, key}.CacheKey()
	return &emptypb.Empty{}, b.cache.Del(ctx, cacheKey)
}

// Get retrieves a value from the namespaced cache by its key.
func (b *namespacedBiz) Get(ctx context.Context, key string) (*v1.GetResponse, error) {
	cacheKey := NamespacedKey{b.namespace, key}.CacheKey()
	value, ttl, err := b.cache.GetWithTTL(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	return &v1.GetResponse{Value: value, Expire: durationpb.New(ttl)}, nil
}
