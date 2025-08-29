# ServerInfo 全局注册表使用说明

## 概述

ServerInfo 模块提供了全局的 Provider 注册表，支持在任何地方注册 Provider 并构建 ServerInfo 实例。这个设计使得模块更加灵活，便于在不同场景下使用。

## 核心功能

### 1. 简化的全局访问

- **`Get()`**: 获取全局唯一的 ServerInfo 实例（推荐使用）
- **`RegisterGlobalProvider(provider Provider)`**: 全局注册 Provider 的便捷函数
- **`GetGlobalRegistry()`**: 获取全局 ProviderRegistry 实例（用于高级操作）

### 2. 依赖注入支持

- **`ProvideServerInfoForWire()`**: 用于依赖注入框架的函数，内部调用 `Get()`
- **`ProvideServerInfo()`**: 兼容性函数，返回 ServerInfo 和错误

## 使用方式

### 基本使用（推荐）

```go
package main

import (
    "github.com/go-kratos/kratos/v2/serverinfo"
    "github.com/go-kratos/kratos/v2/serverinfo/provider"
)

func main() {
    // 1. 创建 Provider
    k8sProvider := provider.NewK8sProvider()
    configProvider := provider.NewConfigProvider()
    
    // 2. 全局注册
    serverinfo.RegisterGlobalProvider(k8sProvider)
    serverinfo.RegisterGlobalProvider(configProvider)
    
    // 3. 获取 ServerInfo（推荐方式）
    info := serverinfo.Get()
    
    // 4. 使用 ServerInfo
    fmt.Printf("Service: %s, Namespace: %s\n", 
        info.ServiceName, info.Namespace)
}
```

### 在应用启动时注册

```go
func init() {
    // 应用启动时注册所有 Provider
    serverinfo.RegisterGlobalProvider(provider.NewK8sProvider())
    serverinfo.RegisterGlobalProvider(provider.NewConfigProvider())
    serverinfo.RegisterGlobalProvider(provider.NewBuildProvider())
    serverinfo.RegisterGlobalProvider(provider.NewLocalProvider())
}
```

### 在中间件中使用

```go
func ServerInfoMiddleware() middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req interface{}) (interface{}, error) {
            // 使用 Get 方法获取 ServerInfo
            info := serverinfo.Get()
            
            // 注入到 Context
            ctx = serverinfo.WithServerInfo(ctx, info)
            return handler(ctx, req)
        }
    }
}
```

### 在 gRPC 拦截器中使用

```go
func ServerInfoInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 使用 Get 方法获取 ServerInfo
        serverInfo := serverinfo.Get()
        
        // 注入到 Context
        ctx = serverinfo.WithServerInfo(ctx, serverInfo)
        return handler(ctx, req)
    }
}
```

## 高级用法

### 获取注册表实例进行更多操作

```go
registry := serverinfo.GetGlobalRegistry()

// 检查 Provider 数量
count := registry.GetProviderCount()
fmt.Printf("注册的 Provider 数量: %d\n", count)

// 按优先级获取 Provider 列表
providers := registry.GetProvidersByPriority()
for i, p := range providers {
    fmt.Printf("%d. %s (优先级: %d)\n", i+1, p.GetName(), p.GetPriority())
}

// 检查特定 Provider 是否存在
if registry.HasProvider("k8s") {
    fmt.Println("K8s Provider 已注册")
}
```

### 动态管理 Provider

```go
registry := serverinfo.GetGlobalRegistry()

// 移除特定 Provider
registry.RemoveProvider("config")

// 清空所有 Provider
registry.ClearProviders()

// 重新注册
serverinfo.RegisterGlobalProvider(provider.NewConfigProvider())
```

## 设计优势

1. **简化使用**：通过 `Get()` 方法一键获取 ServerInfo，无需处理错误
2. **全局访问**：在任何地方都可以访问和注册 Provider
3. **单例模式**：确保全局只有一个 ServerInfo 实例
4. **线程安全**：使用读写锁保证并发安全
5. **优先级管理**：支持 Provider 优先级排序
6. **灵活配置**：支持动态添加和移除 Provider
7. **依赖注入友好**：提供适合依赖注入框架的接口

## 注意事项

1. **启动时注册**：建议在应用启动时注册所有 Provider
2. **优先级设置**：合理设置 Provider 优先级，避免冲突
3. **错误处理**：`Get()` 方法内部处理错误，返回空实例而不是错误
4. **并发安全**：全局注册表是线程安全的，可以并发使用
5. **内存管理**：注册的 Provider 会一直保存在内存中，注意内存使用
6. **兼容性**：原有的 `BuildGlobalServerInfo()` 等方法仍然可用

## 示例代码

完整的示例代码请参考 `global_usage.go` 文件，展示了三种使用方式：

1. **`Get()` 方法**：推荐使用，简单直接
2. **`BuildGlobalServerInfo()` 方法**：保持兼容性
3. **全局注册表操作**：高级用法，用于动态管理
