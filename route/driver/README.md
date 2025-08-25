# State模块

State模块是有状态路由系统的核心组件，实现了客户端和服务端驱动，负责管理服务状态、负载均衡和用户链接关系。

## 功能特性

### 客户端驱动 (StatefulRouteForClientDriver)
- **Pod计算**: 计算最佳可用Pod，支持缓存和Redis查询
- **链接管理**: 管理用户与Pod的链接关系
- **缓存策略**: 本地缓存优化，支持过期和更新
- **命名空间隔离**: 只支持本地命名空间操作

### 服务端驱动 (StatefulRouteForServerDriver)
- **状态管理**: 管理本地Pod的服务状态、负载状态、路由状态
- **定时更新**: 定时更新服务状态和工作负载状态
- **链接操作**: 支持Pod链接的设置、尝试设置、移除等操作
- **配置验证**: 验证基础配置的有效性

### 驱动独立性
- **独立生命周期**: 客户端和服务端驱动各自管理自己的生命周期
- **直接管理**: 应用层可以直接管理驱动的启动、停止
- **简化架构**: 减少不必要的抽象层，提高代码可维护性

## 架构设计

```
┌─────────────────┐    ┌─────────────────┐
│  客户端驱动      │    │  服务端驱动      │
│                 │    │                 │
│ • 计算最佳Pod   │    │ • 状态管理      │
│ • 缓存管理      │    │ • 定时更新      │
│ • 链接管理      │    │ • 链接操作      │
│ • 独立生命周期  │    │ • 独立生命周期  │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────────────────┘
                    │
         ┌─────────────────┐
         │  依赖组件        │
         │                 │
         │ • StatefulExecutor│
         │ • RouteInfoDriver│
         │ • ServiceStateCache│
         └─────────────────┘
```

## 使用方法

### 1. 创建配置

```go
import "github.com/go-kratos/kratos/v2/route"

// 创建基础配置
baseConfig := &route.BaseConfig{
    ServiceStateUpdatePeriodSeconds: 30,
    CpuFactor:                       0.7,
    TempLinkInfoCacheTimeSeconds:    300,
    UserLinkInfoCacheTimeSeconds:    240,
}

// 创建服务器信息
serverInfo := &route.ServerInfo{
    Namespace:   "default",
    ServiceName: "my-service",
    PodIndex:    0,
}
```

### 2. 创建驱动实例

```go
import "github.com/go-kratos/kratos/v2/route/driver"

// 创建客户端驱动
clientDriver := driver.NewStatefulRouteForClientDriverImpl(
    statefulExecutor,
    routeInfoDriver,
    serverInfo,
    logger,
)

// 创建服务端驱动
serverDriver := driver.NewStatefulRouteForServerDriverImpl(
    baseConfig,
    statefulExecutor,
    routeInfoDriver,
    serverInfo,
    logger,
)
```

### 3. 直接使用驱动

```go
// 启动驱动
ctx := context.Background()

// 初始化客户端驱动
if err := clientDriver.Init(); err != nil {
    log.Fatal(err)
}

// 启动服务端驱动
if err := serverDriver.Start(ctx); err != nil {
    log.Fatal(err)
}

// 使用客户端驱动
podIndex, err := clientDriver.ComputeLinkedPod(ctx, "default", "user123", "my-service")
if err != nil {
    log.Error(err)
}

// 使用服务端驱动
if err := serverDriver.SetLoadState(ctx, 150); err != nil {
    log.Error(err)
}

// 停止驱动
defer func() {
    if err := serverDriver.Stop(ctx); err != nil {
        log.Error(err)
    }
    if err := clientDriver.Close(); err != nil {
        log.Error(err)
    }
}()
```

## 配置说明

### BaseConfig
- `ServiceStateUpdatePeriodSeconds`: 服务状态更新周期（秒）
- `CpuFactor`: CPU因子，用于计算加权状态
- `TempLinkInfoCacheTimeSeconds`: 临时链接信息缓存时间（秒）
- `UserLinkInfoCacheTimeSeconds`: 用户链接信息缓存时间（秒）

### ServerInfo
- `Namespace`: 命名空间
- `ServiceName`: 服务名称
- `PodIndex`: Pod索引

## 注意事项

1. **命名空间隔离**: 所有操作都限制在本地命名空间内，不支持跨命名空间操作
2. **缓存策略**: 客户端驱动使用本地缓存，缓存时间比Redis临时缓存短60秒
3. **状态一致性**: 服务端驱动定时更新状态，确保状态一致性
4. **优雅关闭**: 支持优雅关闭，确保资源正确释放
5. **错误处理**: 所有操作都有完善的错误处理和日志记录

## 依赖关系

- **StatefulExecutor**: Redis操作和状态持久化
- **RouteInfoDriver**: 路由决策和状态查询
- **ServiceStateCache**: 状态缓存管理
- **go-cache**: 本地缓存库
- **Kratos**: 日志和依赖注入框架

## 扩展性

State模块设计为可扩展的架构：
- 支持自定义缓存策略
- 支持自定义负载均衡算法
- 支持自定义状态计算逻辑
- 支持插件式驱动扩展
