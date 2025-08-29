# ServerInfo 模块

## 概述

ServerInfo 模块是一个轻量级的服务器信息管理模块，用于收集、管理和提供服务器的运行时信息。该模块采用 Provider 模式设计，支持多种信息源，并提供 Context 注入功能。

## 主要特性

- 🚀 **Provider 模式**：支持多种信息源（K8s、配置中心、构建信息、本地环境）
- 🔧 **注册表管理**：统一的 Provider 注册表，支持优先级排序
- 📝 **Context 注入**：完整的 Context 注入和提取功能
- 🧪 **测试友好**：完整的单元测试覆盖
- 🔌 **Wire 支持**：可选的 Google Wire 依赖注入支持

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/go-kratos/kratos/v2/serverinfo"
)

func main() {
    // 创建注册表
    registry := serverinfo.NewProviderRegistry()
    
    // 注册 Provider
    registry.RegisterProvider(serverinfo.NewK8sProvider())
    registry.RegisterProvider(serverinfo.NewConfigProvider())
    registry.RegisterProvider(serverinfo.NewBuildProvider())
    
    // 构建 ServerInfo
    info, err := registry.BuildServerInfo()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Service: %s, Version: %s\n", info.ServiceName, info.Version)
}
```

### Context 注入

```go
package main

import (
    "context"
    "github.com/go-kratos/kratos/v2/serverinfo"
)

func main() {
    // 创建 ServerInfo
    info := serverinfo.NewServerInfo()
    info.ServiceName = "my-service"
    info.Version = "1.0.0"
    
    // 注入到 Context
    ctx := serverinfo.WithServerInfo(context.Background(), info)
    
    // 从 Context 提取
    if extractedInfo, ok := serverinfo.FromServerInfo(ctx); ok {
        fmt.Printf("Service: %s\n", extractedInfo.ServiceName)
    }
}
```

### Wire 集成

```go
// wire.go
package main

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
    serverinfo.ProvideServerInfo,
)
```

## 架构设计

### 核心组件

1. **ServerInfo**：服务器信息的数据模型
2. **Provider**：信息提供者接口
3. **ProviderRegistry**：Provider 注册表
4. **Context**：Context 注入和提取功能

### Provider 类型

- **K8sProvider**：从 K8s 环境变量获取 Pod 信息
- **ConfigProvider**：从配置中心获取业务配置
- **BuildProvider**：从构建信息获取版本等
- **LocalProvider**：从本地环境获取信息

### 数据流

```
Provider 注册 → ProviderRegistry 管理 → 按优先级构建 RuntimeInfo → Context 注入
```

## API 参考

### ServerInfo

```go
type ServerInfo struct {
    ServiceName string            // 服务名称
    Version     string            // 版本
    Namespace   string            // 命名空间
    PodName     string            // Pod 名称
    PodIndex    string            // Pod 索引
    AppId       string            // 应用 ID
    ArtifactId  string            // 制品 ID
    RegionId    string            // 区域 ID
    ChannelId   string            // 渠道 ID
    Metadata    map[string]string // 元数据
    CreatedAt   int64             // 创建时间
    UpdatedAt   int64             // 更新时间
}
```

### Provider 接口

```go
type Provider interface {
    GetName() string                                    // 获取 Provider 名称
    GetPriority() int                                  // 获取优先级
    CanProvide(field string) bool                      // 检查是否可提供字段
    Provide(field string) (string, error)              // 提供字段值
}
```

### ProviderRegistry

```go
type ProviderRegistry struct {
    // 私有字段
}

func NewProviderRegistry() *ProviderRegistry
func (r *ProviderRegistry) RegisterProvider(provider Provider)
func (r *ProviderRegistry) BuildRuntimeInfo() (*RuntimeInfo, error)
func (r *ProviderRegistry) BuildRuntimeInfoWithFields(fields map[string]string) (*RuntimeInfo, error)
```

## 配置说明

### 环境变量

#### K8s 环境
- `POD_NAMESPACE`：Pod 命名空间
- `POD_NAME`：Pod 名称
- `NAMESPACE`：命名空间（备用）

#### 构建信息
- `VERSION`：版本号
- `APP_VERSION`：应用版本
- `BUILD_VERSION`：构建版本
- `SERVICE_NAME`：服务名称
- `APP_NAME`：应用名称
- `BUILD_TIME`：构建时间
- `GIT_COMMIT`：Git 提交
- `GIT_BRANCH`：Git 分支
- `GIT_TAG`：Git 标签

#### 配置中心
- `CONFIG_CENTER_URL`：配置中心地址
- `CONFIG_CENTER_NAMESPACE`：配置中心命名空间
- `CONFIG_CENTER_GROUP`：配置中心分组

## 最佳实践

### 1. Provider 优先级设置

```go
// 高优先级：K8s 环境信息
k8sProvider := serverinfo.NewK8sProvider()
k8sProvider.SetPriority(1)

// 中优先级：配置中心
configProvider := serverinfo.NewConfigProvider()
configProvider.SetPriority(2)

// 低优先级：构建信息
buildProvider := serverinfo.NewBuildProvider()
buildProvider.SetPriority(3)
```

### 2. Context 中间件

```go
func ServerInfoMiddleware() middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
                // 注入 ServerInfo
    info := getServerInfo() // 从注册表获取
    ctx = serverinfo.WithServerInfo(ctx, info)
            
            return handler(ctx, req)
        }
    }
}
```

### 3. 错误处理

```go
info, err := registry.BuildRuntimeInfo()
if err != nil {
    // 记录错误日志
    log.Errorf("Failed to build RuntimeInfo: %v", err)
    
    // 返回默认值或降级处理
    info = serverinfo.NewServerInfo()
    info.ServiceName = "unknown-service"
}
```

## 测试

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定测试
go test ./provider/... -v

# 运行测试并生成覆盖率报告
go test ./... -cover
```

### 测试环境配置

```bash
# 设置测试环境变量
export POD_NAMESPACE=test-namespace
export POD_NAME=test-pod
export VERSION=test-version
export SERVICE_NAME=test-service
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 更新日志

### v1.0.0
- 初始版本发布
- 支持基本的 Provider 模式
- 实现 Context 注入功能
- 提供完整的单元测试
