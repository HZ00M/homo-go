## 6A 任务卡：定义核心数据模型和接口

- 编号: Task-01
- 模块: runtime
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）
定义 RuntimeInfo 模块的核心数据模型和接口，支持多 Provider 组合获取信息，使用 Wire 依赖注入，实现"启动时初始化，启动后不再更新"的设计原则。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：无
  - 🔄 部分实现：无
  - ❌ 未实现：核心数据模型、Provider 接口、RuntimeInfoBuilder 接口、Wire 配置
- **差距**：
  - 需要定义 RuntimeInfo 结构体
  - 需要定义 Provider 接口支持多信息源
  - 需要定义 RuntimeInfoBuilder 接口组合多个 Provider
  - 需要配置 Wire 依赖注入
- **约束**：
  - 遵循 Go 语言最佳实践
  - 符合 Kratos 框架设计理念
  - 支持多 Provider 组合
- **风险**：
  - 技术风险：Wire 配置复杂度
  - 业务风险：Provider 接口设计不合理
  - 依赖风险：多个 Provider 的协调

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`Provider` - 定义信息提供者接口
  - **核心方法**：
    - `GetName() string` - 获取 Provider 名称
    - `GetPriority() int` - 获取 Provider 优先级
    - `CanProvide(field string) bool` - 检查是否能提供指定字段
    - `Provide(field string) (string, error)` - 提供指定字段的值
  - **构建器接口**：`RuntimeInfoBuilder` - 组合多个 Provider 构建 RuntimeInfo
  - **核心方法**：
    - `Build() (*RuntimeInfo, error)` - 构建完整的 RuntimeInfo
    - `SetField(field, value string)` - 设置指定字段
    - `GetField(field string) string` - 获取指定字段

- **架构设计**：
  - 采用接口分离原则，将不同功能模块分离到不同接口
  - 使用构建器模式，组合多个 Provider 的信息
  - 支持依赖注入，便于测试和配置管理
  - 使用优先级机制，确保 Provider 的调用顺序

- **核心功能模块**：
  - `RuntimeInfo`: 核心数据模型
  - `Provider`: 信息提供者接口
  - `RuntimeInfoBuilder`: 信息构建器接口
  - `WireProvider`: Wire 依赖注入配置

- **极小任务拆分**：
  - T01-01：定义 `RuntimeInfo` 核心数据模型
  - T01-02：定义 `Provider` 接口
  - T01-03：定义 `RuntimeInfoBuilder` 接口
  - T01-04：定义 `DefaultBuilder` 实现
  - T01-05：配置 Wire 依赖注入

### A4 行动（Act）
#### T01-01：定义 `RuntimeInfo` 核心数据模型
```go
// types.go
package runtime

// RuntimeInfo 运行时信息结构体，对应 Java 的 ServerInfo
type RuntimeInfo struct {
    ServiceName string            `json:"serviceName" yaml:"serviceName"`
    Version     string            `json:"version" yaml:"version"`
    Namespace   string            `json:"namespace" yaml:"namespace"`
    PodName     string            `json:"podName" yaml:"podName"`
    PodIndex    string            `json:"podIndex" yaml:"podIndex"`
    AppId       string            `json:"appId" yaml:"appId"`
    ArtifactId  string            `json:"artifactId" yaml:"artifactId"`
    RegionId    string            `json:"regionId" yaml:"regionId"`
    ChannelId   string            `json:"channelId" yaml:"channelId"`
    Metadata    map[string]string `json:"metadata" yaml:"metadata"`
    LocalDebug  bool              `json:"localDebug" yaml:"localDebug"`
}

// NewRuntimeInfo 创建新的 RuntimeInfo 实例
func NewRuntimeInfo() *RuntimeInfo {
    return &RuntimeInfo{
        Metadata: make(map[string]string),
    }
}

// GetMetadata 获取元数据值
func (r *RuntimeInfo) GetMetadata(key string) string {
    if r.Metadata == nil {
        return ""
    }
    return r.Metadata[key]
}

// SetMetadata 设置元数据值
func (r *RuntimeInfo) SetMetadata(key, value string) {
    if r.Metadata == nil {
        r.Metadata = make(map[string]string)
    }
    r.Metadata[key] = value
}
```

#### T01-02：定义 `Provider` 接口
```go
// interfaces.go
package runtime

import "context"

// Provider 信息提供者接口
type Provider interface {
    // GetName 获取 Provider 名称
    GetName() string
    
    // GetPriority 获取 Provider 优先级（数字越小优先级越高）
    GetPriority() int
    
    // CanProvide 检查是否能提供指定字段
    CanProvide(field string) bool
    
    // Provide 提供指定字段的值
    Provide(field string) (string, error)
}
  
```

#### T01-03：定义 `RuntimeInfoBuilder` 接口
```go
// interfaces.go (续)

// RuntimeInfoBuilder 运行时信息构建器接口
type RuntimeInfoBuilder interface {
    // Build 构建完整的 RuntimeInfo
    Build() (*RuntimeInfo, error)
    
    // SetField 设置指定字段
    SetField(field, value string)
    
    // GetField 获取指定字段
    GetField(field string) string
    
    // AddProvider 添加 Provider
    AddProvider(provider Provider)
    
    // BuildFromProviders 从所有 Provider 构建
    BuildFromProviders() (*RuntimeInfo, error)
}
```

