package base

import (
	"context"
	"sync"
	"time"

	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

// BaseEntity 提供实体的通用基础实现
// 功能：
// - 标识管理：Type/ID/SetID（Type 通过 SetTypeName 设置）
// - 生命周期钩子：Init 记录首次加载时间；Destroy 逐一 Detach 已注册能力
// - 能力管理：AddAbility/GetAbility；与 ability.BaseAbility.Attach 协作
// - 脏标支持：实现 DirtyOperator，记录最近置脏时间（毫秒）
// 使用方式：业务实体组合该结构（匿名嵌入）并实现/设置自身 TypeName
//
//	type User struct { base.BaseEntity }
//	func NewUser() *User { u := &User{}; u.SetTypeName("user"); return u }
//
// 说明：SaveAble 的 Save/AutoSetDirty 由业务实体自行实现，本结构不做约束。
type BaseEntity struct {
	mu          sync.RWMutex
	id          string
	typeName    string
	loadTimeMs  int64
	abilities   map[string]facade.Ability
	dirty       bool
	dirtyTimeMs int64
}

// SetTypeName 设置实体类型名（建议在构造函数中调用）
func (b *BaseEntity) SetTypeName(name string) { b.typeName = name }

// Type 返回实体类型名
func (b *BaseEntity) Type() string { return b.typeName }

// ID 返回实体唯一 ID
func (b *BaseEntity) ID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.id
}

// SetID 设置实体唯一 ID
func (b *BaseEntity) SetID(id string) {
	b.mu.Lock()
	b.id = id
	b.mu.Unlock()
}

// Init 初始化实体：记录首次加载时间并初始化能力表
func (b *BaseEntity) Init(ctx context.Context) error {
	b.mu.Lock()
	if b.abilities == nil {
		b.abilities = make(map[string]facade.Ability)
	}
	if b.loadTimeMs == 0 {
		b.loadTimeMs = time.Now().UnixMilli()
	}
	b.mu.Unlock()
	return nil
}

// Destroy 释放资源：从实体解绑已注册的能力
func (b *BaseEntity) Destroy(ctx context.Context) error {
	b.mu.Lock()
	abilities := b.abilities
	b.abilities = make(map[string]facade.Ability)
	b.mu.Unlock()
	for _, a := range abilities {
		_ = a.Detach(ctx)
	}
	return nil
}

// LoadTime 返回首次加载时间（毫秒）
func (b *BaseEntity) LoadTime() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.loadTimeMs
}

// GetAbility 按名称获取能力实例
func (b *BaseEntity) GetAbility(name string) facade.Ability {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.abilities == nil {
		return nil
	}
	return b.abilities[name]
}

// AddAbility 绑定能力实例（通常由 Ability.Attach 调用）
func (b *BaseEntity) AddAbility(a facade.Ability) {
	if a == nil {
		return
	}
	b.mu.Lock()
	if b.abilities == nil {
		b.abilities = make(map[string]facade.Ability)
	}
	b.abilities[a.Name()] = a
	b.mu.Unlock()
}

// SetDirty 设置脏标；当置脏为 true 时记录当前毫秒时间
func (b *BaseEntity) SetDirty(dirty bool) {
	b.mu.Lock()
	b.dirty = dirty
	if dirty {
		b.dirtyTimeMs = time.Now().UnixMilli()
	}
	b.mu.Unlock()
}

// IsDirty 查询脏标
func (b *BaseEntity) IsDirty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.dirty
}

// SetDirtyTimeUnixMilli 设置最近置脏时间（毫秒）
func (b *BaseEntity) SetDirtyTimeUnixMilli(ms int64) {
	b.mu.Lock()
	b.dirtyTimeMs = ms
	b.mu.Unlock()
}

// DirtyTimeUnixMilli 返回最近置脏时间（毫秒）
func (b *BaseEntity) DirtyTimeUnixMilli() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.dirtyTimeMs
}

var _ facade.Entity = (*BaseEntity)(nil)
var _ facade.DirtyOperator = (*BaseEntity)(nil)
