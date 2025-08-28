// Package module 提供模块生命周期管理功能
package module

import (
	"context"
)

// Module 模块基础接口
type Module interface {
	// Name 返回模块名称
	Name() string

	// Order 返回模块的初始化顺序，低值优先
	Order() int

	// Init 初始化模块
	Init(ctx context.Context) error

	// AfterAllInit 所有模块初始化后调用
	AfterAllInit(ctx context.Context)

	// AfterStart 应用启动后调用
	AfterStart(ctx context.Context)

	// BeforeClose 模块关闭前调用
	BeforeClose(ctx context.Context) error

	// AfterStop 应用停止后调用
	AfterStop(ctx context.Context)
}