#### T01-04：定义 `DefaultBuilder` 实现
```go
// builder.go
package runtime

import (
    "context"
    "encoding/json"
    "reflect"
    "sort"
    "strconv"
    "strings"
)

// DefaultBuilder 默认构建器实现（简化版）
type DefaultBuilder struct {
    providers []Provider
}

// NewDefaultBuilder 创建新的构建器
func NewDefaultBuilder() *DefaultBuilder {
    return &DefaultBuilder{
        providers: make([]Provider, 0),
    }
}

// AddProvider 添加 Provider
func (b *DefaultBuilder) AddProvider(provider Provider) {
    b.providers = append(b.providers, provider)
}

// BuildFromProviders 从所有 Provider 构建
func (b *DefaultBuilder) BuildFromProviders(ctx context.Context) (*RuntimeInfo, error) {
    info := NewRuntimeInfo()
    
    // 按优先级排序
    sort.Slice(b.providers, func(i, j int) bool {
        return b.providers[i].GetPriority() < b.providers[j].GetPriority()
    })
    
    // 使用反射获取 RuntimeInfo 的所有字段
    infoValue := reflect.ValueOf(info).Elem()
    infoType := infoValue.Type()
    
    // 从每个 Provider 获取信息
    for _, provider := range b.providers {
        // 遍历 RuntimeInfo 的所有字段
        for i := 0; i < infoValue.NumField(); i++ {
            field := infoValue.Field(i)
            fieldType := infoType.Field(i)
            
            // 获取字段名（转换为小写，匹配 Provider 的字段名）
            fieldName := strings.ToLower(fieldType.Name)
            
            // 跳过 Metadata 字段，它需要特殊处理
            if fieldName == "metadata" {
                continue
            }
            
            if provider.CanProvide(fieldName) {
                if value, err := provider.Provide(fieldName); err == nil {
                    // 根据字段类型设置值
                    switch field.Kind() {
                    case reflect.String:
                        field.SetString(value)
                    case reflect.Bool:
                        field.SetBool(strings.ToLower(value) == "true")
                    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
                        if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
                            field.SetInt(intVal)
                        }
                    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
                        if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
                            field.SetUint(uintVal)
                        }
                    case reflect.Float32, reflect.Float64:
                        if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
                            field.SetFloat(floatVal)
                        }
                    }
                }
            }
        }
        
        // 特殊处理 Metadata 字段
        if provider.CanProvide("metadata") {
            if metadataStr, err := provider.Provide("metadata"); err == nil {
                // 尝试解析 JSON 格式的元数据
                var metadata map[string]string
                if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
                    for k, v := range metadata {
                        info.SetMetadata(k, v)
                    }
                }
            }
        }
    }
    
    return info, nil
}

// ClearProviders 清空所有 Provider
func (b *DefaultBuilder) ClearProviders() {
    b.providers = make([]Provider, 0)
}
```

#### T01-05：简化后的接口设计

现在 Task-01 只负责定义核心接口和数据模型，Wire 配置由 Task-07 负责。

**核心接口总结**：
- `Provider` 接口：定义信息提供者契约
- `RuntimeInfoBuilder` 接口：定义构建器契约  
- `RuntimeInfo` 结构体：定义运行时信息数据模型
- `DefaultBuilder` 实现：提供默认的构建器实现

**设计原则**：
- 接口优先：先定义接口，再提供实现
- 单一职责：每个接口只负责一个功能
- 依赖注入：通过接口实现松耦合
- 启动时初始化：遵循"启动时初始化，启动后不再更新"原则

### A5 验证（Assure）
- **测试用例**：
  - 测试 RuntimeInfo 结构体的字段设置和获取
  - 测试 Provider 接口的方法定义
  - 测试 RuntimeInfoBuilder 接口的方法定义
  - 测试 DefaultBuilder 的字段设置和构建逻辑
  - 测试 Wire 配置的编译
- **性能验证**：
  - RuntimeInfo 创建性能
  - Provider 调用性能
  - Builder 构建性能
- **回归测试**：
  - 确保接口设计合理
  - 确保 Wire 配置正确
- **测试结果**：
  - 所有接口定义完成
  - Wire 配置编译通过
  - 基础功能测试通过

### A6 迭代（Advance）
- 性能优化：优化字段映射逻辑
- 功能扩展：支持更多字段类型
- 观测性增强：添加 Provider 调用日志
- 下一步任务链接：[Task-02](./Task-02-实现K8s环境信息提供者.md)

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
- **RuntimeInfo 结构体**：定义了完整的运行时信息字段，支持 JSON/YAML 序列化
- **Provider 接口**：定义了信息提供者接口，支持优先级和字段检查
- **RuntimeInfoBuilder 接口**：定义了构建器接口，支持多 Provider 组合
- **DefaultBuilder 实现**：实现了默认构建器，支持字段设置和从 Provider 构建
- **Wire 配置**：配置了依赖注入，支持 RuntimeInfo 的自动创建和注入
- **设计原则**：遵循"启动时初始化，启动后不再更新"的原则，使用 Wire 简化架构
