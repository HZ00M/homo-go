# ModuleManager 与 Kratos 集成示例

本文档详细说明了如何在 Kratos 应用中使用 ModuleManager 进行模块生命周期管理。

## 🎯 核心概念

### 1. 模块生命周期管理
- **初始化顺序**：按 `Order()` 值顺序初始化，相同 Order 的模块并发初始化
- **依赖注入**：通过 Wire 框架自动注入模块依赖
- **优雅关闭**：按反向 Order 顺序关闭模块
- **错误处理**：任何模块初始化失败都会阻止系统启动

### 2. 与 Kratos 的集成点
- **应用启动**：在 `app.Start()` 之前初始化所有模块
- **应用关闭**：在 `app.Stop()` 之后关闭所有模块
- **日志系统**：复用 Kratos 的 `log.Logger`
- **配置管理**：复用 Kratos 的配置系统

## 🏗️ 架构设计

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Kratos App   │    │ ModuleManager    │    │   Modules      │
│                 │    │                  │    │                 │
│  ┌───────────┐ │    │  ┌─────────────┐ │    │ ┌─────────────┐ │
│  │   Start  │ │───▶│  │  InitAll    │ │───▶│ │Database    │ │
│  └───────────┘ │    │  └─────────────┘ │    │ │Order: 10   │ │
│                 │    │                  │    │ └─────────────┘ │
│  ┌───────────┐ │    │  ┌─────────────┐ │    │ ┌─────────────┐ │
│  │   Stop   │ │◀───│  │  CloseAll   │ │◀───│ │Redis       │ │
│  └───────────┘ │    │  └─────────────┘ │    │ │Order: 20   │ │
└─────────────────┘    └──────────────────┘    │ └─────────────┘ │
                                               │ ┌─────────────┐ │
                                               │ │Service     │ │
                                               │ │Order: 100  │ │
                                               │ └─────────────┘ │
                                               └─────────────────┘
```

## 📋 模块定义示例

### 1. 数据库模块（高优先级）
```go
type DatabaseModule struct {
    logger log.Logger
}

func (d *DatabaseModule) Order() int {
    return 10 // 最先初始化
}

func (d *DatabaseModule) Init(ctx context.Context) error {
    d.logger.Log(log.LevelInfo, "DatabaseModule: 开始初始化")
    // 初始化数据库连接
    return nil
}
```

### 2. Redis 模块（中优先级）
```go
type RedisModule struct {
    logger log.Logger
}

func (r *RedisModule) Order() int {
    return 20 // 在数据库之后初始化
}
```

### 3. 服务模块（低优先级）
```go
type ServiceModule struct {
    logger log.Logger
}

func (s *ServiceModule) Order() int {
    return 100 // 在基础设施之后初始化
}
```

## 🔧 集成实现

### 1. 创建集成结构体
```go
type ModuleManagerIntegration struct {
    moduleManager *module.ModuleManager
    logger        log.Logger
}

