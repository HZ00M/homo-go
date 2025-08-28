# ModuleKit 使用示例

本目录包含了 ModuleKit 的使用示例，展示了如何实现模块接口和集成到生命周期管理中。

## 文件说明

### basic.go
基础使用示例，包含三个模块：
- **LogModule**: 日志模块，优先级 10，最先初始化
- **ConfigModule**: 配置模块，优先级 20，在日志模块之后初始化
- **BusinessModule**: 业务模块，优先级 100，在基础设施模块之后初始化

### wire.go
依赖注入示例，展示如何通过 Wire 框架注入模块列表到 ModuleManager。

## 使用方法

### 1. 实现 Module 接口

```go
type MyModule struct {
    // 模块字段
}

// 实现 Order 方法
func (m *MyModule) Order() int {
    return 50 // 设置初始化优先级
}

// 实现 Init 方法
func (m *MyModule) Init(ctx context.Context) error {
    // 初始化逻辑
    return nil
}

// 实现 AfterAllInit 方法
func (m *MyModule) AfterAllInit(ctx context.Context) {
    // 所有模块初始化完成后的回调
}

// 实现 BeforeClose 方法
func (m *MyModule) BeforeClose(ctx context.Context) {
    // 模块关闭前的清理
}
```

### 2. 优先级建议

- **0-9**: 系统级模块（如日志、配置）
- **10-99**: 基础设施模块（如数据库、缓存）
- **100-999**: 业务服务模块
- **1000+**: 业务逻辑模块

### 3. 生命周期流程

1. **Init**: 按 Order 顺序初始化，相同 Order 的模块并发初始化
2. **AfterAllInit**: 所有模块初始化完成后，按 Order 顺序执行回调
3. **BeforeClose**: 系统关闭时，按反向 Order 顺序执行清理

### 4. 错误处理

- 如果任何模块的 `Init` 方法返回错误，整个系统不能达到可服务状态
- 使用 Kratos 的错误处理机制，支持错误码和错误信息

### 5. Wire 依赖注入

ModuleKit 支持通过 Wire 框架进行依赖注入：

```go
// 在 wire.go 中
func InitializeApp(ctx context.Context) (*examples.App, error) {
    wire.Build(
        // 提供日志器
        provideLogger,
        
        // 提供各模块
        examples.ProvideLogModule,
        examples.ProvideConfigModule,
        examples.ProvideBusinessModule,
        
        // 收集所有模块到列表
        collectModules,
        
        // 提供模块管理器（接收模块列表）
        module.ProvideModuleManager,
        
        // 构建应用
        examples.NewApp,
    )
    return nil, nil
}

// 收集模块到列表
func collectModules(
    logModule examples.Module,
    configModule examples.Module,
    businessModule examples.Module,
) []examples.Module {
    return []examples.Module{
        logModule,
        configModule,
        businessModule,
    }
}
```

**关键特性：**
- `ProvideModuleManager(modules []Module)` 接收模块列表作为参数
- 各模块通过各自的 Provider 函数提供，返回类型为 `module.Module`
- Wire 自动处理依赖关系，确保正确的初始化顺序
- 支持模块间的依赖注入（如 BusinessModule 依赖 ConfigModule）
- 采用依赖注入模式，每个应用实例都有独立的 ModuleManager，无需单例模式

## 注意事项

1. **并发安全**: 相同 Order 的模块会并发初始化，确保模块实现是线程安全的
2. **依赖管理**: 通过 Order 值控制初始化顺序，避免循环依赖
3. **资源清理**: 在 `BeforeClose` 方法中正确清理资源，避免内存泄漏
4. **错误传播**: 初始化错误会阻止系统启动，确保错误信息清晰明确

## 扩展建议

1. **监控集成**: 在生命周期方法中添加监控指标
2. **健康检查**: 实现健康检查接口，监控模块状态
3. **配置热更新**: 支持运行时配置更新
4. **优雅关闭**: 实现超时控制和强制关闭机制
