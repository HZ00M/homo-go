// Package examples 提供 ModuleKit 与 Kratos 集成的示例
package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/module"
)

// 示例：如何在 Kratos 应用中使用 ModuleManager

// 1. 定义模块接口实现
// DatabaseModule 数据库模块
type DatabaseModule struct {
	logger log.Logger
}

func NewDatabaseModule(logger log.Logger) *DatabaseModule {
	return &DatabaseModule{logger: logger}
}

func (d *DatabaseModule) Order() int {
	return 10 // 高优先级，最先初始化
}

func (d *DatabaseModule) Init(ctx context.Context) error {
	d.logger.Log(log.LevelInfo, "DatabaseModule: 开始初始化")

	// 模拟数据库连接
	time.Sleep(100 * time.Millisecond)

	d.logger.Log(log.LevelInfo, "DatabaseModule: 初始化完成")
	return nil
}

func (d *DatabaseModule) AfterAllInit(ctx context.Context) {
	d.logger.Log(log.LevelInfo, "DatabaseModule: 所有模块初始化完成，数据库就绪")
}

func (d *DatabaseModule) BeforeClose(ctx context.Context) {
	d.logger.Log(log.LevelInfo, "DatabaseModule: 开始关闭数据库连接")
	time.Sleep(50 * time.Millisecond)
	d.logger.Log(log.LevelInfo, "DatabaseModule: 数据库连接已关闭")
}

// RedisModule Redis 模块
type RedisModule struct {
	logger log.Logger
}

func NewRedisModule(logger log.Logger) *RedisModule {
	return &RedisModule{logger: logger}
}

func (r *RedisModule) Order() int {
	return 20 // 在数据库模块之后初始化
}

func (r *RedisModule) Init(ctx context.Context) error {
	r.logger.Log(log.LevelInfo, "RedisModule: 开始初始化")

	// 模拟 Redis 连接
	time.Sleep(80 * time.Millisecond)

	r.logger.Log(log.LevelInfo, "RedisModule: 初始化完成")
	return nil
}

func (r *RedisModule) AfterAllInit(ctx context.Context) {
	r.logger.Log(log.LevelInfo, "RedisModule: 所有模块初始化完成，Redis 就绪")
}

func (r *RedisModule) BeforeClose(ctx context.Context) {
	r.logger.Log(log.LevelInfo, "RedisModule: 开始关闭 Redis 连接")
	time.Sleep(30 * time.Millisecond)
	r.logger.Log(log.LevelInfo, "RedisModule: Redis 连接已关闭")
}

// ServiceModule 服务模块
type ServiceModule struct {
	logger log.Logger
}

func NewServiceModule(logger log.Logger) *ServiceModule {
	return &ServiceModule{logger: logger}
}

func (s *ServiceModule) Order() int {
	return 100 // 在基础设施模块之后初始化
}

func (s *ServiceModule) Init(ctx context.Context) error {
	s.logger.Log(log.LevelInfo, "ServiceModule: 开始初始化")

	// 模拟服务逻辑初始化
	time.Sleep(120 * time.Millisecond)

	s.logger.Log(log.LevelInfo, "ServiceModule: 初始化完成")
	return nil
}

func (s *ServiceModule) AfterAllInit(ctx context.Context) {
	s.logger.Log(log.LevelInfo, "ServiceModule: 所有模块初始化完成，服务系统就绪")
}

func (s *ServiceModule) BeforeClose(ctx context.Context) {
	s.logger.Log(log.LevelInfo, "ServiceModule: 开始关闭服务系统")
	time.Sleep(80 * time.Millisecond)
	s.logger.Log(log.LevelInfo, "ServiceModule: 服务系统已关闭")
}

// 2. 模块管理集成示例
// ModuleManagerIntegration 展示如何在 Kratos 应用中使用 ModuleManager
type ModuleManagerIntegration struct {
	moduleManager *module.ModuleManager
	logger        log.Logger
}

// NewModuleManagerIntegration 创建模块管理集成实例
func NewModuleManagerIntegration(
	modules []module.Module,
	logger log.Logger,
) *ModuleManagerIntegration {
	return &ModuleManagerIntegration{
		moduleManager: module.NewModuleManager(modules),
		logger:        logger,
	}
}

// Start 启动所有模块
func (m *ModuleManagerIntegration) Start(ctx context.Context) error {
	m.logger.Log(log.LevelInfo, "开始启动模块系统...")

	// 1. 初始化所有模块
	if err := m.moduleManager.InitAll(ctx); err != nil {
		return fmt.Errorf("模块初始化失败: %w", err)
	}

	// 2. 执行所有模块的 AfterAllInit 回调
	m.moduleManager.AfterAllInit(ctx)

	m.logger.Log(log.LevelInfo, "模块系统启动完成")
	return nil
}

// Stop 停止所有模块
func (m *ModuleManagerIntegration) Stop(ctx context.Context) error {
	m.logger.Log(log.LevelInfo, "开始停止模块系统...")

	// 关闭所有模块
	if err := m.moduleManager.CloseAll(ctx); err != nil {
		return fmt.Errorf("模块关闭失败: %w", err)
	}

	m.logger.Log(log.LevelInfo, "模块系统已停止")
	return nil
}

// 3. 使用示例
func ExampleUsage() {
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
		logger.Log(log.LevelError, "启动失败", "error", err)
		return
	}

	// 模拟应用运行
	time.Sleep(2 * time.Second)

	// 停止模块系统
	if err := integration.Stop(ctx); err != nil {
		logger.Log(log.LevelError, "停止失败", "error", err)
		return
	}
}

// 4. Wire 依赖注入示例（注释形式）
/*
// 在 Wire 配置中使用
var ProviderSet = wire.NewSet(
	// 提供模块实例
	NewDatabaseModule,
	NewRedisModule,
	NewServiceModule,

	// 收集模块到列表
	func(db *DatabaseModule, redis *RedisModule, service *ServiceModule) []module.Module {
		return []module.Module{db, redis, service}
	},

	// 提供模块管理器
	module.ProvideModuleManager,

	// 提供集成实例
	NewModuleManagerIntegration,
)

// 在 Kratos 应用中使用
func main() {
	// 通过 Wire 获取应用实例
	app, err := InitializeApp()
	if err != nil {
		log.Fatal(err)
	}

	// 启动应用
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	if err := app.Stop(); err != nil {
		log.Errorf("应用关闭失败: %v", err)
	}
}
*/

// 关键特性说明：
// 1. 模块按 Order 值顺序初始化
// 2. 相同 Order 的模块并发初始化
// 3. 支持优雅关闭
// 4. 完整的错误处理
// 5. 与 Kratos 框架完美集成
// 6. 支持 Wire 依赖注入
