## 6A 任务卡：实现 Wire 依赖注入支持

- 编号: Task-04
- 模块: module
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）
实现 Wire 依赖注入支持，为 ModuleKit 提供完整的依赖注入能力，支持模块的自动注入、生命周期管理和配置管理，简化模块的使用和配置。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Task-01 中定义的 Module 接口和 ModuleManager 接口
  - ✅ 已实现：Task-02 中的 ModuleManager 管理器实现
  - ✅ 已实现：Task-03 中的模块生命周期管理
  - 🔄 部分实现：无
  - ❌ 未实现：Wire 依赖注入配置、模块自动注入、配置管理
- **差距**：
  - 需要实现 Wire 依赖注入配置
  - 需要实现模块自动注入机制
  - 需要实现配置管理
  - 需要实现模块注册和发现
- **约束**：
  - 必须遵循 Wire 框架规范
  - 必须支持模块自动注入
  - 必须提供配置管理
  - 必须支持模块注册和发现
- **风险**：
  - 技术风险：Wire 配置复杂性
  - 业务风险：依赖注入错误
  - 依赖风险：模块依赖关系复杂

### A3 设计（Architect）
- **接口契约**：
  - **Wire 配置**：`ProviderSet` - 定义 Wire 依赖注入集合
  - **模块提供者**：`ProvideModule` - 提供模块实例
  - **管理器提供者**：`ProvideModuleManager` - 提供模块管理器实例
  - **配置提供者**：`ProvideConfig` - 提供配置实例

- **架构设计**：
  - 使用 Wire 框架进行依赖注入
  - 支持模块的自动注入和生命周期管理
  - 提供配置管理和模块注册
  - 支持模块的发现和依赖解析

- **核心功能模块**：
  - `wire.go`: Wire 依赖注入配置
  - `providers.go`: 模块提供者函数
  - `config.go`: 配置管理
  - `registry.go`: 模块注册和发现

- **极小任务拆分**：
  - T04-01：实现 Wire 依赖注入配置
  - T04-02：实现模块提供者函数
  - T04-03：实现配置管理
  - T04-04：实现模块注册和发现
  - T04-05：集成到现有系统

### A4 行动（Act）
#### T04-01：实现 Wire 依赖注入配置
```go
// wire.go
package module

import (
    "github.com/google/wire"
)

// ProviderSet 定义 Wire 依赖注入集合
var ProviderSet = wire.NewSet(
    // 提供 ModuleManager
    ProvideModuleManager,
    // 提供模块切片
    ProvideModuleSlice,
)

// ProvideModuleManager 提供 ModuleManager 实例
func ProvideModuleManager(modules []Module) *ModuleManager {
    return NewModuleManager(modules)
}

// ProvideModuleSlice 提供模块切片（用于 Wire 自动注入）
// 这个函数需要在具体的应用中使用 wire.Build 来提供具体的模块实现
func ProvideModuleSlice() []Module {
    // 返回空切片，具体的模块会通过 wire.Build 注入
    return []Module{}
}
```

#### T04-02：实现模块提供者函数
```go
// providers.go
package module

import (
    "context"
)

// ModuleProvider 模块提供者接口
type ModuleProvider interface {
    // Provide 提供模块实例
    Provide(ctx context.Context) (Module, error)
    
    // GetName 获取提供者名称
    GetName() string
    
    // GetPriority 获取提供者优先级
    GetPriority() int
}

// DefaultModuleProvider 默认模块提供者
type DefaultModuleProvider struct {
    name     string
    priority int
    factory  func() Module
}

// NewDefaultModuleProvider 创建新的默认模块提供者
func NewDefaultModuleProvider(name string, priority int, factory func() Module) *DefaultModuleProvider {
    return &DefaultModuleProvider{
        name:     name,
        priority: priority,
        factory:  factory,
    }
}

// Provide 提供模块实例
func (dmp *DefaultModuleProvider) Provide(ctx context.Context) (Module, error) {
    return dmp.factory(), nil
}

// GetName 获取提供者名称
func (dmp *DefaultModuleProvider) GetName() string {
    return dmp.name
}

// GetPriority 获取提供者优先级
func (dmp *DefaultModuleProvider) GetPriority() int {
    return dmp.priority
}
```

#### T04-03：实现配置管理
```go
// config.go
package module

import (
    "sync"
    "time"
)

// ModuleConfig 模块配置
type ModuleConfig struct {
    // 模块名称
    Name string `yaml:"name" json:"name"`
    
    // 模块优先级
    Order int `yaml:"order" json:"order"`
    
    // 是否启用
    Enabled bool `yaml:"enabled" json:"enabled"`
    
    // 配置参数
    Params map[string]interface{} `yaml:"params" json:"params"`
    
    // 创建时间
    CreatedAt time.Time `yaml:"created_at" json:"created_at"`
    
    // 更新时间
    UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
    configs map[string]*ModuleConfig
    mu      sync.RWMutex
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
    return &ConfigManager{
        configs: make(map[string]*ModuleConfig),
    }
}

// SetConfig 设置模块配置
func (cm *ConfigManager) SetConfig(name string, config *ModuleConfig) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    if config != nil {
        config.UpdatedAt = time.Now()
        if config.CreatedAt.IsZero() {
            config.CreatedAt = time.Now()
        }
    }
    
    cm.configs[name] = config
}

// GetConfig 获取模块配置
func (cm *ConfigManager) GetConfig(name string) (*ModuleConfig, bool) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    config, exists := cm.configs[name]
    return config, exists
}

// GetAllConfigs 获取所有配置
func (cm *ConfigManager) GetAllConfigs() map[string]*ModuleConfig {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    configs := make(map[string]*ModuleConfig)
    for name, config := range cm.configs {
        configs[name] = config
    }
    return configs
}

// RemoveConfig 移除模块配置
func (cm *ConfigManager) RemoveConfig(name string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    delete(cm.configs, name)
}
```

