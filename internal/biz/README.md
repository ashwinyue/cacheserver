# Biz - 业务逻辑层

业务逻辑层负责实现核心业务规则，不依赖具体的存储实现。

## 目录结构

```
biz/
├── biz.go              # 业务接口定义和 Wire ProviderSet
├── greeter.go          # Greeter 示例业务
├── namespaced/         # 命名空间缓存业务
│   └── namespaced.go   # NamespacedBiz 实现
└── secret/             # Secret 业务
    └── secret.go       # SecretBiz 实现
```

## 核心接口

### ICacheBiz

```go
type ICacheBiz interface {
    NamespacedV1(namespace string) namespaced.NamespacedBiz
    SecretV1() secret.SecretBiz
}
```

### NamespacedBiz

处理命名空间缓存的业务逻辑：

```go
type NamespacedBiz interface {
    Set(ctx context.Context, key string, value *anypb.Any, ttl *durationpb.Duration) (*emptypb.Empty, error)
    Del(ctx context.Context, key string) (*emptypb.Empty, error)
    Get(ctx context.Context, key string) (*v1.GetResponse, error)
}
```

### SecretBiz

处理 Secret 管理的业务逻辑：

```go
type SecretBiz interface {
    Set(ctx context.Context, rq *v1.SetSecretRequest) (*emptypb.Empty, error)
    Del(ctx context.Context, rq *v1.DelSecretRequest) (*emptypb.Empty, error)
    Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error)
}
```

## 设计原则

1. **接口隔离**: 每个业务模块定义独立的接口
2. **依赖倒置**: 业务层定义接口，数据层实现接口
3. **单一职责**: 每个业务模块只处理一类业务
