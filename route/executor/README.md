# StatefulExecutor - 有状态服务执行器

## 概述

`StatefulExecutor` 是一个用Go语言实现的有状态服务管理执行器，功能与Java版本的 `StatefulRedisExecutor` 完全一致。它提供了完整的服务状态管理、Pod链接管理和工作负载状态管理功能。

## 主要特性

### 🚀 核心功能
- **服务状态管理**: 设置和获取特定服务的Pod状态
- **工作负载状态管理**: 管理整个工作负载（服务）的状态
- **Pod链接管理**: 建立、维护和移除Pod与UID之间的链接
- **Lua脚本支持**: 使用Redis Lua脚本确保操作的原子性

### 🔧 技术特性
- **Go语言原生**: 完全用Go语言实现，符合Go语言最佳实践
- **Redis集成**: 基于 `go-redis/v9` 客户端库
- **Lua脚本**: 支持所有必要的Lua脚本操作
- **结构化日志**: 使用Kratos日志系统
- **上下文支持**: 支持超时、取消等上下文控制

## 架构设计

### 单文件架构设计
所有功能都集成在 `stateful_executor.go` 单一文件中，采用Go语言的最佳实践：

- **接口实现**: 实现 `route.StatefulExecutor` 接口
- **Lua脚本管理**: 使用 `embed.FS` 嵌入Lua脚本，支持脚本预加载和缓存
- **Redis操作封装**: 封装所有Redis操作，提供统一的错误处理
- **错误处理和日志**: 集成Kratos日志系统，提供结构化日志记录

### 核心组件
```go
type StatefulExecutorImpl struct {
    redisClient *redis.Client           // Redis客户端
    logger      log.Logger              // 日志记录器
    scriptCache map[string]string      // Lua脚本缓存
    scriptSHAs  map[string]string      // 脚本SHA缓存
}
```

### 接口定义位置
核心接口定义位于 `route/interfaces.go`，包括：
- `StatefulExecutor`: 主要业务接口
- `RedisManager`: Redis管理接口
- 其他相关接口和类型定义

## 使用方法

### 1. 创建执行器实例

```go
import (
    "github.com/go-kratos/kratos/v2/log"
    "github.com/redis/go-redis/v9"
    "github.com/go-kratos/kratos/v2/route/executor"
)

// 创建Redis客户端
redisClient := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// 创建日志记录器
logger := log.DefaultLogger

// 创建执行器实例
executor := executor.NewStatefulExecutor(redisClient, logger)
```

### 2. 服务状态管理

```go
ctx := context.Background()

// 设置服务状态
err := executor.SetServiceState(ctx, "default", "my-service", 1, "Running")
if err != nil {
    log.Error("设置服务状态失败", err)
}

// 获取服务状态
podStates, err := executor.GetServiceState(ctx, "default", "my-service")
if err != nil {
    log.Error("获取服务状态失败", err)
}
// podStates: map[int]string{1: "Running", 2: "Pending"}
```

### 3. 工作负载状态管理

```go
// 设置工作负载状态
err := executor.SetWorkloadState(ctx, "default", "my-service", "Available")
if err != nil {
    log.Error("设置工作负载状态失败", err)
}

// 获取工作负载状态
state, err := executor.GetWorkloadState(ctx, "default", "my-service")
if err != nil {
    log.Error("获取工作负载状态失败", err)
}

// 批量获取工作负载状态
serviceNames := []string{"service1", "service2", "service3"}
states, err := executor.GetWorkloadStateBatch(ctx, "default", serviceNames)
if err != nil {
    log.Error("批量获取工作负载状态失败", err)
}
```

### 4. Pod链接管理

