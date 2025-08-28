// Package module 提供模块生命周期管理功能
package module

import (
	"github.com/google/wire"
)

// ProviderSet 提供 Wire 依赖注入的 Provider 集合
var ProviderSet = wire.NewSet(
	// 提供 ModuleManager
	ProvideModuleManager,
)

// ProvideModuleManager 提供 ModuleManager 实例
func ProvideModuleManager(modules []Module) *ModuleManager {
	return NewModuleManager(modules)
}

// ProvideModuleSlice 提供模块切片（用于 Wire 自动注入）
// 这个函数需要在具体的应用中使用 wire.Build 来提供具体的模块实现
func ProvideModuleSlice() []Module {
	// 返回空切片，具体的模块会通过 wire.Build 注入
	return []Module{}
}
