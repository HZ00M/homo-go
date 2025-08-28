## 6A 任务卡：定义核心接口和类型

- 编号: Task-01
- 模块: module
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）
定义 ModuleKit 的核心接口和类型，包括 Module 接口、ModuleManager 接口、错误类型、生命周期阶段等，为模块生命周期管理提供基础契约。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：无
  - 🔄 部分实现：无
  - ❌ 未实现：核心接口定义、类型定义、错误码定义
- **差距**：
  - 需要定义 Module 接口
  - 需要定义 ModuleManager 接口
  - 需要定义错误类型和错误码
  - 需要定义生命周期阶段枚举
- **约束**：
  - 遵循 Go 语言最佳实践
  - 符合 Kratos 框架设计理念
  - 支持 Wire 依赖注入
- **风险**：
  - 技术风险：接口设计不合理
  - 业务风险：生命周期管理复杂
  - 依赖风险：与 Kratos 框架兼容性

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`Module` - 定义模块基础接口
  - **核心方法**：
    - `Name() string` - 返回模块名称
    - `Order() int` - 返回模块初始化顺序
    - `Init(ctx context.Context) error` - 初始化模块
    - `AfterAllInit(ctx context.Context)` - 所有模块初始化后调用
    - `AfterStart(ctx context.Context)` - 应用启动后调用
    - `BeforeClose(ctx context.Context) error` - 模块关闭前调用
    - `AfterStop(ctx context.Context)` - 应用停止后调用
  - **管理器接口**：`ModuleManager` - 定义模块管理器接口
  - **核心方法**：
    - `SetModules(modules []Module)` - 设置模块列表
    - `GetModules() []Module` - 获取模块列表
    - `InitAll(ctx context.Context) error` - 初始化所有模块
    - `AfterAllInit(ctx context.Context)` - 调用所有模块的 AfterAllInit
    - `AfterStart(ctx context.Context)` - 调用所有模块的 AfterStart
    - `CloseAll(ctx context.Context) error` - 关闭所有模块
    - `AfterStop(ctx context.Context)` - 调用所有模块的 AfterStop

- **架构设计**：
  - 采用接口分离原则，将不同功能模块分离到不同接口
  - 使用 Wire 依赖注入，简化架构设计
  - 支持模块优先级管理
  - 提供完整的生命周期管理

- **核心功能模块**：
  - `Module`: 模块基础接口
  - `ModuleManager`: 模块管理器接口
  - `ErrorTypes`: 错误类型定义
  - `LifecycleStage`: 生命周期阶段枚举

- **极小任务拆分**：
  - T01-01：定义 `Module` 核心接口
  - T01-02：定义 `ModuleManager` 接口
  - T01-03：定义错误类型和错误码
  - T01-04：定义生命周期阶段枚举
  - T01-05：定义模块状态和汇总类型

### A4 行动（Act）
#### T01-01：定义 `Module` 核心接口
```go
// interfaces.go
package module

import (
    "context"
)

// Module 模块基础接口
type Module interface {
    // Name 返回模块名称
    Name() string
    
    // Order 返回模块的初始化顺序，低值优先
    Order() int
    
    // Init 初始化模块
    Init(ctx context.Context) error
    
    // AfterAllInit 所有模块初始化后调用
    AfterAllInit(ctx context.Context)
    
    // AfterStart 应用启动后调用
    AfterStart(ctx context.Context)
    
    // BeforeClose 模块关闭前调用
    BeforeClose(ctx context.Context) error
    
    // AfterStop 应用停止后调用
    AfterStop(ctx context.Context)
}
```

#### T01-02：定义 `ModuleManager` 接口
```go
// interfaces.go (续)

// ModuleManager 模块管理器接口
type ModuleManager interface {
    // SetModules 设置模块列表
    SetModules(modules []Module)
    
    // GetModules 获取模块列表（已排序）
    GetModules() []Module
    
    // InitAll 初始化所有模块
    InitAll(ctx context.Context) error
    
    // AfterAllInit 调用所有模块的 AfterAllInit
    AfterAllInit(ctx context.Context)
    
    // AfterStart 调用所有模块的 AfterStart
    AfterStart(ctx context.Context)
    
    // CloseAll 关闭所有模块
    CloseAll(ctx context.Context) error
    
    // AfterStop 调用所有模块的 AfterStop
    AfterStop(ctx context.Context)
    
    // RegisterToApp 将模块生命周期注册到应用
    RegisterToApp(app interface{}) error
}
```

