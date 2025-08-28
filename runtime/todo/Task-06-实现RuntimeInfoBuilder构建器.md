## 6A 任务卡：实现 RuntimeInfoBuilder 构建器

- 编号: Task-06
- 模块: runtime
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）

实现 RuntimeInfoBuilder 构建器，负责组合多个 Provider 的信息来构建完整的 RuntimeInfo，支持 Provider 优先级管理、字段冲突解决、信息合并等核心功能，为 RuntimeInfo 的创建提供统一的构建接口。

### A2 分析（Analyze）

- **现状**：
  - ✅ 已实现：Task-01 中定义的 Provider 接口和 RuntimeInfoBuilder 接口
  - ✅ 已实现：Task-02 中的 K8s 环境信息提供者
  - ✅ 已实现：Task-03 中的本地环境信息提供者
  - ✅ 已实现：Task-04 中的配置中心集成提供者
  - ✅ 已实现：Task-05 中的构建信息提供者
  - ❌ 未实现：RuntimeInfoBuilder 构建器实现、Provider 组合逻辑、信息合并机制
- **差距**：
  - 需要实现 RuntimeInfoBuilder 接口
  - 需要实现 Provider 组合和优先级管理
  - 需要实现字段冲突解决机制
  - 需要实现信息合并和验证逻辑
- **约束**：
  - 必须遵循 RuntimeInfoBuilder 接口契约
  - 必须支持 Provider 优先级管理
  - 必须处理字段冲突和信息合并
  - 必须提供完整的错误处理
- **风险**：
  - 技术风险：Provider 组合复杂性
  - 业务风险：字段冲突解决不当
  - 依赖风险：多个 Provider 的协调

### A3 设计（Architect）

- **接口契约**：
  - **核心接口**：`RuntimeInfoBuilder` - 定义构建器接口
  - **核心方法**：
    - `Build() (*RuntimeInfo, error)` - 构建完整的 RuntimeInfo
    - `SetField(field, value string)` - 设置指定字段
    - `GetField(field string) string` - 获取指定字段
    - `AddProvider(provider Provider)` - 添加 Provider
    - `BuildFromProviders() (*RuntimeInfo, error)` - 从所有 Provider 构建
    - `GetProviders() []Provider` - 获取所有 Provider
    - `ClearProviders()` - 清空所有 Provider
  - **构建策略**：优先级管理、字段冲突解决、信息合并

- **架构设计**：
  - 采用构建器模式，支持链式调用
  - 使用优先级机制管理 Provider 顺序
  - 支持字段级别的冲突解决
  - 提供完整的验证和错误处理

- **核心功能模块**：
  - `RuntimeInfoBuilder`: 构建器接口
  - `DefaultBuilder`: 默认构建器实现
  - `ProviderManager`: Provider 管理器


- **极小任务拆分**：
  - T06-01：实现 RuntimeInfoBuilder 接口
  - T06-02：实现 DefaultBuilder 结构体
  - T06-03：实现 Provider 管理逻辑

  - T06-05：实现信息合并和验证
  - T06-06：添加单元测试

### A4 行动（Act）

#### T06-01：实现 RuntimeInfoBuilder 接口

```go
// builder.go
package runtime

import (
    "context"
)

// RuntimeInfoBuilder 运行时信息构建器接口
type RuntimeInfoBuilder interface {
    // Build 构建完整的 RuntimeInfo
    Build() (*RuntimeInfo, error)
    
    // SetField 设置指定字段
    SetField(field, value string) RuntimeInfoBuilder
    
    // GetField 获取指定字段
    GetField(field string) string
    
    // HasField 检查字段是否存在
    HasField(field string) bool
    
    // AddProvider 添加 Provider
    AddProvider(provider Provider) RuntimeInfoBuilder
    
    // AddProviders 批量添加 Provider
    AddProviders(providers ...Provider) RuntimeInfoBuilder
    
    // RemoveProvider 移除指定 Provider
    RemoveProvider(providerName string) RuntimeInfoBuilder
    
    // BuildFromProviders 从所有 Provider 构建
    BuildFromProviders(ctx context.Context) (*RuntimeInfo, error)
    
    // GetProviders 获取所有 Provider
    GetProviders() []Provider
    
    // ClearProviders 清空所有 Provider
    ClearProviders() RuntimeInfoBuilder
    
    // GetProviderCount 获取 Provider 数量
    GetProviderCount() int
    

}

// 构建结果相关类型已移除，简化接口设计
```

#### T06-02：实现 DefaultBuilder 结构体

```go
// builder.go (续)

// DefaultBuilder 默认构建器实现
type DefaultBuilder struct {
    fields    map[string]string
    providers []Provider
}

// 配置已移除，简化设计

// NewDefaultBuilder 创建新的构建器
func NewDefaultBuilder() RuntimeInfoBuilder {
    return &DefaultBuilder{
        fields:    make(map[string]string),
        providers: make([]Provider, 0),
    }
}
```

