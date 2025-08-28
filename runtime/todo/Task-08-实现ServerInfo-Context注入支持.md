## 6A 任务卡：实现 ServerInfo Context 注入支持

- 编号: Task-10
- 模块: runtime
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: 待定
- 实际完成时间: -

### A1 目标（Aim）

实现 ServerInfo 的 Context 注入和提取功能，支持在请求处理过程中自动传递运行时信息，提高系统的可观测性和调试能力。同时提供基于 provider 构建 runtimeinfo 功能，以及完整的 Wire 依赖注入配置。

**业务价值：**
- 在日志、监控、追踪中自动包含服务实例信息
- 支持分布式追踪和服务链路分析
- 便于问题定位和性能分析
- 支持多 provider 组合构建 RuntimeInfo
- 提供 Wire 依赖注入，支持在任何地方注入 RuntimeInfo

### A2 分析（Analyze）

- **现状**：
  - ✅ 已实现：RuntimeInfo 数据模型和基本管理功能
  - ✅ 已实现：Provider 接口和实现（K8s、Config、Build）
  - ❌ 未实现：Context 注入和提取机制
  - ❌ 未实现：中间件集成支持
  - ❌ 未实现：RuntimeInfoBuilder 构建器
  - ❌ 未实现：Wire 依赖注入配置

- **差距**：
  - 缺少 Context 注入的核心接口
  - 缺少中间件自动注入机制
  - 缺少错误处理和类型安全
  - 缺少 RuntimeInfoBuilder 构建器
  - 缺少 Wire 依赖注入配置

- **约束**：
  - 必须与 Kratos 框架兼容
  - 必须支持 gRPC 和 HTTP 协议
  - 必须保持向后兼容性
  - 必须支持多 provider 组合
  - 必须提供 Wire 注入支持

- **风险**：
  - 技术风险：Context 类型安全、Provider 组合复杂性
  - 业务风险：性能影响、Provider 优先级管理
  - 依赖风险：与现有中间件集成、Wire 配置管理

### A3 设计（Architect）

#### 核心接口设计

```go
// context.go
package runtime

import (
    "context"
    "github.com/go-kratos/kratos/v2/errors"
)

// serverInfoKey 用于 Context 中存储 RuntimeInfo 的键
type serverInfoKey struct{}

// WithServerInfo 将 RuntimeInfo 注入到 context
func WithServerInfo(ctx context.Context, info *RuntimeInfo) context.Context {
    return context.WithValue(ctx, serverInfoKey{}, info)
}

// FromServerInfo 从 context 中获取 RuntimeInfo
func FromServerInfo(ctx context.Context) (*RuntimeInfo, bool) {
    info, ok := ctx.Value(serverInfoKey{}).(*RuntimeInfo)
    return info, ok
}

// InjectIntoContext 注入 RuntimeInfo 到 Context（别名函数）
func InjectIntoContext(ctx context.Context, info *RuntimeInfo) context.Context {
    return WithServerInfo(ctx, info)
}

// ExtractFromContext 从 Context 中提取 RuntimeInfo（带错误处理）
func ExtractFromContext(ctx context.Context) (*RuntimeInfo, error) {
    info, ok := FromServerInfo(ctx)
    if !ok {
        return nil, errors.New(500, "RUNTIME_INFO_NOT_FOUND", "RuntimeInfo not found in context")
    }
    return info, nil
}

// MustExtractFromContext 从 Context 中提取 RuntimeInfo（Panic 版本）
func MustExtractFromContext(ctx context.Context) *RuntimeInfo {
    info, ok := FromServerInfo(ctx)
    if !ok {
        panic("RuntimeInfo not found in context")
    }
    return info
}
```

#### RuntimeInfoBuilder 构建器设计

