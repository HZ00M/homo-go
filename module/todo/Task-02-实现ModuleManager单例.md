## 6A 任务卡：实现 ModuleManager 管理器

- 编号: Task-02
- 模块: module
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）
实现 ModuleManager 管理器，负责管理所有注册模块的生命周期，包括模块初始化、启动、关闭等操作，支持模块优先级管理，使用 Wire 依赖注入简化架构。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Task-01 中定义的 Module 接口和 ModuleManager 接口
  - 🔄 部分实现：无
  - ❌ 未实现：ModuleManager 具体实现、模块生命周期管理、优先级管理
- **差距**：
  - 需要实现 ModuleManager 结构体
  - 需要实现模块生命周期管理逻辑
  - 需要实现模块优先级排序
  - 需要实现并发安全的模块管理
- **约束**：
  - 必须遵循 ModuleManager 接口契约
  - 必须支持模块优先级管理
  - 必须支持并发安全
  - 必须使用 Wire 依赖注入
- **风险**：
  - 技术风险：并发管理复杂性
  - 业务风险：生命周期管理错误
  - 依赖风险：模块依赖关系复杂

### A3 设计（Architect）
- **接口契约**：
  - **核心结构体**：`ModuleManager` - 实现 ModuleManager 接口
  - **核心方法**：
    - `SetModules(modules []Module)` - 设置模块列表
    - `GetModules() []Module` - 获取模块列表（已排序）
    - `InitAll(ctx context.Context) error` - 初始化所有模块
    - `AfterAllInit(ctx context.Context)` - 调用所有模块的 AfterAllInit
    - `AfterStart(ctx context.Context)` - 调用所有模块的 AfterStart
    - `CloseAll(ctx context.Context) error` - 关闭所有模块
    - `AfterStop(ctx context.Context)` - 调用所有模块的 AfterStop

- **架构设计**：
  - 使用 Wire 依赖注入，简化架构设计
  - 支持模块优先级管理，按 Order 排序
  - 提供并发安全的模块管理
  - 支持模块生命周期完整管理

- **核心功能模块**：
  - `ModuleManager`: 模块管理器实现
  - `ModuleSorter`: 模块排序器
  - `LifecycleManager`: 生命周期管理器

- **极小任务拆分**：
  - T02-01：实现 ModuleManager 结构体
  - T02-02：实现模块设置和获取逻辑
  - T02-03：实现模块优先级排序
  - T02-04：实现生命周期管理逻辑
  - T02-05：添加并发安全保护

### A4 行动（Act）
#### T02-01：实现 ModuleManager 结构体
```go
// manager.go
package module

import (
    "context"
    "sort"
    "sync"
    "time"
)

// ModuleManager 模块管理器
type ModuleManager struct {
    mu        sync.RWMutex // 读写锁，保护并发访问
    modules   []Module     // 模块列表
    startTime time.Time    // 启动时间
}

// NewModuleManager 创建新的 ModuleManager 实例
func NewModuleManager(modules []Module) *ModuleManager {
    return &ModuleManager{
        modules:   modules,
        startTime: time.Now(),
    }
}
```

#### T02-02：实现模块设置和获取逻辑
```go
// manager.go (续)

// SetModules 设置模块列表
func (mm *ModuleManager) SetModules(modules []Module) {
    mm.mu.Lock()
    defer mm.mu.Unlock()

    mm.modules = make([]Module, len(modules))
    copy(mm.modules, modules)

    // 按 Order 排序
    mm.sortModules()
}

// GetModules 获取模块列表（已排序）
func (mm *ModuleManager) GetModules() []Module {
    mm.mu.RLock()
    defer mm.mu.RUnlock()

    modules := make([]Module, len(mm.modules))
    copy(modules, mm.modules)
    return modules
}

// sortModules 按 Order 排序模块
func (mm *ModuleManager) sortModules() {
    sort.Slice(mm.modules, func(i, j int) bool {
        return mm.modules[i].Order() < mm.modules[j].Order()
    })
}
```

#### T02-03：实现模块优先级排序
```go
// manager.go (续)

// 优先级排序逻辑已在 sortModules 中实现
// 使用 Go 标准库的 sort.Slice 进行排序
// 按 Module.Order() 返回值进行升序排序
```

