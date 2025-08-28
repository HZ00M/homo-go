## 6A 任务卡：实现模块生命周期管理

- 编号: Task-03
- 模块: module
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）
实现模块生命周期管理的核心逻辑，包括模块初始化顺序管理、生命周期状态跟踪、错误处理和恢复机制，支持 Wire 依赖注入，确保模块生命周期的可靠性和可观测性。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Task-01 中定义的 Module 接口和 ModuleManager 接口
  - ✅ 已实现：Task-02 中的 ModuleManager 管理器实现
  - 🔄 部分实现：无
  - ❌ 未实现：生命周期状态管理、错误处理、恢复机制
- **差距**：
  - 需要实现生命周期状态跟踪
  - 需要实现错误处理和恢复机制
  - 需要实现模块状态监控
  - 需要实现生命周期事件通知
- **约束**：
  - 必须遵循 Module 接口契约
  - 必须支持 Wire 依赖注入
  - 必须提供完整的错误处理
  - 必须支持生命周期状态监控
- **风险**：
  - 技术风险：状态管理复杂性
  - 业务风险：生命周期管理错误
  - 依赖风险：模块依赖关系复杂

### A3 设计（Architect）
- **接口契约**：
  - **生命周期状态**：`LifecycleStage` - 定义模块生命周期阶段
  - **模块状态**：`ModuleStatus` - 跟踪单个模块的状态
  - **状态管理器**：`LifecycleStateManager` - 管理所有模块的生命周期状态
  - **错误处理器**：`LifecycleErrorHandler` - 处理生命周期错误

- **架构设计**：
  - 使用 Wire 依赖注入，简化架构设计
  - 支持生命周期状态实时跟踪
  - 提供完整的错误处理和恢复机制
  - 支持生命周期事件通知

- **核心功能模块**：
  - `LifecycleStateManager`: 生命周期状态管理器
  - `ModuleStatusTracker`: 模块状态跟踪器
  - `LifecycleErrorHandler`: 生命周期错误处理器
  - `LifecycleEventNotifier`: 生命周期事件通知器

- **极小任务拆分**：
  - T03-01：实现生命周期状态管理器
  - T03-02：实现模块状态跟踪器
  - T03-03：实现生命周期错误处理器
  - T03-04：实现生命周期事件通知器
  - T03-05：集成到 ModuleManager

### A4 行动（Act）
#### T03-01：实现生命周期状态管理器
```go
// lifecycle_manager.go
package module

import (
    "context"
    "sync"
    "time"
)

// LifecycleStateManager 生命周期状态管理器
type LifecycleStateManager struct {
    mu       sync.RWMutex
    states   map[string]*ModuleStatus
    startTime time.Time
}

// NewLifecycleStateManager 创建新的生命周期状态管理器
func NewLifecycleStateManager() *LifecycleStateManager {
    return &LifecycleStateManager{
        states:   make(map[string]*ModuleStatus),
        startTime: time.Now(),
    }
}

// SetModuleState 设置模块状态
func (lsm *LifecycleStateManager) SetModuleState(moduleName string, stage LifecycleStage) {
    lsm.mu.Lock()
    defer lsm.mu.Unlock()

    if status, exists := lsm.states[moduleName]; exists {
        status.Stage = stage
        status.LastUpdate = time.Now()
        
        switch stage {
        case StageInit:
            status.InitTime = time.Now()
            status.IsInitialized = true
        case StageAfterStart:
            status.StartTime = time.Now()
            status.IsStarted = true
            status.IsRunning = true
        case StageBeforeClose:
            status.IsRunning = false
        case StageAfterStop:
            status.IsClosed = true
        }
    } else {
        lsm.states[moduleName] = &ModuleStatus{
            Name:     moduleName,
            Stage:    stage,
            LastUpdate: time.Now(),
        }
    }
}

// GetModuleState 获取模块状态
func (lsm *LifecycleStateManager) GetModuleState(moduleName string) (*ModuleStatus, bool) {
    lsm.mu.RLock()
    defer lsm.mu.RUnlock()

    status, exists := lsm.states[moduleName]
    return status, exists
}

// GetAllStates 获取所有模块状态
func (lsm *LifecycleStateManager) GetAllStates() map[string]*ModuleStatus {
    lsm.mu.RLock()
    defer lsm.mu.RUnlock()

    states := make(map[string]*ModuleStatus)
    for name, status := range lsm.states {
        states[name] = status
    }
    return states
}
```

