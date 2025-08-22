## 6A 任务卡：核心接口定义与数据结构设计

- 编号: Task-01
- 模块: route/executor
- 责任人: AI助手
- 优先级: 🔴 高
- 状态: ✅ 已完成
- 预计完成时间: 2025-01-27
- 实际完成时间: 2025-01-27

### A1 目标（Aim）
定义有状态服务执行器的核心接口和数据结构，为有状态服务提供统一的状态管理、Pod链接管理和工作负载状态管理接口，支持Go语言的context和error模式。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Java版本的StatefulExecutor接口定义完整
  - ✅ 已完成：Go版本的接口定义已在 `route/interfaces.go` 中完成
  - ✅ 已完成：所有数据模型已在 `route/types.go` 中定义
  - ✅ 已完成：Promise模式已转换为Go的context和error模式
- **差距**：
  - 无，接口定义已完成
- **约束**：
  - 已保持与Java版本接口的兼容性
  - 已符合Go语言的接口设计规范
  - 已支持Kratos框架的依赖注入
- **风险**：
  - 无，接口已实现

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`StatefulExecutor` - 定义所有有状态操作的核心接口
  - **核心方法**：
    - `SetServiceState(ctx context.Context, namespace, serviceName string, podId int, state string) error`
    - `GetServiceState(ctx context.Context, namespace, serviceName string) (map[int]string, error)`
    - `SetWorkloadState(ctx context.Context, namespace, serviceName, state string) error`
    - `GetWorkloadState(ctx context.Context, namespace, serviceNames []string) (map[string]string, error)`
    - `SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId, persistSeconds int) (int, error)`
    - `TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId, persistSeconds int) (bool, int, error)`
    - `GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)`
    - `RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string, persistSeconds int) (bool, error)`
    - `GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error)`

- **架构设计**：
  - 采用接口分离原则，将不同功能模块分离到不同接口
  - 使用Go的context包进行超时和取消控制
  - 使用Go的多返回值特性返回结果和错误
  - 支持依赖注入，便于测试和配置管理
  - **单文件架构**: 所有实现整合到 `stateful_executor.go` 中

- **核心功能模块**：
  - `StatefulExecutor`: 主执行器接口（已在 `route/interfaces.go` 中定义）
  - `StatefulExecutorImpl`: 主执行器实现（已在 `route/executor/stateful_executor.go` 中实现）
  - `ServiceLinkInfo`: 服务链接信息结构体

- **极小任务拆分**：
  - ✅ T01-01：定义`StatefulExecutor`核心接口（已完成）
  - ✅ T01-02：定义`ServiceLinkInfo`数据结构体（已完成）
  - ✅ T01-03：实现`StatefulExecutorImpl`主实现（已完成）

### A4 行动（Act）
#### ✅ T01-01：定义`StatefulExecutor`核心接口（已完成）
```go
// route/interfaces.go
package route

import "context"

// StatefulExecutor 定义所有有状态操作的核心接口
type StatefulExecutor interface {
    // 服务状态管理
    SetServiceState(ctx context.Context, namespace, serviceName string, podId int, state string) error
    GetServiceState(ctx context.Context, namespace, serviceName string) (map[int]string, error)
    
    // 工作负载状态管理
    SetWorkloadState(ctx context.Context, namespace, serviceName, state string) error
    GetWorkloadState(ctx context.Context, namespace, serviceNames []string) (map[string]string, error)
    
    // Pod链接管理
    SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId, persistSeconds int) (int, error)
    TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId, persistSeconds int) (bool, int, error)
    SetLinkedPodIfAbsent(ctx context.Context, namespace, uid, serviceName string, podId, persistSeconds int) (int, error)
    GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)
    GetLinkedPodIfPersist(ctx context.Context, namespace, uid, serviceName string) (int, error)
    BatchGetLinkedPod(ctx context.Context, namespace string, keys []string, serviceName string) (map[int][]string, error)
    RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string, persistSeconds int) (bool, error)
    RemoveLinkedPodWithId(ctx context.Context, namespace, uid, serviceName string, persistSeconds, podId int) (bool, error)
    GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error)
}
```

#### ✅ T01-02：定义`ServiceLinkInfo`数据结构体（已完成）
```go
// route/executor/stateful_executor.go
package executor

// ServiceLinkInfo 服务链接信息
type ServiceLinkInfo struct {
    UID         string `json:"uid"`         // 用户ID
    Namespace   string `json:"namespace"`   // 命名空间
    ServiceName string `json:"serviceName"` // 服务名称
    PodID       int    `json:"podID"`       // Pod ID
    UpdateTime  int64  `json:"updateTime"`  // 更新时间
}
```

#### ✅ T01-03：实现`StatefulExecutorImpl`主实现（已完成）
```go
// route/executor/stateful_executor.go
package executor

// StatefulExecutorImpl 有状态执行器实现
type StatefulExecutorImpl struct {
    redisClient *redis.Client
    logger      log.Logger
    scriptCache map[string]string
}

// NewStatefulExecutor 创建新的有状态执行器
func NewStatefulExecutor(redisClient *redis.Client, logger log.Logger) *StatefulExecutorImpl {
    executor := &StatefulExecutorImpl{
        redisClient: redisClient,
        logger:      logger,
        scriptCache: make(map[string]string),
    }
    
    // 预加载所有Lua脚本
    if err := executor.preloadScripts(context.Background()); err != nil {
        logger.Log(log.LevelError, "msg", "Failed to preload Lua scripts", "error", err)
    }
    
    return executor
}
```

### A5 验证（Assure）
- **测试用例**：
  - ✅ 接口定义完整性测试（已完成）
  - ✅ 方法签名正确性测试（已完成）
  - ✅ 数据结构定义测试（已完成）
- **性能验证**：
  - ✅ 接口调用性能基准测试（已完成）
- **回归测试**：
  - ✅ 与Java版本接口兼容性验证（已完成）
- **测试结果**：
  - ✅ 所有测试通过

### A6 迭代（Advance）
- 性能优化：接口方法已优化，减少不必要的参数传递
- 功能扩展：支持更多状态类型和路由策略
- 观测性增强：已添加接口调用监控和追踪
- 下一步任务链接：Task-02 Redis连接池与Lua脚本管理（已完成）

### 📋 质量检查
- [x] 代码质量检查完成
- [x] 文档质量检查完成
- [x] 测试质量检查完成

### 📋 完成总结
Task-01已成功完成，StatefulExecutor核心接口已在 `route/interfaces.go` 中定义完成，包括：
1. 完整的接口方法定义，支持Go的context和error模式
2. 与Java版本接口的完全兼容性
3. 符合Go语言接口设计规范
4. 支持Kratos框架的依赖注入
5. 所有相关数据模型已在 `route/executor/stateful_executor.go` 中定义
6. 采用单文件架构，减少模块复杂度

下一步将进行Task-02的实现。
