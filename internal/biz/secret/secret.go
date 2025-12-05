package secret

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "cacheserver/api/cacheserver/v1"
)

// SecretBiz defines the interface for handling secret requests.
type SecretBiz interface {
	Set(ctx context.Context, rq *v1.SetSecretRequest) (*emptypb.Empty, error)
	Del(ctx context.Context, rq *v1.DelSecretRequest) (*emptypb.Empty, error)
	Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error)
}

// SecretM represents a secret model.
type SecretM struct {
	ID          int64
	UserID      string
	Name        string
	SecretID    string
	SecretKey   string
	Expires     int64
	Status      int32
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SecretStore defines the interface for secret storage operations.
type SecretStore interface {
	Set(ctx context.Context, key string, value *SecretM) error
	Get(ctx context.Context, key string) (*SecretM, error)
	Del(ctx context.Context, key string) error
}

// secretBiz is the implementation of SecretBiz.
type secretBiz struct {
	store SecretStore
}

// Ensure that *secretBiz implements the SecretBiz.
var _ SecretBiz = (*secretBiz)(nil)

// New creates and returns a new instance of *secretBiz.
func New(store SecretStore) SecretBiz {
	return &secretBiz{store: store}
}

// Set stores a secret in the cache.
func (b *secretBiz) Set(ctx context.Context, rq *v1.SetSecretRequest) (*emptypb.Empty, error) {
	secret := &SecretM{
		Name:        rq.Name,
		SecretID:    rq.Key,
		Description: rq.Description,
	}
	if rq.Expire != nil {
		secret.Expires = time.Now().Add(rq.Expire.AsDuration()).Unix()
	}

	return &emptypb.Empty{}, b.store.Set(ctx, rq.Key, secret)
}

// Del deletes a secret from the cache.
func (b *secretBiz) Del(ctx context.Context, rq *v1.DelSecretRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, b.store.Del(ctx, rq.Key)
}

// Get retrieves a secret from the cache.
func (b *secretBiz) Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error) {
	secret, err := b.store.Get(ctx, rq.Key)
	if err != nil {
		return nil, err
	}

	return &v1.GetSecretResponse{
		UserID:      secret.UserID,
		Name:        secret.Name,
		SecretID:    secret.SecretID,
		SecretKey:   secret.SecretKey,
		Expires:     secret.Expires,
		Status:      secret.Status,
		Description: secret.Description,
		CreatedAt:   timestamppb.New(secret.CreatedAt),
		UpdatedAt:   timestamppb.New(secret.UpdatedAt),
	}, nil
}
