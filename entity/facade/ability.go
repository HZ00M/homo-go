package facade

import "context"

// Ability 能力最小契约
// 作用：封装实体可挂载的横向能力（存储、定时、发布订阅等）。
type Ability interface {
	// Name: 能力名（唯一，建议常量）
	Name() string
	// Attach: 绑定到实体（做内部初始化）
	Attach(ctx context.Context, owner Entity) error
	// Detach: 从实体解绑（做内部清理）
	Detach(ctx context.Context) error
}
type AbilitySystem interface {
	Init(ctx context.Context, eMgr EntityMgr)
}