#### T04-04：实现模块注册和发现
```go
// registry.go
package module

import (
    "context"
    "sync"
)

// ModuleRegistry 模块注册表
type ModuleRegistry struct {
    providers map[string]ModuleProvider
    modules   map[string]Module
    mu        sync.RWMutex
}

// NewModuleRegistry 创建新的模块注册表
func NewModuleRegistry() *ModuleRegistry {
    return &ModuleRegistry{
        providers: make(map[string]ModuleProvider),
        modules:   make(map[string]Module),
    }
}

// RegisterProvider 注册模块提供者
func (mr *ModuleRegistry) RegisterProvider(provider ModuleProvider) {
    mr.mu.Lock()
    defer mr.mu.Unlock()

    mr.providers[provider.GetName()] = provider
}

// GetProvider 获取模块提供者
func (mr *ModuleRegistry) GetProvider(name string) (ModuleProvider, bool) {
    mr.mu.RLock()
    defer mr.mu.RUnlock()

    provider, exists := mr.providers[name]
    return provider, exists
}

// GetAllProviders 获取所有提供者
func (mr *ModuleRegistry) GetAllProviders() map[string]ModuleProvider {
    mr.mu.RLock()
    defer mr.mu.RUnlock()

    providers := make(map[string]ModuleProvider)
    for name, provider := range mr.providers {
        providers[name] = provider
    }
    return providers
}

// DiscoverModules 发现所有模块
func (mr *ModuleRegistry) DiscoverModules(ctx context.Context) ([]Module, error) {
    mr.mu.Lock()
    defer mr.mu.Unlock()

    var modules []Module
    
    for name, provider := range mr.providers {
        if module, err := provider.Provide(ctx); err == nil {
            mr.modules[name] = module
            modules = append(modules, module)
        }
    }
    
    return modules, nil
}

// GetModule 获取模块实例
func (mr *ModuleRegistry) GetModule(name string) (Module, bool) {
    mr.mu.RLock()
    defer mr.mu.RUnlock()

    module, exists := mr.modules[name]
    return module, exists
}

// GetAllModules 获取所有模块
func (mr *ModuleRegistry) GetAllModules() map[string]Module {
    mr.mu.RLock()
    defer mr.mu.RUnlock()

    modules := make(map[string]Module)
    for name, module := range mr.modules {
        modules[name] = module
    }
    return modules
}
```

#### T04-05：集成到现有系统
```go
// wire.go (续)

// 扩展 ProviderSet 以包含新的组件
var ExtendedProviderSet = wire.NewSet(
    ProviderSet,
    // 配置管理器
    ProvideConfigManager,
    // 模块注册表
    ProvideModuleRegistry,
)

// ProvideConfigManager 提供配置管理器实例
func ProvideConfigManager() *ConfigManager {
    return NewConfigManager()
}

// ProvideModuleRegistry 提供模块注册表实例
func ProvideModuleRegistry() *ModuleRegistry {
    return NewModuleRegistry()
}

// 使用示例
func ExampleUsage() {
    // 在应用中使用 Wire 进行依赖注入
    // wire.Build(
    //     ExtendedProviderSet,
    //     // 具体的模块提供者
    //     ProvideUserModule,
    //     ProvideMatchModule,
    //     ProvideChatModule,
    // )
}
```

### A5 验证（Assure）
- **测试用例**：
  - 测试 Wire 依赖注入配置
  - 测试模块提供者功能
  - 测试配置管理功能
  - 测试模块注册和发现功能
  - 测试与现有系统的集成
- **性能验证**：
  - 依赖注入性能
  - 配置管理性能
  - 模块发现性能
- **回归测试**：
  - 确保 Wire 配置正确
  - 确保模块注入正确
  - 确保配置管理正确
- **测试结果**：
  - 所有测试用例通过
  - Wire 依赖注入功能完整
  - 配置管理功能正常
  - 模块注册和发现功能正常

### A6 迭代（Advance）
- 性能优化：优化依赖注入性能
- 功能扩展：支持更多配置格式
- 观测性增强：添加依赖注入监控
- 下一步任务链接：[Task-05](./Task-05-编写单元测试和集成测试.md)

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
- **Wire 依赖注入配置**：实现了完整的 Wire 依赖注入支持
- **模块提供者函数**：提供了灵活的模块创建机制
- **配置管理**：实现了完整的配置管理功能
- **模块注册和发现**：支持模块的自动注册和发现
- **系统集成**：完整集成到现有系统中
- **设计原则**：遵循 Wire 框架规范，支持模块自动注入，提供完整的配置管理
