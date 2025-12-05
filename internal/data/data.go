package data

import (
	"context"

	"github.com/dgraph-io/ristretto"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"cacheserver/internal/biz/namespaced"
	"cacheserver/internal/biz/secret"
	"cacheserver/internal/conf"
	"cacheserver/pkg/cache"
	redisstore "cacheserver/pkg/cache/store/redis"
	ristrettostore "cacheserver/pkg/cache/store/ristretto"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewGreeterRepo,
	NewNamespacedCache,
	NewSecretChainCache,
	wire.Bind(new(namespaced.Cache), new(*namespacedCache)),
	wire.Bind(new(secret.SecretStore), new(*secretChainStore)),
)

// Data .
type Data struct {
	db         *gorm.DB
	rdb        *redis.Client
	localCache *ristretto.Cache
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	helper := log.NewHelper(logger)

	// Initialize MySQL
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	// Auto migrate
	if err := db.AutoMigrate(&SecretModel{}); err != nil {
		return nil, nil, err
	}

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})

	// Test Redis connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		helper.Warnf("failed to connect to redis: %v", err)
	}

	// Initialize Ristretto local cache
	localCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000,   // number of keys to track frequency
		MaxCost:     1 << 20, // maximum cost of cache (1MB)
		BufferItems: 64,      // number of keys per Get buffer
	})
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		helper.Info("closing the data resources")
		localCache.Close()
		if err := rdb.Close(); err != nil {
			helper.Errorf("failed to close redis: %v", err)
		}
	}

	return &Data{db: db, rdb: rdb, localCache: localCache}, cleanup, nil
}

// DB returns the database connection.
func (d *Data) DB() *gorm.DB {
	return d.db
}

// RDB returns the Redis client.
func (d *Data) RDB() *redis.Client {
	return d.rdb
}

// LocalCache returns the local Ristretto cache.
func (d *Data) LocalCache() *ristretto.Cache {
	return d.localCache
}

// NewNamespacedCache creates a two-level cache (Local + Redis) for namespaced data.
func NewNamespacedCache(data *Data, logger log.Logger) *namespacedCache {
	helper := log.NewHelper(logger)

	// Level 1: Local Ristretto cache
	localStore := ristrettostore.NewRistretto(data.LocalCache())
	localCache := cache.New[any](localStore)

	// Level 2: Redis cache
	redisStore := redisstore.NewRedis(data.RDB())
	redisCache := cache.New[any](redisStore)

	// Create chain cache: Local -> Redis
	chainCache := cache.NewChain[any](localCache, redisCache)

	helper.Info("initialized two-level cache: Local(Ristretto) -> Redis")

	return &namespacedCache{chain: chainCache, log: helper}
}

// NewSecretChainCache creates a three-level cache (Local + Redis + MySQL) for secrets.
func NewSecretChainCache(data *Data, logger log.Logger) *secretChainStore {
	helper := log.NewHelper(logger)

	// Level 1: Local Ristretto cache
	localStore := ristrettostore.NewRistretto(data.LocalCache())
	localCache := cache.New[any](localStore)

	// Level 2: Redis cache
	redisStore := redisstore.NewRedis(data.RDB())
	redisCache := cache.New[any](redisStore)

	// Level 3: MySQL store
	mysqlStore := NewMySQLSecretStore(data.DB())
	mysqlCache := cache.New[any](mysqlStore)

	// Create chain cache: Local -> Redis -> MySQL
	chainCache := cache.NewChain[any](localCache, redisCache, mysqlCache)

	helper.Info("initialized three-level cache: Local(Ristretto) -> Redis -> MySQL")

	return &secretChainStore{chain: chainCache, log: helper}
}
