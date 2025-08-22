## 6A 任务卡：实现服务端驱动（StatefulRouteForServerDriver）

- 编号: Task-02
- 模块: route/state
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: 待定
- 实际完成时间: -

### A1 目标（Aim）
实现有状态路由模块的服务端驱动，负责管理服务负载状态、路由状态、用户链接关系等核心功能，确保服务状态的一致性和可靠性。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Java版本的完整服务端驱动实现
  - 🔄 部分实现：接口已定义，需要实现具体逻辑
  - ❌ 未实现：Go版本的服务端驱动实现
- **差距**：
  - Java实现逻辑需要转换为Go实现
  - 定时器、状态管理、错误处理需要适配Go生态
  - 依赖注入和生命周期管理需要符合Kratos规范
- **约束**：
  - 必须保持与Java版本功能逻辑一致
  - 使用现有的routeInfoDriver和StatefulRedisExecutor
  - 符合Go语言和Kratos框架设计规范
- **风险**：
  - 技术风险：Go定时器实现可能与Java不同
  - 业务风险：状态管理逻辑转换可能遗漏
  - 依赖风险：依赖组件接口变更影响

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`StatefulRouteForServerDriver` - 服务端驱动接口
  - **核心方法**：
    - 负载状态管理：`SetLoadState`、`GetLoadState`
    - 路由状态管理：`SetRoutingState`、`GetRoutingState`
    - 链接管理：`SetLinkedPod`、`TrySetLinkedPod`、`RemoveLinkedPod`
  - **输入输出参数及错误码**：
    - 使用`context.Context`进行超时和取消控制
    - 返回`error`接口类型，支持错误包装
    - 使用Go的多返回值特性

- **架构设计**：
  - 采用组合模式，集成现有组件
  - 使用Go的定时器机制管理状态更新
  - 支持优雅关闭和状态清理

- **核心功能模块**：
  - `StatefulRouteForServerDriverImpl`: 服务端驱动实现
  - 状态管理：负载状态、路由状态、服务状态
  - 定时任务：状态更新、健康检查
  - 链接管理：用户与pod的关联关系

- **极小任务拆分**：
  - T02-01：实现基础结构和依赖注入
  - T02-02：实现负载状态和路由状态管理
  - T02-03：实现链接管理功能
  - T02-04：实现定时任务和状态更新
  - T02-05：实现优雅关闭和资源清理

### A4 行动（Act）
#### T02-01：实现基础结构和依赖注入
```go
// route/driver/server_driver.go
package driver

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    "sync"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "github.com/go-kratos/kratos/v2/transport"
    "github.com/redis/go-redis/v9"
    
    "tpf-service-go/stateful-route/interfaces"
    "tpf-service-go/stateful-route/types"
)

// StatefulRouteForServerDriverImpl 有状态服务端驱动实现
type StatefulRouteForServerDriverImpl struct {
    mu sync.RWMutex
    
    // 配置信息
    baseConfig     *BaseConfig
    serverInfo     *ServerInfo
    
    // 依赖组件
    statefulRedisExecutor interfaces.StatefulRedisExecutor
    routeInfoDriver       interfaces.RouteInfoDriver
    redisPool            *redis.Client
    
    // 状态信息
    loadState     int
    routingState  types.RoutingState
    podIndex     int
    serviceName   string
    namespace    string
    
    // 定时器
    stateTimer     *time.Ticker
    workloadTimer  *time.Ticker
    
    // 控制
    ctx    context.Context
    cancel context.CancelFunc
    logger log.Logger
}

// NewStatefulRouteForServerDriverImpl 创建新的服务端驱动实例
func NewStatefulRouteForServerDriverImpl(
    baseConfig *BaseConfig,
    statefulRedisExecutor interfaces.StatefulRedisExecutor,
    routeInfoDriver interfaces.RouteInfoDriver,
    redisPool *redis.Client,
    logger log.Logger,
) *StatefulRouteForServerDriverImpl {
    return &StatefulRouteForServerDriverImpl{
        baseConfig:             baseConfig,
        statefulRedisExecutor:  statefulRedisExecutor,
        routeInfoDriver:        routeInfoDriver,
        redisPool:              redisPool,
        logger:                 logger,
        routingState:           types.RoutingStateReady,
    }
}
```

#### T02-02：实现负载状态和路由状态管理
```go
// route/driver/server_driver.go
package driver

// SetLoadState 设置服务器负载状态
func (s *StatefulRouteForServerDriverImpl) SetLoadState(ctx context.Context, loadState int) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.loadState = loadState
    s.logger.Infof("Set load state: %d", loadState)
    return nil
}

// GetLoadState 获取服务器负载状态
func (s *StatefulRouteForServerDriverImpl) GetLoadState(ctx context.Context) (int, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    return s.loadState, nil
}

// SetRoutingState 设置路由状态
func (s *StatefulRouteForServerDriverImpl) SetRoutingState(ctx context.Context, state types.RoutingState) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.routingState = state
    s.logger.Infof("Set routing state: %s", state.String())
    return nil
}

// GetRoutingState 获取路由状态
func (s *StatefulRouteForServerDriverImpl) GetRoutingState(ctx context.Context) (types.RoutingState, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    return s.routingState, nil
}
```