func NewModuleManagerIntegration(
    modules []module.Module,
    logger log.Logger,
) *ModuleManagerIntegration {
    return &ModuleManagerIntegration{
        moduleManager: module.NewModuleManager(modules),
        logger:        logger,
    }
}
```

### 2. 启动模块系统
```go
func (m *ModuleManagerIntegration) Start(ctx context.Context) error {
    // 1. 初始化所有模块
    if err := m.moduleManager.InitAll(ctx); err != nil {
        return fmt.Errorf("模块初始化失败: %w", err)
    }
    
    // 2. 执行所有模块的 AfterAllInit 回调
    m.moduleManager.AfterAllInit(ctx)
    
    return nil
}
```

### 3. 停止模块系统
```go
func (m *ModuleManagerIntegration) Stop(ctx context.Context) error {
    // 关闭所有模块
    if err := m.moduleManager.CloseAll(ctx); err != nil {
        return fmt.Errorf("模块关闭失败: %w", err)
    }
    
    return nil
}
```

## 🚀 使用方式

### 1. 直接使用
```go
func main() {
    // 创建日志记录器
    logger := log.DefaultLogger
    
    // 创建模块列表
    modules := []module.Module{
        NewDatabaseModule(logger),
        NewRedisModule(logger),
        NewServiceModule(logger),
    }
    
    // 创建模块管理集成
    integration := NewModuleManagerIntegration(modules, logger)
    
    // 启动模块系统
    ctx := context.Background()
    if err := integration.Start(ctx); err != nil {
        log.Fatal(err)
    }
    
    // 应用运行...
    
    // 停止模块系统
    if err := integration.Stop(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### 2. 通过 Wire 依赖注入
```go
// Wire 配置
var ProviderSet = wire.NewSet(
    NewDatabaseModule,
    NewRedisModule,
    NewServiceModule,
    
    // 收集模块到列表
    func(db *DatabaseModule, redis *RedisModule, service *ServiceModule) []module.Module {
        return []module.Module{db, redis, service}
    },
    
    module.ProvideModuleManager,
    NewModuleManagerIntegration,
)

// 在 Kratos 应用中使用
func main() {
    app, err := InitializeApp()
    if err != nil {
        log.Fatal(err)
    }
    
    // 启动应用（包含模块初始化）
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
    
    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // 优雅关闭（包含模块关闭）
    if err := app.Stop(); err != nil {
        log.Errorf("应用关闭失败: %v", err)
    }
}
```

## 📊 初始化顺序示例

假设有以下模块：
- DatabaseModule: Order = 10
- RedisModule: Order = 20
- CacheModule: Order = 20
- ServiceModule: Order = 100

初始化流程：
```
1. Order 10: DatabaseModule (单独初始化)
2. Order 20: RedisModule + CacheModule (并发初始化)
3. Order 100: ServiceModule (单独初始化)
4. AfterAllInit: 所有模块按 Order 顺序执行回调
```

## ⚠️ 注意事项

### 1. 模块设计原则
- **单一职责**：每个模块只负责一个功能领域
- **依赖明确**：通过 Order 值明确表达依赖关系
- **错误处理**：模块初始化失败必须返回错误
- **资源清理**：在 `BeforeClose` 中清理所有资源

### 2. 性能考虑
- **并发初始化**：相同 Order 的模块并发初始化
- **超时控制**：使用 context 控制初始化超时
- **资源限制**：避免同时创建过多连接

### 3. 错误处理
- **初始化失败**：任何模块失败都会阻止系统启动
- **关闭失败**：记录错误但不阻止其他模块关闭
- **日志记录**：详细记录每个阶段的执行情况

## 🔍 调试和监控

### 1. 日志输出
```
DatabaseModule: 开始初始化
DatabaseModule: 初始化完成
RedisModule: 开始初始化
RedisModule: 初始化完成
ServiceModule: 开始初始化
ServiceModule: 初始化完成
DatabaseModule: 所有模块初始化完成，数据库就绪
RedisModule: 所有模块初始化完成，Redis 就绪
ServiceModule: 所有模块初始化完成，服务系统就绪
```

### 2. 健康检查
- 提供 `/health` 端点检查系统状态
- 提供 `/modules/status` 端点检查模块状态
- 监控模块初始化和关闭时间

## 📚 扩展阅读

- [ModuleKit 核心接口](./interfaces.go)
- [ModuleManager 实现](./manager.go)
- [错误处理](./errors.go)
- [基础示例](./basic.go)
- [Wire 集成](./wire.go)

## 🎉 总结

ModuleManager 与 Kratos 的集成提供了：

1. **统一的模块生命周期管理**
2. **灵活的依赖注入支持**
3. **可靠的错误处理机制**
4. **优雅的启动和关闭流程**
5. **清晰的架构设计模式**

这种集成方式特别适合需要管理多个功能模块的微服务应用，能够确保系统的可靠性和可维护性。
