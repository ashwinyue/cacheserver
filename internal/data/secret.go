package data

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"cacheserver/internal/biz/secret"
	"cacheserver/pkg/cache"
	"cacheserver/pkg/cache/store"
)

// SecretModel represents the database model for secrets.
type SecretModel struct {
	gorm.Model
	UserID      string `gorm:"column:user_id;type:varchar(64)"`
	Name        string `gorm:"column:name;type:varchar(253)"`
	SecretID    string `gorm:"column:secret_id;type:varchar(64);uniqueIndex"`
	SecretKey   string `gorm:"column:secret_key;type:varchar(255)"`
	Expires     int64  `gorm:"column:expires"`
	Status      int32  `gorm:"column:status;default:1"`
	Description string `gorm:"column:description;type:varchar(256)"`
}

// TableName returns the table name for SecretModel.
func (SecretModel) TableName() string {
	return "secrets"
}

// secretChainStore implements the secret.SecretStore interface using chain cache.
type secretChainStore struct {
	chain *cache.ChainCache[any]
	log   *log.Helper
}

// Set stores or updates a secret in the chain cache.
func (s *secretChainStore) Set(ctx context.Context, key string, value *secret.SecretM) error {
	// Serialize to JSON for storage in cache layers
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.chain.Set(ctx, key, string(data))
}

// Get retrieves a secret from the chain cache.
func (s *secretChainStore) Get(ctx context.Context, key string) (*secret.SecretM, error) {
	result, err := s.chain.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var secretM secret.SecretM
	if str, ok := result.(string); ok {
		if err := json.Unmarshal([]byte(str), &secretM); err != nil {
			return nil, err
		}
	}
	return &secretM, nil
}

// Del removes a secret from the chain cache.
func (s *secretChainStore) Del(ctx context.Context, key string) error {
	return s.chain.Del(ctx, key)
}

// mysqlSecretStore implements store.Store interface for MySQL.
type mysqlSecretStore struct {
	db *gorm.DB
}

// NewMySQLSecretStore creates a new MySQL secret store.
func NewMySQLSecretStore(db *gorm.DB) *mysqlSecretStore {
	return &mysqlSecretStore{db: db}
}

// Get retrieves a secret from MySQL.
func (s *mysqlSecretStore) Get(ctx context.Context, key any) (any, error) {
	var model SecretModel
	if err := s.db.WithContext(ctx).Where(SecretModel{SecretID: key.(string)}).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, store.ErrKeyNotFound
		}
		return nil, err
	}

	secretM := &secret.SecretM{
		ID:          int64(model.ID),
		UserID:      model.UserID,
		Name:        model.Name,
		SecretID:    model.SecretID,
		SecretKey:   model.SecretKey,
		Expires:     model.Expires,
		Status:      model.Status,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	// Return as JSON string for consistency with other cache layers
	data, err := json.Marshal(secretM)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// GetWithTTL retrieves a secret and its TTL from MySQL.
func (s *mysqlSecretStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return nil, 0, err
	}
	// MySQL doesn't have TTL, return 0
	return value, 0, nil
}

// Set stores a secret in MySQL.
func (s *mysqlSecretStore) Set(ctx context.Context, key any, value any) error {
	var secretM secret.SecretM
	if str, ok := value.(string); ok {
		if err := json.Unmarshal([]byte(str), &secretM); err != nil {
			return err
		}
	}

	model := &SecretModel{
		UserID:      secretM.UserID,
		Name:        secretM.Name,
		SecretID:    secretM.SecretID,
		SecretKey:   secretM.SecretKey,
		Expires:     secretM.Expires,
		Status:      secretM.Status,
		Description: secretM.Description,
	}

	return s.db.WithContext(ctx).Where(SecretModel{SecretID: key.(string)}).
		Assign(model).
		FirstOrCreate(model).Error
}

// SetWithTTL stores a secret in MySQL (TTL is ignored for MySQL).
func (s *mysqlSecretStore) SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) error {
	return s.Set(ctx, key, value)
}

// Del removes a secret from MySQL.
func (s *mysqlSecretStore) Del(ctx context.Context, key any) error {
	err := s.db.WithContext(ctx).Where(SecretModel{SecretID: key.(string)}).Delete(&SecretModel{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

// Clear clears all secrets from MySQL.
func (s *mysqlSecretStore) Clear(ctx context.Context) error {
	return s.db.WithContext(ctx).Where("1 = 1").Delete(&SecretModel{}).Error
}

// Wait waits for all operations to complete.
func (s *mysqlSecretStore) Wait(ctx context.Context) {}