```go
// 设置Pod链接
podID, err := executor.SetLinkedPod(ctx, "default", "user123", "my-service", 1, 3600)
if err != nil {
    log.Error("设置Pod链接失败", err)
}

// 尝试设置Pod链接
success, currentPodID, err := executor.TrySetLinkedPod(ctx, "default", "user123", "my-service", 2, 3600)
if err != nil {
    log.Error("尝试设置Pod链接失败", err)
}

// 获取链接的Pod
podID, err := executor.GetLinkedPod(ctx, "default", "user123", "my-service")
if err != nil {
    log.Error("获取链接的Pod失败", err)
}

// 移除Pod链接
success, err := executor.RemoveLinkedPod(ctx, "default", "user123", "my-service", 300)
if err != nil {
    log.Error("移除Pod链接失败", err)
}

// 获取链接的服务
services, err := executor.GetLinkService(ctx, "default", "user123")
if err != nil {
    log.Error("获取链接的服务失败", err)
}
```

## Lua脚本支持

### 脚本嵌入
使用Go的 `embed.FS` 将Lua脚本直接嵌入到二进制文件中：

```go
//go:embed lua_scripts/*.lua
var luaScripts embed.FS
```

### 预加载脚本
执行器启动时会自动预加载以下Lua脚本：
- `statefulSetLink`: 设置Pod链接
- `statefulSetLinkIfAbsent`: 仅在不存在时设置Pod链接
- `statefulTrySetLink`: 尝试设置Pod链接
- `statefulRemoveLink`: 移除Pod链接
- `statefulRemoveLinkWithId`: 移除指定Pod ID的链接
- `statefulGetLinkIfPersist`: 获取持久链接的Pod
- `statefulComputeLinkIfAbsent`: 计算链接（如果不存在）
- `statefulGetServicePod`: 获取服务Pod
- `statefulGetService`: 获取服务信息
- `statefulSetState`: 设置状态
- `statefulGetLinkService`: 获取链接服务

### 脚本执行
```go
// 直接执行脚本
result, err := executor.executeScript(ctx, "scriptName", keys, args)
```

## Redis键格式

### 服务状态键
```
sf:{<namespace>:state:<serviceName>}
```

### 工作负载状态键
```
sf:{<namespace>:workload:<serviceName>}
```

### UID链接键
```
sf:{<namespace>:lk:<uid>}
```

### UID与特定服务链接键
```
sf:{<namespace>:lk:<uid>}<serviceName>
```

## 错误处理

所有方法都返回标准的Go错误，支持错误包装和类型断言：

```go
if err != nil {
    // 处理错误
    log.Error("操作失败", err)
    return err
}
```

## 性能优化

### 连接池管理
- 使用Redis客户端内置的连接池
- 支持连接池大小和空闲连接数配置

### 脚本缓存
- Lua脚本在启动时预加载到Redis
- 本地缓存脚本内容和SHA，避免重复读取和计算

### 批量操作
- 支持批量获取工作负载状态
- 支持批量获取Pod链接

## 与Java版本对比

| 功能 | Java版本 | Go版本 | 状态 |
|------|----------|--------|------|
| 服务状态管理 | ✅ | ✅ | 完全一致 |
| 工作负载状态管理 | ✅ | ✅ | 完全一致 |
| Pod链接管理 | ✅ | ✅ | 完全一致 |
| Lua脚本支持 | ✅ | ✅ | 完全一致 |
| 批量操作 | ✅ | ✅ | 完全一致 |
| 错误处理 | ✅ | ✅ | 完全一致 |
| 日志记录 | ✅ | ✅ | 完全一致 |

## 测试

### 编译测试
```bash
cd route/executor
go build -o stateful_executor_test .
```

### 接口实现测试
```bash
go test -v -run TestInterfaceImplementation
```

### 运行示例
```bash
go run example_v9_usage.go
```

## 项目结构

```
route/
├── interfaces.go                    # 所有接口定义
├── types.go                         # 数据模型定义
└── executor/
    ├── lua_scripts/                 # Lua脚本文件（嵌入到二进制）
    ├── stateful_executor.go         # 完整实现（单文件架构）
    ├── interface_test.go            # 接口实现测试
    ├── example_v9_usage.go          # 使用示例
    └── README.md                    # 项目说明
```

## 依赖

- Go 1.19+
- `github.com/go-kratos/kratos/v2` v2.x
- `github.com/redis/go-redis/v9` v9.x

## 许可证

遵循项目整体许可证。
