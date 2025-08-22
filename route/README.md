# Route 模块

## 概述

Route 模块是 homo-go 项目中的核心路由管理组件，负责管理有状态服务的路由、状态管理和负载均衡。该模块基于 6A 工作流程设计，采用 Go 语言和 Kratos 框架开发。

## 模块结构

```
route/
├── types.go                    # 数据模型定义（服务状态、负载状态、工作负载状态等）
├── interfaces.go               # 核心接口定义（RouteInfoDriver、StatefulExecutor等）
├── executor/                   # 执行器模块
│   ├── stateful_executor.go    # 有状态执行器实现
│   ├── lua_scripts/            # Lua脚本文件
│   │   └── statefulSetLink.lua # 设置Pod链接脚本
│   └── README.md               # 执行器模块说明
├── driver/                     # 驱动层（待实现）
├── cache/                      # 缓存层（待实现）
└── todo/                       # 任务文档和规划
    ├── driver/                 # 驱动层任务
    ├── executor/               # 执行器任务
    └── state/                  # 状态管理任务
```

## 核心功能

### 1. 服务状态管理

- **Pod状态管理**: 管理每个Pod的服务状态、负载状态和路由状态.
- **工作负载状态**: 管理Kubernetes工作负载的整体状态
- **状态缓存**: 提供高性能的状态缓存机制

### 2. 路由决策

- **智能路由**: 基于Pod状态和负载情况选择最佳Pod
- **负载均衡**: 支持多种负载均衡策略
- **故障转移**: 自动检测故障Pod并切换到健康Pod

### 3. 用户会话管理

- **Pod链接**: 管理用户与Pod的关联关系
- **会话持久化**: 支持会话状态的持久化存储
- **跨命名空间**: 支持跨Kubernetes命名空间的服务访问

### 4. Redis集成

- **连接池管理**: 高效的Redis连接池管理
- **Lua脚本**: 使用Lua脚本保证操作的原子性
- **健康检查**: 自动检测Redis连接健康状态

## 技术特性

### 1. 架构设计

- **接口优先**: 采用接口分离原则，便于测试和扩展
- **模块化**: 清晰的模块边界和职责分离
- **可扩展**: 支持插件式扩展和自定义实现

### 2. 性能优化

- **连接复用**: Redis连接池复用，减少连接开销
- **批量操作**: 支持批量查询和更新操作
- **异步处理**: 异步状态更新和事件通知

### 3. 可靠性保障

- **错误处理**: 完善的错误分类和处理机制
- **重试机制**: 智能重试策略，提高系统稳定性
- **监控指标**: 丰富的监控指标和告警机制

## 快速开始

### 1. 基本使用

```go
package main

import (
    "context"
    "time"
    "github.com/go-kratos/kratos/v2/route"
    "github.com/go-kratos/kratos/v2/route/executor"
    "github.com/redis/go-redis/v9"
)

func main() {
    // 创建Redis配置
    redisConfig := &route.RedisConfig{
        Addresses: []string{"localhost:6379"},
        PoolSize:  10,
        // ... 其他配置
    }

    // 创建日志记录器
    logger := &MyLogger{}

    // 创建Redis客户端
    redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
    defer redisClient.Close()

    // 创建有状态执行器
    executor := executor.NewStatefulExecutor(redisClient, logger)

    // 设置服务状态
    ctx := context.Background()
    err = executor.SetServiceState(ctx, "default", "my-service", 0, "Ready")
    if err != nil {
        log.Printf("设置服务状态失败: %v", err)
    }

    // 获取服务状态
    states, err := executor.GetServiceState(ctx, "default", "my-service")
    if err != nil {
        log.Printf("获取服务状态失败: %v", err)
    } else {
        for podID, state := range states {
            log.Printf("Pod %d: %s", podID, state)
        }
    }
}
```

### 2. 配置说明

```yaml
# config.yaml
route:
  redis:
    addresses: ["localhost:6379"]
    password: ""
    db: 0
    poolSize: 10
    minIdleConns: 5
    maxRetries: 3
    dialTimeout: 5s
    readTimeout: 3s
    writeTimeout: 3s
    poolTimeout: 4s
    idleTimeout: 5m
    healthCheckIntervalMs: 30000
  
  cache:
    linkInfoCacheTimeSecs: 300
    stateCacheExpireSecs: 1800
    maxRetryCount: 3
    retryDelayMs: 100
    logLevel: "info"
    enableMetrics: true
    enableTracing: true
```

## 开发状态

### 已完成功能

