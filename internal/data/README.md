# Data - 数据访问层

数据访问层负责实现数据的存储和读取，包括多级缓存的实现。

## 目录结构

```
data/
├── data.go       # 数据层初始化，Wire ProviderSet
├── greeter.go    # Greeter 示例数据访问
├── cache.go      # 命名空间缓存实现 (namespacedCache)
└── secret.go     # Secret 存储实现 (secretChainStore, mysqlSecretStore)
```

## 核心组件

### Data 结构

```go
type Data struct {
    db         *gorm.DB          // MySQL 连接
    rdb        *redis.Client     // Redis 客户端
    localCache *ristretto.Cache  // 本地 Ristretto 缓存
}
```

### 缓存实现

#### namespacedCache (两级缓存)

用于命名空间缓存，实现 `namespaced.Cache` 接口：

```
Level 1: Ristretto (本地内存)
Level 2: Redis
```

#### secretChainStore (三级缓存)

用于 Secret 存储，实现 `secret.SecretStore` 接口：

```
Level 1: Ristretto (本地内存)
Level 2: Redis
Level 3: MySQL
```

### MySQL 模型

```go
type SecretModel struct {
    gorm.Model
    UserID      string
    Name        string
    SecretID    string  // 唯一索引
    SecretKey   string
    Expires     int64
    Status      int32
    Description string
}
```

## Wire 依赖注入

```go
var ProviderSet = wire.NewSet(
    NewData,
    NewGreeterRepo,
    NewNamespacedCache,
    NewSecretChainCache,
    wire.Bind(new(namespaced.Cache), new(*namespacedCache)),
    wire.Bind(new(secret.SecretStore), new(*secretChainStore)),
)
```

## 初始化流程

1. 连接 MySQL 数据库
2. 自动迁移数据库表
3. 连接 Redis
4. 初始化 Ristretto 本地缓存
5. 创建 ChainCache 实例