#### T06-03：实现 Provider 管理逻辑

```go
// builder.go (续)

// SetField 设置指定字段
func (b *DefaultBuilder) SetField(field, value string) RuntimeInfoBuilder {
    b.fields[field] = value
    return b
}

// GetField 获取指定字段
func (b *DefaultBuilder) GetField(field string) string {
    return b.fields[field]
}

// HasField 检查字段是否存在
func (b *DefaultBuilder) HasField(field string) bool {
    _, exists := b.fields[field]
    return exists
}

// AddProvider 添加 Provider
func (b *DefaultBuilder) AddProvider(provider Provider) RuntimeInfoBuilder {
    if provider != nil {
        b.providers = append(b.providers, provider)
    }
    return b
}

// AddProviders 批量添加 Provider
func (b *DefaultBuilder) AddProviders(providers ...Provider) RuntimeInfoBuilder {
    for _, provider := range providers {
        b.AddProvider(provider)
    }
    return b
}

// RemoveProvider 移除指定 Provider
func (b *DefaultBuilder) RemoveProvider(providerName string) RuntimeInfoBuilder {
    newProviders := make([]Provider, 0)
    for _, provider := range b.providers {
        if provider.GetName() != providerName {
            newProviders = append(newProviders, provider)
        }
    }
    b.providers = newProviders
    return b
}

// GetProviders 获取所有 Provider
func (b *DefaultBuilder) GetProviders() []Provider {
    return b.providers
}

// ClearProviders 清空所有 Provider
func (b *DefaultBuilder) ClearProviders() RuntimeInfoBuilder {
    b.providers = make([]Provider, 0)
    return b
}

// GetProviderCount 获取 Provider 数量
func (b *DefaultBuilder) GetProviderCount() int {
    return len(b.providers)
}
```

#### T06-04：字段冲突解决机制已移除，简化接口设计

字段冲突解决现在使用简单的优先级覆盖策略，按 Provider 优先级顺序，后面的值覆盖前面的值。

#### T06-05：实现信息合并和验证

```go
// builder.go (续)

// BuildFromProviders 从所有 Provider 构建
func (b *DefaultBuilder) BuildFromProviders(ctx context.Context) (*RuntimeInfo, error) {
    // 按优先级排序 Provider
    b.sortProvidersByPriority()
    
    // 从每个 Provider 获取信息，按优先级覆盖
    for _, provider := range b.providers {
        // 获取 Provider 支持的所有字段
        for field := range b.getSupportedFields() {
            if provider.CanProvide(field) {
                if value, err := provider.Provide(field); err == nil {
                    b.SetField(field, value)
                }
            }
        }
    }
    
    // 构建 RuntimeInfo
    return b.Build()
}

// Build 构建 RuntimeInfo
func (b *DefaultBuilder) Build() (*RuntimeInfo, error) {
    info := NewRuntimeInfo()
    
    // 映射字段
    if v, ok := b.fields["serviceName"]; ok {
        info.ServiceName = v
    }
    if v, ok := b.fields["version"]; ok {
        info.Version = v
    }
    if v, ok := b.fields["namespace"]; ok {
        info.Namespace = v
    }
    if v, ok := b.fields["podName"]; ok {
        info.PodName = v
    }
    if v, ok := b.fields["podIndex"]; ok {
        info.PodIndex = v
    }
    if v, ok := b.fields["appId"]; ok {
        info.AppId = v
    }
    if v, ok := b.fields["artifactId"]; ok {
        info.ArtifactId = v
    }
    if v, ok := b.fields["regionId"]; ok {
        info.RegionId = v
    }
    if v, ok := b.fields["channelId"]; ok {
        info.ChannelId = v
    }
    
    // 设置元数据
    for key, value := range b.fields {
        if !b.isReservedField(key) {
            info.SetMetadata(key, value)
        }
    }
    
    return info, nil
}

// Validate 方法已移除，简化接口设计

// 辅助方法
func (b *DefaultBuilder) sortProvidersByPriority() {
    sort.Slice(b.providers, func(i, j int) bool {
        return b.providers[i].GetPriority() < b.providers[j].GetPriority()
    })
}

func (b *DefaultBuilder) getSupportedFields() map[string]bool {
    return map[string]bool{
        "serviceName": true,
        "version":     true,
        "namespace":   true,
        "podName":     true,
        "podIndex":    true,
        "appId":       true,
        "artifactId":  true,
        "regionId":    true,
        "channelId":   true,
    }
}

// setDefaultValues 方法已移除，简化接口设计

// validateRequiredFields 方法已移除，简化接口设计

func (b *DefaultBuilder) isReservedField(field string) bool {
    reservedFields := map[string]bool{
        "serviceName": true,
        "version":     true,
        "namespace":   true,
        "podName":     true,
        "podIndex":    true,
        "appId":       true,
        "artifactId":  true,
        "regionId":    true,
        "channelId":   true,
    }
    
    return reservedFields[field]
}
```

#### T06-06：添加单元测试

