# ModuleKit 模块管理框架

ModuleKit 是一个基于 Kratos 框架的模块生命周期管理解决方案，提供完整的模块注册、初始化、关闭和依赖管理能力，支持 Wire 依赖注入。

## 🚀 核心特性

- **Wire 依赖注入**: 支持 Google Wire 依赖注入框架，简化模块配置
- **生命周期管理**: 与 Kratos 生命周期完全对齐，自动管理模块的完整生命周期
- **优先级排序**: 按 Order 值自动排序，低值优先初始化
- **并发安全**: 支持高并发场景下的模块管理
- **Kratos 集成**: 复用 Kratos 的错误处理、日志系统、配置管理和应用生命周期

## 📁 目录结构

```
/module                    # ModuleKit 核心包
├── interfaces.go         # Module 接口定义
├── types.go             # 类型和错误定义
├── manager.go           # ModuleManager 实现
├── wire.go              # Wire 依赖注入支持
├── examples/            # 示例代码
│   ├── user_module.go   # 用户模块示例
│   └── wire_example.go  # Wire 依赖注入示例
└── README.md           # 模块说明文档
```

## 🔧 快速开始

### 1. 定义模块

```go
type UserModule struct {
    // 通过 Wire 注入的依赖
}

func (um *UserModule) Name() string { return "user" }
func (um *UserModule) Order() int { return 1 }

func (um *UserModule) Init(ctx context.Context) error {
    // 依赖通过 Wire 自动注入，不需要手动传递
    return nil
}

func (um *UserModule) AfterAllInit(ctx context.Context) { /* ... */ }
func (um *UserModule) AfterStart(ctx context.Context) { /* ... */ }
func (um *UserModule) BeforeClose(ctx context.Context) error { /* ... */ }
func (um *UserModule) AfterStop(ctx context.Context) { /* ... */ }
```

### 2. 使用 Wire 依赖注入

```go
// wire.go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "your-project/module"
    "your-project/module/examples"
)

func InitializeApp() (*App, error) {
    wire.Build(
        module.ProviderSet,
        examples.ModuleSet,
        wire.Struct(new(App), "*"),
    )
    return nil, nil
}

// 各模块的 Provider 函数
func ProvideUserModule() module.Module {
    return examples.NewUserModule()
}

func ProvideMatchModule() module.Module {
    return examples.NewMatchModule()
}

func ProvideChatModule() module.Module {
    return examples.NewChatModule()
}
```

### 3. 运行应用

```go
func main() {
    app, err := InitializeApp()
    if err != nil {
        panic(err)
    }
    
    // 运行应用
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

## 📋 接口说明

### Module 接口

```go
type Module interface {
    Name() string                                    // 模块名称
    Order() int                                     // 初始化顺序
    Init(ctx context.Context) error
    AfterAllInit(ctx context.Context)               // 所有模块初始化后调用
    AfterStart(ctx context.Context)                 // 应用启动后调用
    BeforeClose(ctx context.Context) error          // 模块关闭前调用
    AfterStop(ctx context.Context)                  // 应用停止后调用
}
```

### ModuleManager 接口

```go
type ModuleManager interface {
    SetModules(modules []Module)
    GetModules() []Module
    InitAll(ctx context.Context) error
    AfterAllInit(ctx context.Context)
    AfterStart(ctx context.Context)
    CloseAll(ctx context.Context) error
    AfterStop(ctx context.Context)
}
```

## 🔄 生命周期流程

1. **应用启动** → Wire 自动注入模块
2. **创建 ModuleManager** → 接收模块列表
3. **Kratos BeforeStart** → 初始化所有模块
4. **AfterAllInit** → 模块间依赖协调
5. **Kratos AfterStart** → 模块启动后逻辑
6. **应用运行中** → 模块正常工作
7. **Kratos BeforeStop** → 模块关闭和清理
8. **Kratos AfterStop** → 模块停止后清理

## 🧪 测试

运行测试：

```bash
cd module
go test ./...
```

## 📚 更多信息

- 查看 `examples/` 目录获取更多使用示例
- 参考 `interfaces.go` 了解完整接口定义
- 查看 `types.go` 了解错误码和状态定义
- 查看 `wire.go` 了解 Wire 依赖注入配置

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证。
