package facade

import "context"

// Entity 有状态实体的最小契约
// 作用：抽象业务实体的生命周期、标识、能力管理。
type Entity interface {
	// Type: 实体类型（唯一标识该类实体）
	Type() string
	// ID: 实例唯一 ID
	ID() string
	// SetID: 设置实例 ID（构造后/反序列化后调用）
	SetID(id string)
	// Init: 初始化（同步），用于 attach 能力、加载依赖等
	Init(ctx context.Context) error
	// Destroy: 释放资源（保存、解绑能力、停止协程等）
	Destroy(ctx context.Context) error
	// LoadTime: 首次加载到内存的时间（毫秒）
	LoadTime() int64
	// GetAbility: 按名称获取能力实例（运行时使用，业务侧避免）
	GetAbility(name string) Ability
	// AddAbility: 绑定能力实例（由运行时调用）
	AddAbility(a Ability)
}

// DirtyOperator 脏标管理
type DirtyOperator interface {
	SetDirty(dirty bool)
	IsDirty() bool
	SetDirtyTimeUnixMilli(ms int64)
	DirtyTimeUnixMilli() int64
}

// Saveable 可存储能力
// 作用：定义实体如何被持久化；通常由存储能力实现。
type Saveable interface {
	DirtyOperator
	// Save: 将实体持久化到存储
	Save(ctx context.Context) error
	// AutoSetDirty: 是否在访问时自动置脏
	AutoSetDirty() bool
}