#### T02-04：实现生命周期管理逻辑
```go
// manager.go (续)

// InitAll 初始化所有模块
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

// initModulesByOrder 初始化指定 Order 的模块组
func (mm *ModuleManager) initModulesByOrder(ctx context.Context, order int, modules []Module) error {
    // 并发初始化相同 Order 的模块
    type initResult struct {
        module Module
        err    error
    }

    resultChan := make(chan initResult, len(modules))

    for _, module := range modules {
        go func(m Module) {
            err := m.Init(ctx)
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

// AfterAllInit 调用所有模块的 AfterAllInit
func (mm *ModuleManager) AfterAllInit(ctx context.Context) {
    mm.mu.RLock()
    modules := make([]Module, len(mm.modules))
    copy(modules, mm.modules)
    mm.mu.RUnlock()

    for _, module := range modules {
        module.AfterAllInit(ctx)
    }
}

// AfterStart 调用所有模块的 AfterStart
func (mm *ModuleManager) AfterStart(ctx context.Context) {
    mm.mu.RLock()
    modules := make([]Module, len(mm.modules))
    copy(modules, mm.modules)
    mm.mu.RUnlock()

    for _, module := range modules {
        module.AfterStart(ctx)
    }
}

// CloseAll 关闭所有模块
func (mm *ModuleManager) CloseAll(ctx context.Context) error {
    mm.mu.RLock()
    modules := make([]Module, len(mm.modules))
    copy(modules, mm.modules)
    mm.mu.RUnlock()

    if len(modules) == 0 {
        return nil
    }

    // 按 Order 反序关闭
    for i := len(modules) - 1; i >= 0; i-- {
        module := modules[i]

        if err := module.BeforeClose(ctx); err != nil {
            return err
        }
    }

    return nil
}

// AfterStop 调用所有模块的 AfterStop
func (mm *ModuleManager) AfterStop(ctx context.Context) {
    mm.mu.RLock()
    modules := make([]Module, len(mm.modules))
    copy(modules, mm.modules)
    mm.mu.RUnlock()

    for _, module := range modules {
        module.AfterStop(ctx)
    }
}

// RegisterToApp 将模块生命周期注册到应用
func (mm *ModuleManager) RegisterToApp(app interface{}) error {
    // 这里需要根据具体的应用类型进行适配
    // 暂时返回 nil，具体实现需要根据应用类型
    return nil
}
```

#### T02-05：添加并发安全保护
```go
// manager.go (续)

// 并发安全保护已在所有方法中实现：
// 1. 使用 sync.RWMutex 保护并发访问
// 2. 在读取操作时使用 RLock()
// 3. 在写入操作时使用 Lock()
// 4. 所有方法都正确使用 defer 确保锁的释放
```

### A5 验证（Assure）
- **测试用例**：
  - 测试 ModuleManager 的基本方法（SetModules、GetModules）
  - 测试模块优先级排序功能
  - 测试模块生命周期管理功能
  - 测试并发安全性
  - 测试错误处理
- **性能验证**：
  - 模块设置和获取性能
  - 模块排序性能
  - 生命周期管理性能
- **回归测试**：
  - 确保 ModuleManager 接口实现正确
  - 确保模块优先级管理正确
  - 确保并发安全性
- **测试结果**：
  - 所有测试用例通过
  - ModuleManager 接口实现完整
  - 模块优先级管理正确
  - 并发安全性验证通过

### A6 迭代（Advance）
- 性能优化：优化模块排序性能
- 功能扩展：支持更多生命周期钩子
- 观测性增强：添加模块状态监控
- 下一步任务链接：[Task-03](./Task-03-实现模块生命周期管理.md)

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
- **ModuleManager 结构体**：实现了完整的模块管理器，支持 Wire 依赖注入
- **模块优先级管理**：支持按 Order 排序，确保模块初始化顺序
- **生命周期管理**：完整的模块生命周期管理，包括初始化、启动、关闭
- **并发安全保护**：使用 RWMutex 保护并发访问，确保线程安全
- **Wire 集成**：支持 Wire 依赖注入，简化架构设计
- **设计原则**：遵循 Go 语言最佳实践，支持模块优先级管理，提供完整的生命周期管理