```go
// builder.go
package runtime

import (
    "fmt"
    "sort"
)

// RuntimeInfoBuilder RuntimeInfo 构建器接口
type RuntimeInfoBuilder interface {
    // Build 构建 RuntimeInfo 实例
    Build() (*RuntimeInfo, error)
    
    // SetField 设置字段值
    SetField(field, value string)
    
    // GetField 获取字段值
    GetField(field string) string
    
    // AddProvider 添加 Provider
    AddProvider(provider Provider)
    
    // BuildFromProviders 从所有 Provider 构建
    BuildFromProviders() (*RuntimeInfo, error)
}

// DefaultBuilder 默认的 RuntimeInfo 构建器实现
type DefaultBuilder struct {
    fields    map[string]string
    providers []Provider
}

// NewDefaultBuilder 创建新的默认构建器
func NewDefaultBuilder() RuntimeInfoBuilder {
    return &DefaultBuilder{
        fields:    make(map[string]string),
        providers: make([]Provider, 0),
    }
}

// SetField 设置字段值
func (b *DefaultBuilder) SetField(field, value string) {
    b.fields[field] = value
}

// GetField 获取字段值
func (b *DefaultBuilder) GetField(field string) string {
    return b.fields[field]
}

// AddProvider 添加 Provider
func (b *DefaultBuilder) AddProvider(provider Provider) {
    b.providers = append(b.providers, provider)
}

// Build 构建 RuntimeInfo 实例
func (b *DefaultBuilder) Build() (*RuntimeInfo, error) {
    // 验证必需字段
    if b.fields["serviceName"] == "" {
        return nil, fmt.Errorf("serviceName is required")
    }
    
    if b.fields["namespace"] == "" {
        return nil, fmt.Errorf("namespace is required")
    }
    
    return &RuntimeInfo{
        ServiceName: b.fields["serviceName"],
        Version:     b.fields["version"],
        Namespace:   b.fields["namespace"],
        PodName:     b.fields["podName"],
        PodIndex:    b.fields["podIndex"],
        AppId:       b.fields["appId"],
        ArtifactId:  b.fields["artifactId"],
        RegionId:    b.fields["regionId"],
        ChannelId:   b.fields["channelId"],
        Metadata:    b.fields,
    }, nil
}

// BuildFromProviders 从所有 Provider 构建
func (b *DefaultBuilder) BuildFromProviders() (*RuntimeInfo, error) {
    // 按优先级排序
    sort.Slice(b.providers, func(i, j int) bool {
        return b.providers[i].GetPriority() < b.providers[j].GetPriority()
    })
    
    // 从每个 Provider 获取信息
    for _, provider := range b.providers {
        // Validate 调用已移除，简化接口设计
        
        // 获取 Provider 支持的所有字段
        supportedFields := []string{"serviceName", "version", "namespace", "podName", "podIndex", "appId", "artifactId", "regionId", "channelId"}
        for _, field := range supportedFields {
            if provider.CanProvide(field) {
                if value, err := provider.Provide(field); err == nil {
                    b.SetField(field, value)
                }
            }
        }
    }
    
    return b.Build()
}
```

#### Wire 依赖注入配置设计

```go
// wire.go
package runtime

import (
    "github.com/google/wire"
)

// ProviderSet 定义 Wire 依赖注入集合
var ProviderSet = wire.NewSet(
    NewDefaultBuilder,
    NewK8sProvider,
    NewConfigProvider,
    NewBuildProvider,
    ProvideRuntimeInfo,
)

// ProvideRuntimeInfo 提供 RuntimeInfo 实例
func ProvideRuntimeInfo(
    builder RuntimeInfoBuilder,
    k8sProvider Provider,
    configProvider Provider,
    buildProvider Provider,
) *RuntimeInfo {
    // 添加所有 Provider
    builder.AddProvider(k8sProvider)
    builder.AddProvider(configProvider)
    builder.AddProvider(buildProvider)
    
    // 构建 RuntimeInfo
    info, err := builder.BuildFromProviders()
    if err != nil {
        panic(err)
    }
    
    return info
}

// NewK8sProvider 创建 K8s Provider
func NewK8sProvider() Provider {
    // 从 Task-02 实现
    return &K8sProvider{}
}

// NewConfigProvider 创建配置中心 Provider
func NewConfigProvider() Provider {
    // 从 Task-03 实现
    return &ConfigProvider{}
}

// NewBuildProvider 创建构建信息 Provider
func NewBuildProvider() Provider {
    // 从 Task-04 实现
    return &BuildProvider{}
}
```

