## 6A 任务卡：定义核心接口和数据模型

- 编号: Task-01
- 模块: route/state
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: 待定
- 实际完成时间: -

### A1 目标（Aim）
定义有状态路由模块的核心接口和数据模型，为后续实现提供清晰的契约定义。包括服务端驱动接口、客户端驱动接口、数据模型结构体等。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Java版本的完整接口定义和实现
  - 🔄 部分实现：需要分析Java接口转换为Go接口
  - ❌ 未实现：Go版本的接口和数据模型
- **差距**：
  - Java接口需要转换为Go idiomatic风格
  - 数据模型需要适配Go的struct语法
  - 错误处理需要符合Go的error接口规范
- **约束**：
  - 必须保持与Java版本功能逻辑一致
  - 符合Go语言和Kratos框架设计规范
  - 支持现有的routeInfoDriver和StatefulRedisExecutor
- **风险**：
  - 技术风险：接口设计可能不符合Go最佳实践
  - 业务风险：功能逻辑转换可能遗漏
  - 依赖风险：依赖现有组件的接口变更

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：
    - `StatefulRouteForServerDriver` - 服务端驱动接口
    - `StatefulRouteForClientDriver` - 客户端驱动接口
  - **核心方法**：
    - 服务端：负载状态管理、路由状态管理、链接管理
    - 客户端：pod计算、链接获取、缓存管理
  - **输入输出参数及错误码**：
    - 使用Go标准库的`context.Context`进行超时和取消控制
    - 返回`error`接口类型，支持错误包装和类型断言
    - 使用Go的多返回值特性返回结果和错误

- **架构设计**：
  - 采用接口分离原则，将服务端和客户端功能分离
  - 使用组合模式，支持依赖注入和测试
  - 遵循Go的接口设计最佳实践

- **核心功能模块**：
  - `StatefulRouteForServerDriver`: 服务端驱动接口
  - `StatefulRouteForClientDriver`: 客户端驱动接口
  - `ServiceState`: 服务状态数据模型
  - `RoutingState`: 路由状态枚举

- **极小任务拆分**：
  - T01-01：定义`ServiceState`数据模型
  - T01-02：定义`RoutingState`枚举类型
  - T01-03：定义`StatefulRouteForServerDriver`接口
  - T01-04：定义`StatefulRouteForClientDriver`接口

### A4 行动（Act）
#### T01-01：定义`ServiceState`数据模型
```go
// route/types.go
package route

import "time"

// ServiceState 服务状态数据模型
type ServiceState struct {
    State      int       `json:"state"`       // 服务状态值
    UpdateTime time.Time `json:"update_time"` // 更新时间
}

// NewServiceState 创建新的服务状态
func NewServiceState(state int) *ServiceState {
    return &ServiceState{
        State:      state,
        UpdateTime: time.Now(),
    }
}
```

#### T01-02：定义`RoutingState`枚举类型
```go
// route/types.go
package route

// RoutingState 路由状态枚举
type RoutingState int

const (
    RoutingStateReady        RoutingState = iota // 就绪状态
    RoutingStateDenyNewCall                      // 拒绝新调用
    RoutingStateShutdown                         // 关闭状态
)

// String 返回路由状态的字符串表示
func (r RoutingState) String() string {
    switch r {
    case RoutingStateReady:
        return "READY"
    case RoutingStateDenyNewCall:
        return "DENY_NEW_CALL"
    case RoutingStateShutdown:
        return "SHUTDOWN"
    default:
        return "UNKNOWN"
    }
}
```

#### T01-03：定义`StatefulRouteForServerDriver`接口
```go
// route/interfaces.go
package route

import (
    "context"
)

// StatefulRouteForServerDriver 有状态服务端驱动接口
type StatefulRouteForServerDriver interface {
    // 设置服务器负载状态
    SetLoadState(ctx context.Context, loadState int) error
    
    // 获取服务器负载状态
    GetLoadState(ctx context.Context) (int, error)
    
    // 设置路由状态
    SetRoutingState(ctx context.Context, state RoutingState) error
    
    // 获取路由状态
    GetRoutingState(ctx context.Context) (RoutingState, error)
    
    // 设置连接信息
    SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId int) (int, error)
    
    // 尝试设置指定uid连接的podId
    TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId int) (bool, int, error)
    
    // 移除服务连接信息
    RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string) error
    
    // 移除指定uid连接的podId
    RemoveLinkedPodWithId(ctx context.Context, namespace, uid, serviceName string, podId int, persistSeconds int) error
}
```

#### T01-04：定义`StatefulRouteForClientDriver`接口
```go
// route/interfaces.go
package route

import (
    "context"
)

// StatefulRouteForClientDriver 有状态客户端驱动接口
type StatefulRouteForClientDriver interface {
    // 计算一个最佳podId
    ComputeLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)
    
    // 设置链接pod（如果不存在）
    SetLinkedPodIfAbsent(ctx context.Context, namespace, uid, serviceName string, podIndex int) (int, error)
    
    // 获取uid下的连接信息
    GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)
    
    // 从缓存或注册中心获取连接信息
    GetLinkedPodInCacheOrIfPersist(ctx context.Context, namespace, uid, serviceName string) (int, error)
    
    // 直接从注册中心获取连接信息
    GetLinkedPodNotCache(ctx context.Context, namespace, uid, serviceName string) (int, error)
    
    // 获取缓存中的连接信息
    GetLinkedInCache(ctx context.Context, namespace, uid, serviceName string) (int, error)
    
    // 标识连接信息过期
    CacheExpired(ctx context.Context, namespace, uid, serviceName string) (int, error)
    
    // 设置连接信息到缓存
    SetCache(ctx context.Context, namespace, uid, serviceName string, podIndex int) (int, error)
    
    // 获取用户连接的所有服务
    GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error)
    
    // 批量获取连接信息
    BatchGetLinkedPod(ctx context.Context, namespace string, keys []string, serviceName string) (map[int][]string, error)
}
```

### A5 验证（Assure）
- **测试用例**：
  - 数据模型序列化/反序列化测试
  - 接口方法签名验证
  - 枚举类型转换测试
- **性能验证**：
  - 接口调用性能基准测试
- **回归测试**：
  - 确保与Java版本功能一致
- **测试结果**：
  - 待实现

### A6 迭代（Advance）
- 性能优化：接口设计优化
- 功能扩展：支持更多状态类型
- 观测性增强：添加接口调用监控
- 下一步任务链接：Task-02 实现服务端驱动

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
[总结任务完成情况]
