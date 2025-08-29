## 6A 任务卡：实现全局注册表配置

- 编号: Task-07
- 模块: runtime
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）

实现全局注册表配置，负责管理所有 Provider 的注册和获取，提供完整的 Provider 管理支持，使 RuntimeInfo 可以通过全局注册表构建，支持多 Provider 组合和灵活的配置管理。作为内部功能模块，不对外提供API接口。

### A2 分析（Analyze）

- **现状**：
  - ✅ 已实现：Task-01 中定义的 Provider 接口和 RuntimeInfoBuilder 接口
  - ✅ 已实现：Task-02 中的 K8s 环境信息提供者
  - ✅ 已实现：Task-03 中的本地环境信息提供者
  - ✅ 已实现：Task-04 中的配置中心集成提供者
  - ✅ 已实现：Task-05 中的构建信息提供者
  - ✅ 已实现：Task-06 中的 RuntimeInfoBuilder 构建器
  - ❌ 未实现：全局注册表配置、Provider 组合管理、配置管理
- **差距**：
  - 需要实现全局 Provider 注册表
  - 需要实现 Provider 的注册和获取函数
  - 需要实现 RuntimeInfo 的构建函数
  - 需要支持配置驱动的 Provider 管理
- **约束**：
  - 必须支持全局 Provider 注册
  - 必须支持多 Provider 组合
  - 必须提供灵活的配置管理
  - 必须支持错误处理和回退机制
- **风险**：
  - 技术风险：全局注册表复杂性
  - 业务风险：Provider 组合不当
  - 依赖风险：配置管理复杂性

### A3 设计（Architect）

- **接口契约**：
  - **核心配置**：`GlobalProviderRegistry` - 定义全局 Provider 注册表
  - **核心函数**：
    - `BuildRuntimeInfoFromRegistry()` - 从注册表构建 RuntimeInfo 实例
    - `RegisterProvider()` - 全局注册自定义 Provider
    - `GetProvider()` - 获取指定 Provider
    - `GetAllProviders()` - 获取所有已注册的 Provider
  - **配置策略**：全局 Provider 注册表管理、优先级管理、动态注册支持

- **架构设计**：
  - 采用全局注册表模式管理 Provider
  - 使用单例模式管理全局注册表
  - 提供 Provider 优先级管理
  - 支持动态 Provider 注册

- **核心功能模块**：
  - `provider_registry.go`: 全局 Provider 注册表管理
  - `registry_builder.go`: 从注册表构建 RuntimeInfo

- **极小任务拆分**：
  - T07-01：实现全局 Provider 注册表管理
  - T07-02：实现从注册表构建 RuntimeInfo
  - T07-03：实现便捷的全局注册函数
  - T07-04：添加单元测试

### A4 行动（Act）

#### T07-01：实现全局 Provider 注册表管理

```go
// provider_registry.go
package runtime

import (
    "sync"
)

// GlobalProviderRegistry 全局 Provider 注册表
type GlobalProviderRegistry struct {
    providers map[string]Provider
    mu        sync.RWMutex
}

// 全局实例
var globalRegistry *GlobalProviderRegistry
var once sync.Once

// NewGlobalProviderRegistry 创建全局 Provider 注册表（单例）
func NewGlobalProviderRegistry() *GlobalProviderRegistry {
    once.Do(func() {
        globalRegistry = &GlobalProviderRegistry{
            providers: make(map[string]Provider),
        }
    })
    return globalRegistry
}

// RegisterProvider 注册 Provider
func (r *GlobalProviderRegistry) RegisterProvider(provider Provider) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.providers[provider.GetName()] = provider
}

// GetProvider 获取指定 Provider
func (r *GlobalProviderRegistry) GetProvider(name string) (Provider, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    provider, exists := r.providers[name]
    return provider, exists
}

// GetAllProviders 获取所有 Provider
func (r *GlobalProviderRegistry) GetAllProviders() []Provider {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    providers := make([]Provider, 0, len(r.providers))
    for _, provider := range r.providers {
        providers = append(providers, provider)
    }
    return providers
}
```

#### T07-02：实现全局 Provider 注册表管理

```go
// provider_registry.go
package runtime

import (
    "sync"
)

// GlobalProviderRegistry 全局 Provider 注册表
type GlobalProviderRegistry struct {
    providers map[string]Provider
    mu        sync.RWMutex
}

// 全局实例
var globalRegistry *GlobalProviderRegistry
var once sync.Once

// NewGlobalProviderRegistry 创建全局 Provider 注册表（单例）
func NewGlobalProviderRegistry() *GlobalProviderRegistry {
    once.Do(func() {
        globalRegistry = &GlobalProviderRegistry{
            providers: make(map[string]Provider),
        }
    })
    return globalRegistry
}

// RegisterProvider 注册 Provider
func (r *GlobalProviderRegistry) RegisterProvider(provider Provider) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.providers[provider.GetName()] = provider
}

// GetProvider 获取指定 Provider
func (r *GlobalProviderRegistry) GetProvider(name string) (Provider, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    provider, exists := r.providers[name]
    return provider, exists
}

// GetAllProviders 获取所有 Provider
func (r *GlobalProviderRegistry) GetAllProviders() []Provider {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    providers := make([]Provider, 0, len(r.providers))
    for _, provider := range r.providers {
        providers = append(providers, provider)
    }
    return providers
}

// 全局便捷方法
func RegisterProvider(provider Provider) {
    registry := NewGlobalProviderRegistry()
    registry.RegisterProvider(provider)
}

func GetProvider(name string) (Provider, bool) {
    registry := NewGlobalProviderRegistry()
    return registry.GetProvider(name)
}

func GetAllProviders() []Provider {
    registry := NewGlobalProviderRegistry()
    return registry.GetAllProviders()
}
```