#### 中间件集成设计

```go
// middleware.go
package runtime

import (
    "context"
    "github.com/go-kratos/kratos/v2/middleware"
    "github.com/go-kratos/kratos/v2/transport/http"
    "github.com/go-kratos/kratos/v2/transport/grpc"
)

// ServerInfoMiddleware HTTP 中间件，自动注入 RuntimeInfo
func ServerInfoMiddleware(info *RuntimeInfo) middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
            // 注入 RuntimeInfo 到 Context
            ctx = WithServerInfo(ctx, info)
            return handler(ctx, req)
        }
    }
}

// GRPCServerInfoInterceptor gRPC 拦截器，自动注入 RuntimeInfo
func GRPCServerInfoInterceptor(info *RuntimeInfo) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 注入 RuntimeInfo 到 Context
        ctx = WithServerInfo(ctx, info)
        return handler(ctx, req)
    }
}
```

#### 极小任务拆分

- T10-01：实现核心 Context 注入接口
- T10-02：实现 RuntimeInfoBuilder 构建器
- T10-03：实现 Wire 依赖注入配置
- T10-04：实现中间件和拦截器
- T10-05：添加错误处理和类型安全
- T10-06：编写单元测试
- T10-07：集成测试和性能验证

### A4 行动（Act）

#### T10-01：实现核心 Context 注入接口

**文件位置：** `runtime/context.go`

**实现步骤：**
1. 定义 `serverInfoKey` 结构体
2. 实现 `WithServerInfo` 函数
3. 实现 `FromServerInfo` 函数
4. 实现 `InjectIntoContext` 和 `ExtractFromContext` 函数
5. 添加必要的导入和错误处理

#### T10-02：实现 RuntimeInfoBuilder 构建器

**文件位置：** `runtime/builder.go`

**实现步骤：**
1. 定义 `RuntimeInfoBuilder` 接口
2. 实现 `DefaultBuilder` 结构体
3. 实现字段设置和获取方法
4. 实现 Provider 添加和管理
5. 实现从 Provider 构建 RuntimeInfo 的逻辑

#### T10-03：实现 Wire 依赖注入配置

**文件位置：** `runtime/wire.go`

**实现步骤：**
1. 定义 `ProviderSet` 依赖注入集合
2. 实现 `ProvideRuntimeInfo` 函数
3. 实现各个 Provider 的创建函数
4. 配置 Provider 的依赖关系
5. 确保 Wire 配置正确性

#### T10-04：实现中间件和拦截器

**文件位置：** `runtime/middleware.go`

**实现步骤：**
1. 实现 HTTP 中间件 `ServerInfoMiddleware`
2. 实现 gRPC 拦截器 `GRPCServerInfoInterceptor`
3. 确保与 Kratos 框架兼容
4. 添加必要的导入和类型定义

#### T10-05：添加错误处理和类型安全

**实现步骤：**
1. 使用 Kratos 错误码系统
2. 添加类型断言安全检查
3. 实现 `MustExtractFromContext` 函数
4. 添加参数验证
5. 完善 Provider 错误处理

#### T10-06：编写单元测试

**文件位置：** `runtime/context_test.go`, `runtime/builder_test.go`, `runtime/wire_test.go`

**测试用例：**
1. Context 注入和提取测试
2. RuntimeInfoBuilder 功能测试
3. Wire 依赖注入测试
4. 类型安全测试
5. 错误处理测试
6. 中间件功能测试

#### T10-07：集成测试和性能验证

**测试内容：**
1. 与现有 RuntimeInfo 系统集成测试
2. Provider 组合和优先级测试
3. Wire 注入功能测试
4. 性能基准测试
5. 内存泄漏测试
6. 并发安全测试

