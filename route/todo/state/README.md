# Route State 模块

## 概述

Route State 模块是 Homo-Go 项目中有状态服务路由系统的核心组件，实现**有状态路由的客户端和服务端驱动**。这些功能在Java版本中已经完整实现，包括Pod计算、链接管理、状态管理、跨命名空间通信等，需要转换为Go版本。

## 🎯 核心职责

### 1. 客户端驱动（StatefulRouteForClientDriver）
- **Pod计算**: 计算最佳可用Pod，支持缓存和Redis查询
- **链接管理**: 管理用户与Pod的链接关系
- **缓存策略**: 本地缓存优化，支持过期和更新
- **跨命名空间**: 支持gRPC跨命名空间通信

### 2. 服务端驱动（StatefulRouteForServerDriver）
- **状态管理**: 管理本地Pod的服务状态、负载状态、路由状态
- **定时更新**: 定时更新服务状态和工作负载状态
- **链接操作**: 支持Pod链接的设置、尝试设置、移除等操作
- **配置验证**: 验证基础配置的有效性

### 3. gRPC客户端（RouteGrpcClient）
- **跨命名空间通信**: 支持跨Kubernetes命名空间的Pod链接操作
- **连接管理**: 管理gRPC连接的生命周期
- **健康检查**: 监控远程服务的健康状态

## 🏗️ 架构设计

### 模块关系
```
route/
├── state/                    # 本模块：有状态路由驱动实现
│   ├── client_driver.go        # 客户端驱动实现
│   ├── server_driver.go        # 服务端驱动实现
│   ├── state_manager.go        # 状态管理器
│   └── grpc_client.go          # gRPC客户端
├── driver/                  # 路由驱动（已完成）
├── executor/                # Redis执行器（已完成）
└── cache/                   # 缓存组件（已完成）
```

### 核心接口
- **StatefulRouteForClientDriver**: 客户端驱动主接口
- **StatefulRouteForServerDriver**: 服务端驱动主接口
- **RouteGrpcClient**: gRPC客户端接口
- **StatefulObjCallback**: 异步操作回调接口

### 数据模型
- **缓存管理**: 本地Pod索引缓存
- **状态对象**: 服务状态、负载状态、路由状态
- **链接信息**: 用户与Pod的链接关系

## 🔄 与现有模块的集成

### 依赖模块
- **RouteInfoDriver** (driver模块): 用于路由决策和状态查询
- **StatefulExecutor** (executor模块): 用于Redis操作和状态持久化
- **ServiceStateCache** (cache模块): 用于状态缓存

### 新增模块
- **StatefulRouteForClientDriver**: 客户端驱动接口和实现
- **StatefulRouteForServerDriver**: 服务端驱动接口和实现
- **RouteGrpcClient**: gRPC客户端实现

## 📋 开发任务

| 任务 | 状态 | 描述 |
|------|------|------|
| Task-01 | ❌ 未开始 | 定义客户端和服务端驱动接口 |
| Task-02 | ❌ 未开始 | 实现客户端驱动（StatefulRouteForClientDriverImpl） |
| Task-03 | ❌ 未开始 | 实现服务端驱动（StatefulRouteForServerDriverImpl） |
| Task-04 | ❌ 未开始 | 实现gRPC客户端（跨命名空间通信） |
| Task-05 | ❌ 未开始 | 编写单元测试 |
| Task-06 | ❌ 未开始 | 性能优化和观测性增强 |

## 🚀 使用场景

### 1. 客户端Pod计算
```go
// 计算最佳可用Pod
podIndex, err := clientDriver.ComputeLinkedPod(ctx, "default", "user123", "user-service")
```

### 2. 服务端状态管理
```go
// 设置本地Pod负载状态
err := serverDriver.SetLoadState(ctx, 2) // 低负载状态

// 设置路由状态
err := serverDriver.SetRoutingState(ctx, route.RoutingStateReady)
```

### 3. 跨命名空间通信
```go
// 通过gRPC设置跨命名空间的Pod链接
podIndex, err := grpcClient.SetLinkedPodIfAbsent(ctx, "prod", "user123", "user-service", 1)
```

## 🛠️ 技术特性

- **Go 1.21+**: 使用最新的Go语言特性
- **Kratos v2**: 基于Kratos框架构建
- **本地缓存**: 客户端本地缓存优化性能
- **gRPC通信**: 支持跨命名空间通信
- **异步操作**: 基于回调的异步操作模式
- **定时任务**: 服务端定时状态更新

## 📚 设计原则

1. **功能对等**: 与Java版本功能完全一致
2. **接口分离**: 客户端和服务端功能分离
3. **依赖注入**: 支持Kratos的依赖注入系统
4. **缓存优化**: 本地缓存减少网络开销
5. **异步支持**: 支持异步操作和回调机制

## 🔍 关键特性

### 缓存策略
- **客户端缓存**: 本地Pod索引缓存，减少Redis查询
- **缓存过期**: 支持缓存过期和自动更新
- **失效处理**: 智能处理缓存失效情况

### 状态管理
- **定时更新**: 服务端定时更新服务状态和工作负载状态
- **状态一致性**: 确保状态的一致性和有效性
- **配置验证**: 验证基础配置的有效性

### 跨命名空间
- **gRPC通信**: 支持跨Kubernetes命名空间的Pod链接操作
- **连接管理**: 管理gRPC连接的生命周期
- **健康检查**: 监控远程服务的健康状态

## 🔍 监控和观测

- **性能指标**: Pod计算和链接操作的性能指标
- **缓存命中率**: 本地缓存的命中率统计
- **跨命名空间通信**: gRPC通信的性能和错误统计
- **状态更新**: 服务端状态更新的频率和成功率

---

**版本**: v1.0.0  
**最后更新**: 2025-01-27  
**维护者**: AI助手
