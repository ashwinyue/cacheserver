# CacheServer

基于 [Kratos](https://go-kratos.dev/) 框架实现的多级缓存服务，复刻自 [onex-cacheserver](https://github.com/onexstack/onex)。

## 功能特性

- **三级缓存架构**: Local (Ristretto) → Redis → MySQL
- **命名空间隔离**: 支持按命名空间隔离缓存数据
- **Secret 管理**: 支持密钥的存储、查询和删除
- **异步缓存回填**: 从下层缓存读取后自动回填上层缓存
- **gRPC API**: 提供完整的 gRPC 接口

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                      CacheServer                             │
├─────────────────────────────────────────────────────────────┤
│  gRPC API                                                    │
│  ├── Set / Get / Del          (Namespaced Cache)            │
│  └── SetSecret / GetSecret / DelSecret  (Secret Cache)      │
├─────────────────────────────────────────────────────────────┤
│  Service Layer (internal/service)                            │
├─────────────────────────────────────────────────────────────┤
│  Business Layer (internal/biz)                               │
│  ├── NamespacedBiz            │  SecretBiz                  │
├─────────────────────────────────────────────────────────────┤
│  Data Layer (internal/data)                                  │
│  └── ChainCache                                              │
│      ├── Level 1: Ristretto (Local Memory)                  │
│      ├── Level 2: Redis                                      │
│      └── Level 3: MySQL (Secret only)                       │
└─────────────────────────────────────────────────────────────┘
```

## 快速开始

### 前置条件

- Go 1.21+
- Docker & Docker Compose
- Make

### 1. 启动依赖服务

```bash
# 启动 MySQL 和 Redis
make docker
```

### 2. 运行服务

```bash
# 本地运行
make run
```

服务启动后：
- HTTP: http://localhost:8000
- gRPC: localhost:9000

### 3. 测试 API

```bash
# 安装 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 查看服务列表
grpcurl -plaintext localhost:9000 list

# 设置缓存
grpcurl -plaintext -d '{
  "namespace": "test",
  "key": "user:1",
  "value": {"@type": "type.googleapis.com/google.protobuf.StringValue", "value": "hello"}
}' localhost:9000 cacheserver.v1.CacheServer/Set

# 获取缓存
grpcurl -plaintext -d '{"namespace": "test", "key": "user:1"}' \
  localhost:9000 cacheserver.v1.CacheServer/Get

# 设置 Secret
grpcurl -plaintext -d '{
  "key": "api-key-1",
  "name": "My API Key",
  "description": "Test API Key"
}' localhost:9000 cacheserver.v1.CacheServer/SetSecret

# 获取 Secret
grpcurl -plaintext -d '{"key": "api-key-1"}' \
  localhost:9000 cacheserver.v1.CacheServer/GetSecret
```

## 项目结构

```
cacheserver/
├── api/                          # API 定义
│   └── cacheserver/v1/           # CacheServer gRPC API
│       ├── cacheserver.proto     # 服务定义
│       ├── namespaced.proto      # 命名空间缓存消息
│       └── secret.proto          # Secret 消息
├── cmd/                          # 应用入口
│   └── cacheserver/
│       ├── main.go
│       ├── wire.go               # Wire 依赖注入
│       └── wire_gen.go
├── configs/                      # 配置文件
│   ├── config.yaml               # Docker 环境配置
│   └── config.local.yaml         # 本地开发配置
├── internal/                     # 内部代码
│   ├── biz/                      # 业务逻辑层
│   │   ├── namespaced/           # 命名空间缓存业务
│   │   └── secret/               # Secret 业务
│   ├── conf/                     # 配置结构
│   ├── data/                     # 数据访问层
│   │   ├── cache.go              # 缓存实现
│   │   ├── data.go               # 数据初始化
│   │   └── secret.go             # Secret 存储
│   ├── server/                   # 服务器配置
│   └── service/                  # 服务层
│       └── cacheserver.go        # CacheServer 服务
├── pkg/                          # 公共包
│   └── cache/                    # 缓存包
│       ├── cache.go              # 缓存接口
│       ├── chain.go              # 链式缓存
│       └── store/                # 存储实现
│           ├── redis/            # Redis 存储
│           ├── ristretto/        # Ristretto 存储
│           └── store.go          # 存储接口
├── docker-compose.yaml           # Docker Compose 配置
├── Dockerfile                    # Docker 构建文件
├── Makefile                      # 构建脚本
└── COMPARISON_REPORT.md          # 功能对比报告
```

## API 接口

### CacheServer Service

| 方法 | 描述 | 缓存层级 |
|------|------|----------|
| `Set` | 设置命名空间缓存 | Local → Redis |
| `Get` | 获取命名空间缓存 | Local → Redis |
| `Del` | 删除命名空间缓存 | Local → Redis |
| `SetSecret` | 设置 Secret | Local → Redis → MySQL |
| `GetSecret` | 获取 Secret | Local → Redis → MySQL |
| `DelSecret` | 删除 Secret | Local → Redis → MySQL |

### 消息定义

```protobuf
// 命名空间缓存
message SetRequest {
  string namespace = 1;
  string key = 2;
  google.protobuf.Any value = 3;
  optional google.protobuf.Duration expire = 4;
}

// Secret
message SetSecretRequest {
  string key = 1;
  string name = 2;
  optional google.protobuf.Duration expire = 3;
  string description = 4;
}
```

## 配置说明

```yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s

data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/cacheserver?parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 0.2s
    write_timeout: 0.2s
```

## 开发指南

### 生成代码

```bash
# 生成 API 代码
make api

# 生成配置代码
make config

# 生成 Wire 依赖注入
make generate

# 生成所有
make all
```

### 构建

```bash
# 本地构建
make build

# Docker 构建
docker build -t cacheserver .
```

### Docker Compose

```bash
# 启动所有服务（含应用）
make docker-all

# 仅启动依赖服务
make docker

# 停止所有服务
make docker-down
```

## 缓存机制

### ChainCache 工作原理

1. **写入 (Set)**: 同时写入所有缓存层
2. **读取 (Get)**: 从 Level 1 开始逐层查找，命中后返回
3. **回填 (Backfill)**: 从下层读取后，异步回填到上层缓存
4. **删除 (Del)**: 从所有缓存层删除

```
Write: Client → L1 (Ristretto) → L2 (Redis) → L3 (MySQL)
Read:  Client ← L1 ← L2 ← L3 (miss时逐层查找，命中后回填)
```

### 缓存配置

| 缓存层 | 类型 | 用途 |
|--------|------|------|
| L1 | Ristretto | 本地内存缓存，最快访问 |
| L2 | Redis | 分布式缓存，跨实例共享 |
| L3 | MySQL | 持久化存储，数据不丢失 |

## 许可证

MIT License
