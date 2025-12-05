# Cache - 多级缓存包

提供通用的多级缓存实现，支持链式缓存和异步回填。

## 目录结构

```
cache/
├── cache.go              # Cache 接口和 DelegateCache 实现
├── chain.go              # ChainCache 链式缓存实现
└── store/                # 存储后端
    ├── store.go          # Store 接口定义
    ├── redis/            # Redis 存储实现
    │   └── redis.go
    └── ristretto/        # Ristretto 本地缓存实现
        └── ristretto.go
```

## 核心接口

### Cache[T any]

泛型缓存接口：

```go
type Cache[T any] interface {
    Set(ctx context.Context, key any, obj T) error
    Get(ctx context.Context, key any) (T, error)
    SetWithTTL(ctx context.Context, key any, obj T, ttl time.Duration) error
    GetWithTTL(ctx context.Context, key any) (T, time.Duration, error)
    Del(ctx context.Context, key any) error
    Clear(ctx context.Context) error
    Wait(ctx context.Context)
}
```

### Store

存储后端接口：

```go
type Store interface {
    Get(ctx context.Context, key any) (any, error)
    GetWithTTL(ctx context.Context, key any) (any, time.Duration, error)
    Set(ctx context.Context, key any, value any) error
    SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) error
    Del(ctx context.Context, key any) error
    Clear(ctx context.Context) error
    Wait(ctx context.Context)
}
```

## ChainCache

链式缓存实现，支持多级缓存和异步回填。

### 创建

```go
// 创建三级缓存链
chain := cache.NewChain[any](
    localCache,   // Level 1: Ristretto
    redisCache,   // Level 2: Redis
    mysqlCache,   // Level 3: MySQL
)
```

### 工作原理

#### 写入 (Set)

同时写入所有缓存层：

```
Client → L1 → L2 → L3
```

#### 读取 (Get)

从第一层开始逐层查找，命中后返回并异步回填上层：

```
Client ← L1 (miss) ← L2 (miss) ← L3 (hit)
         ↑           ↑
         └───────────┴── 异步回填
```

#### 删除 (Del)

从所有缓存层删除：

```
Client → L1 (del) → L2 (del) → L3 (del)
```

### 异步回填

当从下层缓存读取到数据后，会通过 channel 异步回填到上层缓存：

```go
func (c *ChainCache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
    for _, cache := range c.caches {
        obj, ttl, err = cache.GetWithTTL(ctx, key)
        if err == nil {
            // 异步回填到上层缓存
            c.setChannel <- &chainKeyValue[T]{key, obj, ttl, cache.id}
            return obj, ttl, nil
        }
    }
    return obj, ttl, err
}
```

## Store 实现

### RistrettoStore

基于 [dgraph-io/ristretto](https://github.com/dgraph-io/ristretto) 的本地内存缓存：

```go
store := ristretto.NewRistretto(ristrettoClient)
cache := cache.New[any](store)
```

特点：
- 高性能本地缓存
- 支持 TTL
- 自动淘汰策略

### RedisStore

基于 [redis/go-redis](https://github.com/redis/go-redis) 的分布式缓存：

```go
store := redis.NewRedis(redisClient)
cache := cache.New[any](store)
```

特点：
- 分布式缓存
- 支持 TTL
- 跨实例共享

## 使用示例

```go
import (
    "cacheserver/pkg/cache"
    "cacheserver/pkg/cache/store/redis"
    "cacheserver/pkg/cache/store/ristretto"
)

// 创建本地缓存
localStore := ristretto.NewRistretto(ristrettoClient)
localCache := cache.New[string](localStore)

// 创建 Redis 缓存
redisStore := redis.NewRedis(redisClient)
redisCache := cache.New[string](redisStore)

// 创建链式缓存
chain := cache.NewChain[string](localCache, redisCache)

// 使用
chain.Set(ctx, "key", "value")
value, err := chain.Get(ctx, "key")
chain.Del(ctx, "key")
```

## KeyGetter 接口

支持自定义缓存 key 生成：

```go
type KeyGetter interface {
    CacheKey() string
}

// 示例
type NamespacedKey struct {
    Namespace string
    Key       string
}

func (k NamespacedKey) CacheKey() string {
    return fmt.Sprintf("namespace:%s:%s", k.Namespace, k.Key)
}
```