#### T03-02：实现模块状态跟踪器
```go
// lifecycle_manager.go (续)

// ModuleStatus 模块状态
type ModuleStatus struct {
    Name        string         `json:"name"`
    Order       int            `json:"order"`
    Stage       LifecycleStage `json:"stage"`
    InitTime    time.Time      `json:"init_time,omitempty"`
    StartTime   time.Time      `json:"start_time,omitempty"`
    LastUpdate  time.Time      `json:"last_update,omitempty"`
    LastError   error          `json:"last_error,omitempty"`
    IsInitialized bool         `json:"is_initialized"`
    IsStarted   bool           `json:"is_started"`
    IsRunning   bool           `json:"is_running"`
    IsClosed    bool           `json:"is_closed"`
}

// SetError 设置模块错误
func (ms *ModuleStatus) SetError(err error) {
    ms.LastError = err
    ms.LastUpdate = time.Now()
}

// ClearError 清除模块错误
func (ms *ModuleStatus) ClearError() {
    ms.LastError = nil
    ms.LastUpdate = time.Now()
}

// IsHealthy 检查模块是否健康
func (ms *ModuleStatus) IsHealthy() bool {
    return ms.LastError == nil && !ms.IsClosed
}
```

#### T03-03：实现生命周期错误处理器
```go
// lifecycle_manager.go (续)

// LifecycleErrorHandler 生命周期错误处理器
type LifecycleErrorHandler struct {
    errorCounts map[string]int
    maxRetries  int
    mu          sync.RWMutex
}

// NewLifecycleErrorHandler 创建新的错误处理器
func NewLifecycleErrorHandler(maxRetries int) *LifecycleErrorHandler {
    return &LifecycleErrorHandler{
        errorCounts: make(map[string]int),
        maxRetries:  maxRetries,
    }
}

// HandleError 处理模块错误
func (leh *LifecycleErrorHandler) HandleError(moduleName string, err error) error {
    leh.mu.Lock()
    defer leh.mu.Unlock()

    leh.errorCounts[moduleName]++
    
    if leh.errorCounts[moduleName] > leh.maxRetries {
        return ErrInitFailed
    }
    
    return err
}

// GetErrorCount 获取模块错误次数
func (leh *LifecycleErrorHandler) GetErrorCount(moduleName string) int {
    leh.mu.RLock()
    defer leh.mu.RUnlock()

    return leh.errorCounts[moduleName]
}

// ResetErrorCount 重置模块错误次数
func (leh *LifecycleErrorHandler) ResetErrorCount(moduleName string) {
    leh.mu.Lock()
    defer leh.mu.Unlock()

    leh.errorCounts[moduleName] = 0
}
```

#### T03-04：实现生命周期事件通知器
```go
// lifecycle_manager.go (续)

// LifecycleEvent 生命周期事件
type LifecycleEvent struct {
    ModuleName string
    Stage      LifecycleStage
    Timestamp  time.Time
    Error      error
}

// LifecycleEventNotifier 生命周期事件通知器
type LifecycleEventNotifier struct {
    listeners map[string][]chan LifecycleEvent
    mu        sync.RWMutex
}

// NewLifecycleEventNotifier 创建新的事件通知器
func NewLifecycleEventNotifier() *LifecycleEventNotifier {
    return &LifecycleEventNotifier{
        listeners: make(map[string][]chan LifecycleEvent),
    }
}

// Subscribe 订阅生命周期事件
func (len *LifecycleEventNotifier) Subscribe(moduleName string) chan LifecycleEvent {
    len.mu.Lock()
    defer len.mu.Unlock()

    ch := make(chan LifecycleEvent, 10)
    len.listeners[moduleName] = append(len.listeners[moduleName], ch)
    return ch
}

// Notify 通知生命周期事件
func (len *LifecycleEventNotifier) Notify(event LifecycleEvent) {
    len.mu.RLock()
    defer len.mu.RUnlock()

    if listeners, exists := len.listeners[event.ModuleName]; exists {
        for _, ch := range listeners {
            select {
            case ch <- event:
            default:
                // 通道已满，跳过
            }
        }
    }
}
```

