package ristretto

import (
	"context"
	"fmt"
	"time"

	"cacheserver/pkg/cache/store"
)

// RistrettoClientInterface represents a dgraph-io/ristretto client.
type RistrettoClientInterface interface {
	Get(key any) (any, bool)
	Set(key, value any, cost int64) bool
	SetWithTTL(key, value any, cost int64, ttl time.Duration) bool
	Del(key any)
	Clear()
	Wait()
}

// RistrettoStore is a store for Ristretto (memory) library.
type RistrettoStore struct {
	client RistrettoClientInterface
}

// NewRistretto creates a new store to Ristretto (memory) library instance.
func NewRistretto(client RistrettoClientInterface) *RistrettoStore {
	return &RistrettoStore{
		client: client,
	}
}

// Get returns data stored from a given key.
func (s *RistrettoStore) Get(_ context.Context, key any) (any, error) {
	value, exists := s.client.Get(key)
	if !exists {
		return nil, store.ErrKeyNotFound
	}
	return value, nil
}

// GetWithTTL returns data stored from a given key and its corresponding TTL.
func (s *RistrettoStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	return value, 0, err
}

// Set defines data in Ristretto memory cache for given key identifier.
func (s *RistrettoStore) Set(_ context.Context, key any, value any) error {
	if set := s.client.Set(key, value, 1); !set {
		return fmt.Errorf("failed to set value for key '%v'", key)
	}
	return nil
}

// SetWithTTL defines data in Ristretto memory cache with TTL.
func (s *RistrettoStore) SetWithTTL(_ context.Context, key any, value any, ttl time.Duration) error {
	if set := s.client.SetWithTTL(key, value, 1, ttl); !set {
		return fmt.Errorf("failed to set value for key '%v'", key)
	}
	return nil
}

// Del removes data in Ristretto memory cache for given key identifier.
func (s *RistrettoStore) Del(_ context.Context, key any) error {
	s.client.Del(key)
	return nil
}

// Clear resets all data in the store.
func (s *RistrettoStore) Clear(_ context.Context) error {
	s.client.Clear()
	return nil
}

// Wait waits for all operations to complete.
func (s *RistrettoStore) Wait(_ context.Context) {
	s.client.Wait()
}
