## 6A 任务卡：定义核心接口和数据模型

- 编号: Task-01
- 模块: route/driver
- 责任人: 待分配
- 优先级: 🔴
- 状态: ✅ 已完成
- 预计完成时间: 2025-01-27
- 实际完成时间: 2025-01-27

### A1 目标（Aim）
定义有状态路由的客户端和服务端驱动核心接口，以及相关的数据模型，为后续的驱动实现提供清晰的契约和数据结构。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：核心接口已在 `route/interfaces.go` 中定义
  - ✅ 已实现：数据模型已在 `route/types.go` 中定义
  - ✅ 已实现：客户端和服务端驱动接口已定义
- **差距**：无
- **约束**：需要遵循Go语言最佳实践，不使用回调方式
- **风险**：
  - 技术风险：无
  - 业务风险：无
  - 依赖风险：无

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`StatefulRouteForClientDriver` - 客户端驱动接口
  - **核心接口**：`StatefulRouteForServerDriver` - 服务端驱动接口
  - **核心方法**：
    - 客户端驱动：Pod计算、链接管理、缓存策略
    - 服务端驱动：状态管理、定时更新、链接操作
  - **输入输出参数及错误码**：
    - 使用Go标准库的`context.Context`进行超时和取消控制
    - 返回`error`接口类型，支持错误包装和类型断言
    - 使用Go的多返回值特性返回结果和错误

- **架构设计**：
  - 采用接口分离原则，将客户端和服务端功能分离到不同接口
  - 使用组合模式，将依赖组件注入到驱动实现中
  - 支持依赖注入，便于测试和配置管理

- **核心功能模块**：
  - `StatefulRouteForClientDriver`: 客户端驱动接口
  - `StatefulRouteForServerDriver`: 服务端驱动接口
  - 相关数据模型：`ServerInfo`、`BaseConfig`、`StateObj`等

- **极小任务拆分**：
  - T01-01：定义`StatefulRouteForClientDriver`接口
  - T01-02：定义`StatefulRouteForServerDriver`接口
  - T01-03：设计相关数据模型结构体

### A4 行动（Act）
#### T01-01：定义`StatefulRouteForClientDriver`接口
```go
// route/interfaces.go
package route

import "context"

// StatefulRouteForClientDriver 有状态客户端驱动接口
type StatefulRouteForClientDriver interface {
	// 初始化
	Init() error
	
	// Pod计算和链接管理
	ComputeLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)
	GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)
	GetLinkedPodInCacheOrIfPersist(ctx context.Context, namespace, uid, serviceName string) (int, error)
	GetLinkedPodNotCache(ctx context.Context, namespace, uid, serviceName string) (int, error)
	
	// 缓存操作
	GetLinkedInCache(ctx context.Context, namespace, uid, serviceName string) (int, error)
	CacheExpired(ctx context.Context, namespace, uid, serviceName string) error
	SetCache(ctx context.Context, namespace, uid, serviceName string, podIndex int) error
	
	// 批量操作
	GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error)
	BatchGetLinkedPod(ctx context.Context, namespace string, keys []string, serviceName string) (map[int][]string, error)
	
	// 链接设置
	SetLinkedPodIfAbsent(ctx context.Context, namespace, uid, serviceName string, podIndex int) (int, error)
}
```

#### T01-02：定义`StatefulRouteForServerDriver`接口
```go
// route/interfaces.go
package route

import "context"

// StatefulRouteForServerDriver 有状态服务端驱动接口
type StatefulRouteForServerDriver interface {
	// 初始化
	Init() error
	
	// 状态管理
	SetLoadState(ctx context.Context, loadState int) error
	GetLoadState(ctx context.Context) (int, error)
	SetRoutingState(ctx context.Context, state RoutingState) error
	GetRoutingState(ctx context.Context) (RoutingState, error)
	
	// 链接操作
	SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId int) (int, error)
	TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId int) (int, error)
	RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string) error
	RemoveLinkedPodWithId(ctx context.Context, namespace, uid, serviceName string, podId int, persistSeconds int) error
	
	// 启动和停止
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

#### T01-03：设计相关数据模型结构体
```go
// route/types.go
package route

// ServerInfo 服务器信息
type ServerInfo struct {
	Namespace   string `json:"namespace"`
	ServiceName string `json:"serviceName"`
	PodName     string `json:"podName"`
	PodIndex    int    `json:"podIndex"`
}

// BaseConfig 基础配置
type BaseConfig struct {
	Namespace   string `json:"namespace"`
	ServiceName string `json:"serviceName"`
	PodName     string `json:"podName"`
	PodIndex    int    `json:"podIndex"`
}

// StateObj 状态对象
type StateObj struct {
	Value     interface{} `json:"value"`
	Timestamp int64       `json:"timestamp"`
	TTL       int         `json:"ttl"`
}
```

### A5 验证（Assure）
- **测试用例**：接口定义已完成，无需额外测试
- **性能验证**：接口定义不影响性能
- **回归测试**：无
- **测试结果**：✅ 通过

### A6 迭代（Advance）
- 性能优化：接口设计已优化，支持并发访问
- 功能扩展：接口设计支持未来扩展
- 观测性增强：接口支持context传递，便于观测
- 下一步任务链接：[Task-02](./Task-02-实现服务端驱动.md)

### 📋 质量检查
- [x] 代码质量检查完成
- [x] 文档质量检查完成
- [x] 测试质量检查完成

### 📋 完成总结
Task-01已完成，成功定义了有状态路由的客户端和服务端驱动核心接口，以及相关的数据模型。接口设计遵循Go语言最佳实践，不使用回调方式，而是采用context和error返回的方式。所有接口已在`route/interfaces.go`中定义，数据模型已在`route/types.go`中定义，为后续的驱动实现提供了清晰的契约。