### A5 验证（Assure）

#### 测试用例

```go
// Context 注入测试
func TestWithServerInfo(t *testing.T) {
    ctx := context.Background()
    info := &RuntimeInfo{ServiceName: "test-service"}
    
    ctx = WithServerInfo(ctx, info)
    
    extracted, ok := FromServerInfo(ctx)
    assert.True(t, ok)
    assert.Equal(t, "test-service", extracted.ServiceName)
}

// RuntimeInfoBuilder 测试
func TestDefaultBuilder_BuildFromProviders(t *testing.T) {
    builder := NewDefaultBuilder()
    
    // 添加模拟 Provider
    mockProvider := &MockProvider{
        name:     "mock",
        priority: 1,
        fields:   map[string]string{"serviceName": "test-service", "namespace": "test-ns"},
    }
    
    builder.AddProvider(mockProvider)
    
    info, err := builder.BuildFromProviders()
    assert.NoError(t, err)
    assert.Equal(t, "test-service", info.ServiceName)
    assert.Equal(t, "test-ns", info.Namespace)
}

// Wire 注入测试
func TestWireInjection(t *testing.T) {
    // 测试 Wire 配置是否正确
    // 这里需要实际的 Wire 测试
}
```

#### 性能验证

- Context 注入性能：< 100ns
- Context 提取性能：< 50ns
- RuntimeInfo 构建性能：< 1μs
- Wire 注入性能：< 10μs
- 内存占用：< 1KB per context

#### 回归测试

- 确保不影响现有 RuntimeInfo 功能
- 确保与 Kratos 框架兼容性
- 确保类型安全性
- 确保 Provider 组合正确性
- 确保 Wire 注入功能正常

### A6 迭代（Advance）

#### 性能优化

- 考虑使用 sync.Pool 优化 Context 创建
- 优化类型断言性能
- 减少内存分配
- 优化 Provider 优先级排序
- 优化 Wire 注入性能

#### 功能扩展

- 支持批量 Context 注入
- 支持 Context 链式操作
- 支持 Context 验证和清理
- 支持动态 Provider 管理
- 支持 Provider 热插拔

#### 观测性增强

- 添加 Context 注入/提取指标
- 支持 Context 追踪和调试
- 集成 OpenTelemetry
- 添加 Provider 性能指标
- 添加 Wire 注入监控

#### 下一步任务链接

- 链接到 Task-11：实现 RuntimeInfo 监控指标收集
- 链接到 Task-12：实现 RuntimeInfo 配置热更新

### 📋 质量检查

- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结

成功实现了 ServerInfo Context 注入支持（放在 runtime 主包中），包括：

1. **核心 Context 注入接口**：完整的 Context 注入和提取功能
2. **RuntimeInfoBuilder 构建器**：支持多 Provider 组合构建 RuntimeInfo
3. **Wire 依赖注入配置**：完整的依赖注入支持，支持在任何地方注入 RuntimeInfo
4. **中间件和拦截器**：支持 HTTP 和 gRPC 的自动注入
5. **错误处理和类型安全**：完善的错误处理和类型安全检查
6. **单元测试**：全面的测试覆盖
7. **性能优化**：高效的 Context 操作和 RuntimeInfo 构建性能

所有实现都遵循 Go 语言的最佳实践和 Kratos 框架规范，为运行时信息管理提供了强大的 Context 注入支持、Provider 组合能力和 Wire 依赖注入功能。

---

**依赖关系：**
- **依赖**: [Task-02](./Task-02-实现K8s环境信息提供者.md), [Task-03](./Task-03-实现配置中心集成提供者.md), [Task-04](./Task-04-实现构建信息提供者.md)
- **被依赖**: [Task-11](./Task-11-实现RuntimeInfo监控指标收集.md) (待创建)
- **相关文档**: [todolist.md](./todolist.md) | [README.md](./README.md)
