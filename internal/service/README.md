# Service - 服务层

服务层负责实现 gRPC/HTTP 接口，将请求转发给业务层处理。

## 目录结构

```
service/
├── service.go        # Wire ProviderSet
├── greeter.go        # Greeter 示例服务
└── cacheserver.go    # CacheServer 服务实现
```

## CacheServerService

实现 `cacheserver.v1.CacheServerServer` 接口：

```go
type CacheServerService struct {
    v1.UnimplementedCacheServerServer
    biz biz.ICacheBiz
}
```

### 接口实现

| 方法 | 描述 |
|------|------|
| `Set` | 设置命名空间缓存 |
| `Get` | 获取命名空间缓存 |
| `Del` | 删除命名空间缓存 |
| `SetSecret` | 设置 Secret |
| `GetSecret` | 获取 Secret |
| `DelSecret` | 删除 Secret |

### 示例代码

```go
func (s *CacheServerService) Set(ctx context.Context, rq *v1.SetRequest) (*emptypb.Empty, error) {
    return s.biz.NamespacedV1(rq.Namespace).Set(ctx, rq.Key, rq.Value, rq.Expire)
}

func (s *CacheServerService) GetSecret(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error) {
    return s.biz.SecretV1().Get(ctx, rq)
}
```

## Wire 依赖注入

```go
var ProviderSet = wire.NewSet(NewGreeterService, NewCacheServerService)
