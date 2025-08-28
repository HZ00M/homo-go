// Package examples 提供 ModuleKit 的使用示例
package examples

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/module"
)

// LogModule 日志模块示例
// 优先级最高，最先初始化
type LogModule struct {
	logger log.Logger
}

// NewLogModule 创建日志模块实例
func NewLogModule(logger log.Logger) *LogModule {
	return &LogModule{
		logger: logger,
	}
}

// Order 返回模块优先级
func (l *LogModule) Order() int {
	return 10 // 高优先级，最先初始化
}

// Init 初始化日志模块
func (l *LogModule) Init(ctx context.Context) error {
	log.Info("LogModule: 开始初始化")

	// 模拟初始化过程
	time.Sleep(100 * time.Millisecond)

	log.Info("LogModule: 初始化完成")
	return nil
}

// AfterAllInit 所有模块初始化完成后的回调
func (l *LogModule) AfterAllInit(ctx context.Context) {
	log.Info("LogModule: 所有模块初始化完成，日志系统就绪")
}

// BeforeClose 模块关闭前的清理
func (l *LogModule) BeforeClose(ctx context.Context) {
	log.Info("LogModule: 开始关闭日志系统")

	// 模拟清理过程
	time.Sleep(50 * time.Millisecond)

	log.Info("LogModule: 日志系统已关闭")
}

// ProvideLogModule Wire Provider 函数
func ProvideLogModule(logger log.Logger) module.Module {
	return NewLogModule(logger)
}

// ConfigModule 配置模块示例
// 优先级较高，在日志模块之后初始化
type ConfigModule struct {
	logger log.Logger
	config map[string]interface{}
}

// NewConfigModule 创建配置模块实例
func NewConfigModule(logger log.Logger) *ConfigModule {
	return &ConfigModule{
		logger: logger,
		config: make(map[string]interface{}),
	}
}

// Order 返回模块优先级
func (c *ConfigModule) Order() int {
	return 20 // 在日志模块之后初始化
}

// Init 初始化配置模块
func (c *ConfigModule) Init(ctx context.Context) error {
	log.Info("ConfigModule: 开始初始化")

	// 模拟加载配置
	c.config["app.name"] = "ModuleKit Demo"
	c.config["app.version"] = "1.0.0"
	c.config["app.env"] = "development"

	// 模拟初始化过程
	time.Sleep(150 * time.Millisecond)

	log.Info("ConfigModule: 初始化完成")
	return nil
}

// AfterAllInit 所有模块初始化完成后的回调
func (c *ConfigModule) AfterAllInit(ctx context.Context) {
	log.Info("ConfigModule: 所有模块初始化完成，配置系统就绪")
}

// BeforeClose 模块关闭前的清理
func (c *ConfigModule) BeforeClose(ctx context.Context) {
	log.Info("ConfigModule: 开始关闭配置系统")

	// 模拟清理过程
	time.Sleep(50 * time.Millisecond)

	log.Info("ConfigModule: 配置系统已关闭")
}

// GetConfig 获取配置值
func (c *ConfigModule) GetConfig(key string) interface{} {
	return c.config[key]
}

// ProvideConfigModule Wire Provider 函数
func ProvideConfigModule(logger log.Logger) module.Module {
	return NewConfigModule(logger)
}

// BusinessModule 业务模块示例
// 优先级较低，在基础设施模块之后初始化
type BusinessModule struct {
	logger log.Logger
	config *ConfigModule
}

// NewBusinessModule 创建业务模块实例
func NewBusinessModule(logger log.Logger, config *ConfigModule) *BusinessModule {
	return &BusinessModule{
		logger: logger,
		config: config,
	}
}

// Order 返回模块优先级
func (b *BusinessModule) Order() int {
	return 100 // 在基础设施模块之后初始化
}

// Init 初始化业务模块
func (b *BusinessModule) Init(ctx context.Context) error {
	log.Info("BusinessModule: 开始初始化")

	// 模拟业务初始化过程
	time.Sleep(200 * time.Millisecond)

	log.Info("BusinessModule: 初始化完成")
	return nil
}

// AfterAllInit 所有模块初始化完成后的回调
func (b *BusinessModule) AfterAllInit(ctx context.Context) {
	appName := b.config.GetConfig("app.name")
	log.Infof("BusinessModule: 所有模块初始化完成，业务系统就绪，应用名称: %v", appName)
}

// BeforeClose 模块关闭前的清理
func (b *BusinessModule) BeforeClose(ctx context.Context) {
	log.Info("BusinessModule: 开始关闭业务系统")

	// 模拟清理过程
	time.Sleep(100 * time.Millisecond)

	log.Info("BusinessModule: 业务系统已关闭")
}

// ProvideBusinessModule Wire Provider 函数
func ProvideBusinessModule(logger log.Logger, config *ConfigModule) module.Module {
	return NewBusinessModule(logger, config)
}
