package facade

import "context"

// EntityMgr 实体生命周期与缓存管理
type EntityMgr interface {
	// Exists: 是否存在（内存或存储）
	Exists(ctx context.Context, entityType, id string) (bool, error)
	// Create: 创建并初始化实体（不落地）
	Create(ctx context.Context, entityType, id string, ctor func() Entity) (Entity, error)
	// Get: 获取实体（命中内存则续期与置脏，未命中可触发加载/Hook）
	Get(ctx context.Context, entityType, id string) (Entity, error)
	// GetNoKeepAlive: 获取实体（不续期、不置脏）
	GetNoKeepAlive(ctx context.Context, entityType, id string) (Entity, error)
	// GetOrCreate: 获取或创建
	GetOrCreate(ctx context.Context, entityType, id string, ctor func() Entity) (Entity, error)
	// Remove: 仅移出内存（不删除持久化数据）
	Remove(ctx context.Context, entityType, id string) error
	// DestroyAllType: 同一 id 的不同 type 都销毁（仅内存）
	DestroyAllType(ctx context.Context, id string) error
	// ReleaseAll: 释放全部实体（等待落地与路由缓存过期）
	ReleaseAll(ctx context.Context) error
	// IsAllLanded: 是否所有可存储实体已落地
	IsAllLanded(ctx context.Context) (bool, error)
	// RegisterNotFoundHook: miss 钩子
	RegisterNotFoundHook(entityType string, fn func(ctx context.Context, id string) (Entity, error))

	// 以下注册逻辑为 Java 侧 TpfObject/AbilityAble 的 Go 等价实现：
	// abilityName 对应 Ability.Name()；TpfObject 对应为 Entity。

	// RegisterCreateProcess: 注册按能力的创建流程钩子
	// abilityName: 能力名；consumer: 钩子函数，收到目标实体
	RegisterCreateProcess(abilityName string, consumer func(ctx context.Context, e Entity))
	// RegisterAddProcess: 注册按能力的添加流程钩子
	RegisterAddProcess(abilityName string, consumer func(ctx context.Context, e Entity))
	// RegisterGetProcess: 注册按能力的获取流程钩子
	RegisterGetProcess(abilityName string, consumer func(ctx context.Context, e Entity))
	// RegisterRemoveProcess: 注册按能力的移除流程钩子
	RegisterRemoveProcess(abilityName string, consumer func(ctx context.Context, e Entity))
	// RegisterBeforeAddProcess: 注册按能力的前置添加流程钩子
	// fun: 前置处理，可返回要添加的 Ability；返回 error 表示失败
	RegisterBeforeAddProcess(abilityName string, fun func(ctx context.Context, e Entity) (Ability, error))
}