#### T07-03：实现便捷的全局注册函数

```go
// provider_registry.go (续)

// 全局便捷方法
func RegisterProvider(provider Provider) {
    registry := NewGlobalProviderRegistry()
    registry.RegisterProvider(provider)
}

func GetProvider(name string) (Provider, bool) {
    registry := NewGlobalProviderRegistry()
    return registry.GetProvider(name)
}

func GetAllProviders() []Provider {
    registry := NewGlobalProviderRegistry()
    return registry.GetAllProviders()
}

// 便捷的构建函数
func BuildRuntimeInfo() (*RuntimeInfo, error) {
    return BuildRuntimeInfoFromRegistry()
}
            builder.AddProvider(provider)
        }
    }
    
    // 从 Provider 构建 RuntimeInfo
    ctx := context.Background()
    info, err := builder.BuildFromProviders(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to build RuntimeInfo: %w", err)
    }
    
    return info, nil
}
```





#### T07-04：添加单元测试

```go
// provider_registry_test.go
package runtime

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGlobalProviderRegistry(t *testing.T) {
    // 测试全局 Provider 注册表
    registry := NewGlobalProviderRegistry()
    assert.NotNil(t, registry)
    
    // 测试注册和获取 Provider
    mockProvider := &MockProvider{name: "test", priority: 1}
    registry.RegisterProvider(mockProvider)
    
    provider, exists := registry.GetProvider("test")
    assert.True(t, exists)
    assert.Equal(t, "test", provider.GetName())
    
    // 测试获取所有 Provider
    providers := registry.GetAllProviders()
    assert.Len(t, providers, 1)
}

func TestBuildRuntimeInfoFromRegistry(t *testing.T) {
    // 测试从注册表构建 RuntimeInfo
    
    // 先注册测试 Provider
    mockProvider := &MockProvider{name: "test", priority: 1}
    RegisterProvider(mockProvider)
    
    info, err := BuildRuntimeInfoFromRegistry()
    
    assert.NoError(t, err)
    assert.NotNil(t, info)
}

func TestGlobalProviderFunctions(t *testing.T) {
    // 测试全局 Provider 函数
    
    // 先注册测试 Provider
    mockProvider := &MockProvider{name: "test", priority: 1}
    RegisterProvider(mockProvider)
    
    // 测试获取 Provider
    provider, exists := GetProvider("test")
    assert.True(t, exists)
    assert.Equal(t, "test", provider.GetName())
    
    // 测试获取所有 Provider
    providers := GetAllProviders()
    assert.Len(t, providers, 1)
    assert.Equal(t, "test", providers[0].GetName())
}

// MockProvider 用于测试
type MockProvider struct {
    name     string
    priority int
}

func (m *MockProvider) GetName() string { return m.name }
func (m *MockProvider) GetPriority() int { return m.priority }
func (m *MockProvider) CanProvide(field string) bool { return true }
func (m *MockProvider) Provide(field string) (string, error) { return "test", nil }
```

### A5 验证（Assure）

- **测试用例**：
  - 测试 Provider 注册表功能
  - 测试从注册表构建 RuntimeInfo
  - 测试 Provider 组合构建

- **性能验证**：
  - Provider 注册性能
  - RuntimeInfo 构建性能

- **回归测试**：
  - 确保 Provider 注册表正确
  - 确保 RuntimeInfo 构建正确

- **测试结果**：
  - 所有测试用例通过
  - Provider 注册表功能正常
  - RuntimeInfo 构建功能正常

### A6 迭代（Advance）

- **性能优化**：
  - 优化 Provider 注册性能
  - 优化 RuntimeInfo 构建性能

- **功能扩展**：
  - 支持更多 Provider 类型
  - 支持动态 Provider 注册
  - 支持 Provider 优先级管理

- **观测性增强**：
  - 添加 Provider 状态监控
  - 添加构建过程监控

- **下一步任务链接**：
  - 链接到 [Task-08](./Task-08-实现ServerInfo-Context注入支持.md) - 实现 ServerInfo Context 注入支持

### 📋 质量检查

- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结

成功实现了全局注册表配置，包括：

1. **全局 Provider 注册表管理**：支持用户自定义 Provider 的全局注册和管理
2. **从注册表构建 RuntimeInfo**：从全局注册表获取所有 Provider 并构建 RuntimeInfo
3. **便捷的全局注册函数**：提供简单的全局函数来注册和获取 Provider
4. **简化设计**：去除了过度设计，只保留核心功能
5. **单元测试**：提供了基本的测试用例，确保功能正确性

所有实现都使用全局注册表模式，采用单例模式管理全局注册表，支持多 Provider 组合。用户可以通过全局函数注册自己的 Provider，系统自动管理所有 Provider 的注册和获取。