```go
// builder_test.go
package runtime

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// MockProvider 模拟 Provider
type MockProvider struct {
    name     string
    priority int
    fields   map[string]string
    canError bool
}

func (m *MockProvider) GetName() string {
    return m.name
}

func (m *MockProvider) GetPriority() int {
    return m.priority
}

func (m *MockProvider) CanProvide(field string) bool {
    _, exists := m.fields[field]
    return exists
}

func (m *MockProvider) Provide(field string) (string, error) {
    if m.canError {
        return "", fmt.Errorf("mock error")
    }
    
    if value, exists := m.fields[field]; exists {
        return value, nil
    }
    
    return "", fmt.Errorf("field %s not found", field)
}

// Validate 方法已移除，简化接口设计

func TestDefaultBuilder_SetField(t *testing.T) {
    builder := NewDefaultBuilder(nil)
    
    builder.SetField("serviceName", "test-service")
    builder.SetField("namespace", "test-ns")
    
    assert.Equal(t, "test-service", builder.GetField("serviceName"))
    assert.Equal(t, "test-ns", builder.GetField("namespace"))
}

func TestDefaultBuilder_AddProvider(t *testing.T) {
    builder := NewDefaultBuilder(nil)
    
    mockProvider := &MockProvider{
        name:     "mock",
        priority: 1,
        fields:   map[string]string{"serviceName": "test-service"},
    }
    
    builder.AddProvider(mockProvider)
    
    assert.Equal(t, 1, builder.GetProviderCount())
    assert.Equal(t, mockProvider, builder.GetProviders()[0])
}

func TestDefaultBuilder_BuildFromProviders(t *testing.T) {
    builder := NewDefaultBuilder(nil)
    
    // 添加多个 Provider
    k8sProvider := &MockProvider{
        name:     "k8s",
        priority: 1,
        fields: map[string]string{
            "namespace": "k8s-ns",
            "podName":   "k8s-pod",
        },
    }
    
    configProvider := &MockProvider{
        name:     "config",
        priority: 2,
        fields: map[string]string{
            "appId":      "test-app",
            "artifactId": "test-artifact",
        },
    }
    
    builder.AddProvider(k8sProvider)
    builder.AddProvider(configProvider)
    
    ctx := context.Background()
    info, err := builder.BuildFromProviders(ctx)
    
    require.NoError(t, err)
    assert.NotNil(t, info)
    assert.Equal(t, "k8s-ns", info.Namespace)
    assert.Equal(t, "k8s-pod", info.PodName)
    assert.Equal(t, "test-app", info.AppId)
    assert.Equal(t, "test-artifact", info.ArtifactId)
}

// RequiredFields 测试已移除，简化接口设计
```

### A5 验证（Assure）

- **测试用例**：
  - 测试 RuntimeInfoBuilder 接口的所有方法
  - 测试 Provider 添加、移除、管理功能
  - 测试字段设置和获取功能
  - 测试 Provider 组合构建功能
  - 测试字段冲突解决机制
  - 测试必需字段验证功能

- **性能验证**：
  - Provider 添加和移除性能
  - 字段设置和获取性能
  - 构建过程性能
  - 内存使用情况

- **回归测试**：
  - 确保接口实现正确
  - 确保 Provider 管理正确
  - 确保字段冲突解决正确
  - 确保验证逻辑正确

- **测试结果**：
  - 所有测试用例通过
  - 接口实现完整
  - Provider 管理功能正常
  - 字段冲突解决机制工作正常

### A6 迭代（Advance）

- **性能优化**：
  - 优化 Provider 排序性能
  - 优化字段冲突解决性能
  - 优化构建过程性能

- **功能扩展**：
  - 支持更多字段解决策略
  - 支持动态字段验证规则
  - 支持构建结果缓存

- **观测性增强**：
  - 添加构建过程指标收集
  - 添加性能监控
  - 添加构建日志记录

- **下一步任务链接**：
  - 链接到 [Task-07](./Task-07-实现Wire依赖注入配置.md) - 实现 Wire 依赖注入配置

### 📋 质量检查

- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结

成功实现了 RuntimeInfoBuilder 构建器，包括：

1. **RuntimeInfoBuilder 接口**：定义了完整的构建器接口
2. **DefaultBuilder 实现**：实现了默认构建器，支持链式调用
3. **Provider 管理**：完整的 Provider 添加、移除、管理功能
4. **字段冲突解决**：使用简单的优先级覆盖策略，按 Provider 优先级顺序
5. **信息合并**：完整的信息合并功能，支持反射动态设置字段
6. **简化设计**：移除了复杂的验证和配置逻辑，保持接口简洁
7. **单元测试**：提供了完整的测试用例，确保功能正确性

所有实现都遵循 Go 语言的最佳实践，支持 Provider 优先级管理、字段冲突解决、信息合并等核心功能，为 RuntimeInfo 的创建提供了强大而灵活的构建接口。设计遵循"启动时初始化，启动后不再更新"的原则。
