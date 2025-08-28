package examples

import (
	"context"
)

// UserModule 用户模块示例
type UserModule struct {
	// 通过 Wire 注入的依赖
}

// NewUserModule 创建用户模块
func NewUserModule() *UserModule {
	return &UserModule{}
}

// Name 返回模块名称
func (um *UserModule) Name() string {
	return "user"
}

// Order 返回模块初始化顺序
func (um *UserModule) Order() int {
	return 1
}

// Init 初始化模块
func (um *UserModule) Init(ctx context.Context) error {
	// 这里可以添加具体的初始化逻辑
	// 例如：初始化数据库连接、加载配置等
	// 依赖通过 Wire 自动注入，不需要手动传递

	return nil
}

// AfterAllInit 所有模块初始化后调用
func (um *UserModule) AfterAllInit(ctx context.Context) {
	// 这里可以添加模块间依赖协调逻辑
	// 例如：注册路由、初始化缓存等
}

// AfterStart 应用启动后调用
func (um *UserModule) AfterStart(ctx context.Context) {
	// 这里可以添加启动后的逻辑
	// 例如：启动定时任务、注册健康检查等
}

// BeforeClose 模块关闭前调用
func (um *UserModule) BeforeClose(ctx context.Context) error {
	// 这里可以添加关闭前的清理逻辑
	// 例如：停止定时任务、关闭连接等

	return nil
}

// AfterStop 应用停止后调用
func (um *UserModule) AfterStop(ctx context.Context) {
	// 这里可以添加停止后的清理逻辑
	// 例如：清理缓存、保存状态等
}
