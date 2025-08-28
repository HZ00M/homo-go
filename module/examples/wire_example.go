package examples

import (
	"context"

	"github.com/google/wire"
)

// Wire 依赖注入示例

// ProvideUserModule 提供用户模块
func ProvideUserModule() Module {
	return NewUserModule()
}

// ProvideMatchModule 提供匹配模块
func ProvideMatchModule() Module {
	return NewMatchModule()
}

// ProvideChatModule 提供聊天模块
func ProvideChatModule() Module {
	return NewChatModule()
}

// ModuleSet 模块集合
var ModuleSet = wire.NewSet(
	ProvideUserModule,
	ProvideMatchModule,
	ProvideChatModule,
)

// MatchModule 匹配模块示例
type MatchModule struct {
	// 通过 Wire 注入的依赖
}

// NewMatchModule 创建匹配模块
func NewMatchModule() *MatchModule {
	return &MatchModule{}
}

// Name 返回模块名称
func (mm *MatchModule) Name() string {
	return "match"
}

// Order 返回模块初始化顺序
func (mm *MatchModule) Order() int {
	return 2
}

// Init 初始化模块
func (mm *MatchModule) Init(ctx context.Context) error {
	// 依赖通过 Wire 自动注入，不需要手动传递
	return nil
}

// AfterAllInit 所有模块初始化后调用
func (mm *MatchModule) AfterAllInit(ctx context.Context) {
	// 模块间依赖协调逻辑
}

// AfterStart 应用启动后调用
func (mm *MatchModule) AfterStart(ctx context.Context) {
	// 启动后逻辑
}

// BeforeClose 模块关闭前调用
func (mm *MatchModule) BeforeClose(ctx context.Context) error {
	// 关闭前清理逻辑
	return nil
}

// AfterStop 应用停止后调用
func (mm *MatchModule) AfterStop(ctx context.Context) {
	// 停止后清理逻辑
}

// ChatModule 聊天模块示例
type ChatModule struct {
	// 通过 Wire 注入的依赖
}

// NewChatModule 创建聊天模块
func NewChatModule() *ChatModule {
	return &ChatModule{}
}

// Name 返回模块名称
func (cm *ChatModule) Name() string {
	return "chat"
}

// Order 返回模块初始化顺序
func (cm *ChatModule) Order() int {
	return 3
}

// Init 初始化模块
func (cm *ChatModule) Init(ctx context.Context) error {
	// 依赖通过 Wire 自动注入，不需要手动传递
	return nil
}

// AfterAllInit 所有模块初始化后调用
func (cm *ChatModule) AfterAllInit(ctx context.Context) {
	// 模块间依赖协调逻辑
}

// AfterStart 应用启动后调用
func (cm *ChatModule) AfterStart(ctx context.Context) {
	// 启动后逻辑
}

// BeforeClose 模块关闭前调用
func (cm *ChatModule) BeforeClose(ctx context.Context) error {
	// 关闭前清理逻辑
	return nil
}

// AfterStop 应用停止后调用
func (cm *ChatModule) AfterStop(ctx context.Context) {
	// 停止后清理逻辑
}
