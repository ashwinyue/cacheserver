package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "cacheserver/api/cacheserver/v1"
	"cacheserver/internal/biz"
)

// CacheServerService provides gRPC methods to handle cache operations.
type CacheServerService struct {
	v1.UnimplementedCacheServerServer

	biz biz.ICacheBiz
}

// Ensure that CacheServerService implements the v1.CacheServerServer interface.
var _ v1.CacheServerServer = (*CacheServerService)(nil)

// NewCacheServerService creates and returns a new instance of CacheServerService.
func NewCacheServerService(biz biz.ICacheBiz) *CacheServerService {
	return &CacheServerService{biz: biz}
}

// Set stores a key-value pair in the cache with an optional expiration time.
func (s *CacheServerService) Set(ctx context.Context, rq *v1.SetRequest) (*emptypb.Empty, error) {
	return s.biz.NamespacedV1(rq.Namespace).Set(ctx, rq.Key, rq.Value, rq.Expire)
}

// Del removes a key from the cache by namespace.
func (s *CacheServerService) Del(ctx context.Context, rq *v1.DelRequest) (*emptypb.Empty, error) {
	return s.biz.NamespacedV1(rq.Namespace).Del(ctx, rq.Key)
}

// Get retrieves a key's value from the cache by namespace.
func (s *CacheServerService) Get(ctx context.Context, rq *v1.GetRequest) (*v1.GetResponse, error) {
	return s.biz.NamespacedV1(rq.Namespace).Get(ctx, rq.Key)
}

// SetSecret stores a secret in the system or updates an existing one.
func (s *CacheServerService) SetSecret(ctx context.Context, rq *v1.SetSecretRequest) (*emptypb.Empty, error) {
	return s.biz.SecretV1().Set(ctx, rq)
}

// DelSecret removes a secret from the system.
func (s *CacheServerService) DelSecret(ctx context.Context, rq *v1.DelSecretRequest) (*emptypb.Empty, error) {
	return s.biz.SecretV1().Del(ctx, rq)
}

// GetSecret retrieves a secret from the system.
func (s *CacheServerService) GetSecret(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error) {
	return s.biz.SecretV1().Get(ctx, rq)
}