#### T02-03：实现链接管理功能
```go
// route/driver/server_driver.go
package driver

// SetLinkedPod 设置连接信息
func (s *StatefulRouteForServerDriverImpl) SetLinkedPod(
    ctx context.Context, 
    namespace, uid, serviceName string, 
    podId int,
) (int, error) {
    // 验证命名空间
    if !s.isLocalNamespace(namespace) {
        return 0, fmt.Errorf("namespace %s is not local", namespace)
    }
    
    // 调用Redis执行器设置链接
    previousPodId, err := s.statefulRedisExecutor.SetLinkedPod(
        ctx, namespace, uid, serviceName, podId, -1, // PERSIST_FOREVER
    )
    if err != nil {
        s.logger.Errorf("Failed to set linked pod: %v", err)
        return 0, err
    }
    
    s.logger.Infof("Set linked pod: namespace=%s, uid=%s, service=%s, podId=%d, previous=%d",
        namespace, uid, serviceName, podId, previousPodId)
    
    return previousPodId, nil
}

// TrySetLinkedPod 尝试设置指定uid连接的podId
func (s *StatefulRouteForServerDriverImpl) TrySetLinkedPod(
    ctx context.Context, 
    namespace, uid, serviceName string, 
    podId int,
) (bool, int, error) {
    // 验证命名空间
    if !s.isLocalNamespace(namespace) {
        return false, 0, fmt.Errorf("namespace %s is not local", namespace)
    }
    
    // 调用Redis执行器尝试设置链接
    success, previousPodId, err := s.statefulRedisExecutor.TrySetLinkedPod(
        ctx, namespace, uid, serviceName, podId, -1, // PERSIST_FOREVER
    )
    if err != nil {
        s.logger.Errorf("Failed to try set linked pod: %v", err)
        return false, 0, err
    }
    
    s.logger.Infof("Try set linked pod: namespace=%s, uid=%s, service=%s, podId=%d, success=%v, previous=%d",
        namespace, uid, serviceName, podId, success, previousPodId)
    
    return success, previousPodId, nil
}
```

#### T02-04：实现定时任务和状态更新
```go
// route/driver/server_driver.go
package driver

// Start 启动服务端驱动
func (s *StatefulRouteForServerDriverImpl) Start(ctx context.Context) error {
    s.ctx, s.cancel = context.WithCancel(ctx)
    
    // 初始化pod信息
    if err := s.initializePodInfo(); err != nil {
        return fmt.Errorf("failed to initialize pod info: %w", err)
    }
    
    // 启动状态更新定时器
    s.startStateUpdateTimer()
    
    // 如果是主pod，启动工作负载状态更新
    if s.podIndex == 0 {
        s.startWorkloadStateUpdateTimer()
    }
    
    s.logger.Infof("Stateful route server driver started: service=%s, pod=%d", s.serviceName, s.podIndex)
    return nil
}

// startStateUpdateTimer 启动状态更新定时器
func (s *StatefulRouteForServerDriverImpl) startStateUpdateTimer() {
    interval := time.Duration(s.baseConfig.ServiceStateUpdatePeriodSeconds) * time.Second
    s.stateTimer = time.NewTicker(interval)
    
    go func() {
        for {
            select {
            case <-s.ctx.Done():
                return
            case <-s.stateTimer.C:
                s.updateServiceState()
            }
        }
    }()
}

// updateServiceState 更新服务状态
func (s *StatefulRouteForServerDriverImpl) updateServiceState() {
    ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
    defer cancel()
    
    // 计算当前状态
    currentState := s.calculateCurrentState()
    
    // 设置服务状态
    err := s.statefulRedisExecutor.SetServiceState(
        ctx, s.namespace, s.serviceName, s.podIndex, currentState,
    )
    if err != nil {
        s.logger.Errorf("Failed to update service state: %v", err)
        return
    }
    
    s.logger.Debugf("Updated service state: %s", currentState)
}
```

#### T02-05：实现优雅关闭和资源清理
```go
// route/driver/server_driver.go
package driver

// Stop 停止服务端驱动
func (s *StatefulRouteForServerDriverImpl) Stop(ctx context.Context) error {
    s.logger.Info("Stopping stateful route server driver...")
    
    // 设置拒绝新调用状态
    s.SetRoutingState(ctx, types.RoutingStateDenyNewCall)
    
    // 等待一段时间确保状态传播
    waitTime := time.Duration(s.baseConfig.ServiceStateUpdatePeriodSeconds*5) * time.Second
    select {
    case <-time.After(waitTime):
    case <-ctx.Done():
        return ctx.Err()
    }
    
    // 取消上下文
    if s.cancel != nil {
        s.cancel()
    }
    
    // 停止定时器
    if s.stateTimer != nil {
        s.stateTimer.Stop()
    }
    if s.workloadTimer != nil {
        s.workloadTimer.Stop()
    }
    
    s.logger.Info("Stateful route server driver stopped")
    return nil
}

// calculateCurrentState 计算当前状态
func (s *StatefulRouteForServerDriverImpl) calculateCurrentState() string {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    // 计算加权状态
    cpuFactor := s.baseConfig.CpuFactor
    weightedState := int(cpuFactor*float64(s.getWaitingTasksNum()) + (1-cpuFactor)*float64(s.loadState))
    
    // 创建状态对象
    stateObj := &types.StateObj{
        State:      weightedState,
        UpdateTime: time.Now(),
        RoutingState: s.routingState,
    }
    
    return stateObj.ToStateString()
}
```

### A5 验证（Assure）
- **测试用例**：
  - 状态管理功能测试
  - 链接管理功能测试
  - 定时任务功能测试
  - 优雅关闭测试
- **性能验证**：
  - 状态更新性能测试
  - 并发访问性能测试
- **回归测试**：
  - 确保与Java版本功能一致
- **测试结果**：
  - 待实现

### A6 迭代（Advance）
- 性能优化：状态更新优化、缓存策略优化
- 功能扩展：支持更多状态类型、监控指标
- 观测性增强：添加详细日志、性能监控
- 下一步任务链接：Task-03 实现客户端驱动

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
[总结任务完成情况]
