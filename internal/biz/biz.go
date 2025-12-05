package biz

import (
	"github.com/google/wire"

	"cacheserver/internal/biz/namespaced"
	"cacheserver/internal/biz/secret"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewGreeterUsecase, NewCacheBiz, wire.Bind(new(ICacheBiz), new(*CacheBiz)))

// ICacheBiz defines the methods that must be implemented by the cache business layer.
type ICacheBiz interface {
	NamespacedV1(namespace string) namespaced.NamespacedBiz
	SecretV1() secret.SecretBiz
}

// CacheBiz is a concrete implementation of ICacheBiz.
type CacheBiz struct {
	cache       namespaced.Cache
	secretStore secret.SecretStore
}

// Ensure that CacheBiz implements the ICacheBiz.
var _ ICacheBiz = (*CacheBiz)(nil)

// NewCacheBiz creates an instance of ICacheBiz.
func NewCacheBiz(cache namespaced.Cache, secretStore secret.SecretStore) *CacheBiz {
	return &CacheBiz{cache: cache, secretStore: secretStore}
}

// NamespacedV1 returns an instance that implements the NamespacedBiz.
func (b *CacheBiz) NamespacedV1(namespace string) namespaced.NamespacedBiz {
	return namespaced.New(b.cache, namespace)
}

// SecretV1 returns an instance that implements the SecretBiz.
func (b *CacheBiz) SecretV1() secret.SecretBiz {
	return secret.New(b.secretStore)
}