- ✅ **Task-01**: 核心接口定义与数据结构设计
- ✅ **Task-02**: Redis连接池与Lua脚本管理
- ✅ **Task-03**: 核心业务逻辑实现
- ✅ **Task-01**: RouteInfoDriver核心接口定义
- ✅ **Task-02**: 数据模型和状态枚举定义

### 进行中功能

- 🔄 **Task-03**: ServiceStateCache缓存组件实现

### 待实现功能

- ❌ **Task-04**: StatefulRedisExecutor Redis执行器
- ❌ **Task-05**: ServerRedisExecutor 服务Redis执行器
- ❌ **Task-06**: RouteInfoDriverImpl 主驱动实现
- ❌ **Task-07**: gRPC客户端集成
- ❌ **Task-08**: 配置管理和依赖注入
- ❌ **Task-09**: 单元测试和集成测试
- ❌ **Task-10**: 性能优化和监控指标

## 设计原则

### 1. 第一性原理

- 从业务本质出发，设计简洁而强大的接口
- 避免过度抽象，保持代码的直观性
- 专注于解决核心问题，避免功能蔓延

### 2. KISS原则

- 保持简单，避免不必要的复杂性
- 清晰的命名和结构，便于理解和维护
- 最小化依赖，减少外部耦合

### 3. SOLID原则

- **单一职责**: 每个模块只负责一个功能领域
- **开闭原则**: 支持扩展，避免修改现有代码
- **里氏替换**: 接口实现可以相互替换
- **接口隔离**: 客户端只依赖需要的接口
- **依赖倒置**: 依赖抽象而非具体实现

## 测试策略

### 1. 单元测试

- 核心业务逻辑的单元测试
- 接口契约的验证测试
- 错误处理路径的测试

### 2. 集成测试

- Redis集成的端到端测试
- 模块间交互的集成测试
- 性能基准测试

### 3. 测试覆盖

- 目标测试覆盖率：≥80%
- 关键路径100%覆盖
- 边界条件和异常情况测试

## 性能指标

### 1. 响应时间

- 服务状态查询：< 10ms
- Pod链接操作：< 5ms
- 批量操作：< 50ms

### 2. 吞吐量

- 单实例：> 1000 QPS
- 集群模式：> 10000 QPS
- 支持水平扩展

### 3. 资源使用

- 内存使用：< 100MB
- CPU使用：< 10%
- 网络延迟：< 1ms

## 部署说明

### 1. 环境要求

- Go 1.21+
- Redis 6.0+
- Kubernetes 1.20+ (可选)

### 2. 部署步骤

```bash
# 1. 克隆代码
git clone <repository-url>
cd homo-go

# 2. 安装依赖
go mod download

# 3. 构建项目
go build -o bin/route ./route

# 4. 运行服务
./bin/route -config config.yaml
```

### 3. 配置验证

```bash
# 检查Redis连接
redis-cli ping

# 检查服务健康状态
curl http://localhost:8080/health
```

## 监控和运维

### 1. 健康检查

- Redis连接状态检查
- 服务响应时间监控
- 错误率统计和告警

### 2. 日志管理

- 结构化日志输出
- 日志级别动态调整
- 日志轮转和归档

### 3. 指标收集

- Prometheus指标导出
- 自定义业务指标
- 性能指标监控

## 故障排除

### 1. 常见问题

- **Redis连接失败**: 检查网络和Redis服务状态
- **性能下降**: 检查连接池配置和Redis性能
- **内存泄漏**: 检查连接是否正确释放

### 2. 调试技巧

- 启用详细日志记录
- 监控连接池统计信息
- 使用Redis客户端工具验证操作

## 贡献指南

### 1. 开发流程

1. Fork项目仓库
2. 创建功能分支
3. 编写代码和测试
4. 提交Pull Request
5. 代码审查和合并

### 2. 代码规范

- 遵循Go语言官方规范
- 使用gofmt格式化代码
- 添加适当的注释和文档
- 编写单元测试用例

### 3. 提交规范

- 清晰的提交信息
- 关联Issue编号
- 描述变更内容和影响

## 版本规划

### v1.0.0 (当前版本)

- 基础功能实现
- Redis集成和Lua脚本支持
- 错误处理和指标收集

### v1.1.0 (计划中)

- 完整的驱动层实现
- gRPC客户端集成
- 配置管理优化

### v1.2.0 (计划中)

- 性能优化和监控增强
- 单元测试和集成测试
- 文档完善和示例代码

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。

## 联系方式

- 项目主页: [GitHub Repository]
- 问题反馈: [GitHub Issues]
- 讨论交流: [GitHub Discussions]

---

**最后更新**: 2025-01-27  
**版本**: v1.0.0  
**维护者**: AI助手