#### T01-03：定义错误类型和错误码
```go
// types.go
package module

import (
    "github.com/go-kratos/kratos/v2/errors"
)

// 错误码定义
const (
    ErrModuleNotFound     = "MODULE_NOT_FOUND"
    ErrModuleInitFailed   = "MODULE_INIT_FAILED"
    ErrModuleCloseFailed  = "MODULE_CLOSE_FAILED"
    ErrInvalidModuleOrder = "INVALID_MODULE_ORDER"
    ErrDuplicateModule    = "DUPLICATE_MODULE"
)

// 错误类型定义
var (
    ErrNotFound     = errors.New(404, ErrModuleNotFound, "module not found")
    ErrInitFailed   = errors.New(500, ErrModuleInitFailed, "module initialization failed")
    ErrCloseFailed  = errors.New(500, ErrModuleCloseFailed, "module close failed")
    ErrInvalidOrder = errors.New(400, ErrInvalidModuleOrder, "invalid module order")
    ErrDuplicate    = errors.New(400, ErrDuplicateModule, "duplicate module")
)
```

#### T01-04：定义生命周期阶段枚举
```go
// types.go (续)

// LifecycleStage 生命周期阶段
type LifecycleStage int

const (
    StageUnknown LifecycleStage = iota
    StageInit
    StageAfterAllInit
    StageAfterStart
    StageRunning
    StageBeforeClose
    StageAfterStop
)

// String 返回生命周期阶段的字符串表示
func (s LifecycleStage) String() string {
    switch s {
    case StageUnknown:
        return "unknown"
    case StageInit:
        return "init"
    case StageAfterAllInit:
        return "after_all_init"
    case StageAfterStart:
        return "after_start"
    case StageRunning:
        return "running"
    case StageBeforeClose:
        return "before_close"
    case StageAfterStop:
        return "after_stop"
    default:
        return "unknown"
    }
}
```

#### T01-05：定义模块状态和汇总类型
```go
// types.go (续)

import (
    "time"
)

// ModuleStatus 模块状态
type ModuleStatus struct {
    Name          string         `json:"name"`
    Order         int            `json:"order"`
    Stage         LifecycleStage `json:"stage"`
    InitTime      time.Time      `json:"init_time,omitempty"`
    StartTime     time.Time      `json:"start_time,omitempty"`
    LastError     error          `json:"last_error,omitempty"`
    IsInitialized bool           `json:"is_initialized"`
    IsStarted     bool           `json:"is_started"`
    IsRunning     bool           `json:"is_running"`
    IsClosed      bool           `json:"is_closed"`
}

// ModuleSummary 模块汇总信息
type ModuleSummary struct {
    TotalCount  int            `json:"total_count"`
    Initialized int            `json:"initialized"`
    Started     int            `json:"started"`
    Running     int            `json:"running"`
    Closed      int            `json:"closed"`
    Failed      int            `json:"failed"`
    StartTime   time.Time      `json:"start_time"`
    Modules     []ModuleStatus `json:"modules"`
}
```

### A5 验证（Assure）
- **测试用例**：
  - 测试 Module 接口的方法定义
  - 测试 ModuleManager 接口的方法定义
  - 测试错误类型和错误码定义
  - 测试生命周期阶段枚举
  - 测试模块状态和汇总类型
- **性能验证**：
  - 接口调用性能
  - 类型转换性能
- **回归测试**：
  - 确保接口设计合理
  - 确保类型定义正确
- **测试结果**：
  - 所有接口定义完成
  - 所有类型定义完成
  - 基础功能测试通过

### A6 迭代（Advance）
- 性能优化：优化接口调用性能
- 功能扩展：支持更多生命周期阶段
- 观测性增强：添加模块状态监控
- 下一步任务链接：[Task-02](./Task-02-实现ModuleManager单例.md)

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
- **Module 接口**：定义了完整的模块生命周期接口，支持 Wire 依赖注入
- **ModuleManager 接口**：定义了模块管理器接口，支持模块生命周期管理
- **错误类型和错误码**：使用 Kratos 错误系统，提供完整的错误处理
- **生命周期阶段枚举**：定义了完整的模块生命周期阶段
- **模块状态和汇总类型**：提供了模块状态管理和监控支持
- **设计原则**：遵循 Go 语言最佳实践，支持 Wire 依赖注入，简化架构设计