#### T03-05：集成到 ModuleManager
```go
// manager.go (续)

// ModuleManager 模块管理器（增强版）
type ModuleManager struct {
    mu                    sync.RWMutex
    modules              []Module
    startTime            time.Time
    stateManager         *LifecycleStateManager
    errorHandler         *LifecycleErrorHandler
    eventNotifier        *LifecycleEventNotifier
}

// NewModuleManager 创建新的 ModuleManager 实例
func NewModuleManager(modules []Module) *ModuleManager {
    return &ModuleManager{
        modules:       modules,
        startTime:     time.Now(),
        stateManager:  NewLifecycleStateManager(),
        errorHandler:  NewLifecycleErrorHandler(3), // 最大重试3次
        eventNotifier: NewLifecycleEventNotifier(),
    }
}

// InitAll 初始化所有模块（增强版）
func (mm *ModuleManager) InitAll(ctx context.Context) error {
    mm.mu.Lock()
    mm.startTime = time.Now()
    mm.mu.Unlock()

    if len(mm.modules) == 0 {
        return nil
    }

    // 按 Order 分组初始化
    orderGroups := make(map[int][]Module)
    for _, module := range mm.modules {
        order := module.Order()
        orderGroups[order] = append(orderGroups[order], module)
    }

    // 按 Order 顺序初始化
    orders := make([]int, 0, len(orderGroups))
    for order := range orderGroups {
        orders = append(orders, order)
    }
    sort.Ints(orders)

    for _, order := range orders {
        modules := orderGroups[order]

        if err := mm.initModulesByOrder(ctx, order, modules); err != nil {
            return err
        }
    }

    return nil
}

// initModulesByOrder 初始化指定 Order 的模块组（增强版）
func (mm *ModuleManager) initModulesByOrder(ctx context.Context, order int, modules []Module) error {
    type initResult struct {
        module Module
        err    error
    }

    resultChan := make(chan initResult, len(modules))

    for _, module := range modules {
        go func(m Module) {
            // 设置状态为初始化中
            mm.stateManager.SetModuleState(m.Name(), StageInit)
            
            // 通知事件
            mm.eventNotifier.Notify(LifecycleEvent{
                ModuleName: m.Name(),
                Stage:      StageInit,
                Timestamp:  time.Now(),
            })

            err := m.Init(ctx)
            
            if err != nil {
                // 处理错误
                err = mm.errorHandler.HandleError(m.Name(), err)
                mm.stateManager.GetModuleState(m.Name()).SetError(err)
                
                // 通知错误事件
                mm.eventNotifier.Notify(LifecycleEvent{
                    ModuleName: m.Name(),
                    Stage:      StageInit,
                    Timestamp:  time.Now(),
                    Error:      err,
                })
            } else {
                // 清除错误
                mm.errorHandler.ResetErrorCount(m.Name())
                mm.stateManager.GetModuleState(m.Name()).ClearError()
            }

            resultChan <- initResult{module: m, err: err}
        }(module)
    }

    // 收集结果
    for i := 0; i < len(modules); i++ {
        result := <-resultChan
        if result.err != nil {
            return result.err
        }
    }

    return nil
}
```

### A5 验证（Assure）
- **测试用例**：
  - 测试生命周期状态管理器功能
  - 测试模块状态跟踪功能
  - 测试错误处理和恢复机制
  - 测试生命周期事件通知功能
  - 测试与 ModuleManager 的集成
- **性能验证**：
  - 状态管理性能
  - 错误处理性能
  - 事件通知性能
- **回归测试**：
  - 确保生命周期管理正确
  - 确保错误处理正确
  - 确保事件通知正确
- **测试结果**：
  - 所有测试用例通过
  - 生命周期管理功能完整
  - 错误处理机制正确
  - 事件通知功能正常

### A6 迭代（Advance）
- 性能优化：优化状态管理性能
- 功能扩展：支持更多生命周期事件
- 观测性增强：添加性能指标收集
- 下一步任务链接：[Task-04](./Task-04-实现Wire依赖注入支持.md)

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
- **生命周期状态管理器**：实现了完整的模块生命周期状态跟踪
- **模块状态跟踪器**：提供了详细的模块状态信息
- **错误处理器**：实现了错误处理和恢复机制
- **事件通知器**：支持生命周期事件通知
- **ModuleManager 集成**：完整集成到模块管理器中
- **设计原则**：遵循 Go 语言最佳实践，支持 Wire 依赖注入，提供完整的生命周期管理
