package base

import (
	"context"
	"reflect"

	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

// BaseAbility 提供 Ability 的基础实现，用于被业务 Ability 组合/嵌入
// - 默认 Name 返回组合结构体的类型名
// - 默认 Attach 记录 owner 并将“真实能力实例”注册到实体（无需各能力重复实现）
// - 默认 Detach 清理 owner 引用
// 组合用法：
//
//	type MyAbility struct { ability.BaseAbility }
//	func NewMyAbility() *MyAbility { a := &MyAbility{}; a.Bind(a); return a }
//	// 之后可仅实现业务方法，如：func (a *MyAbility) Do(ctx context.Context) error { ... }
//
// 注意：使用 Bind(self) 或 NewXxx 构造函数在创建时绑定自身实例，以便 Attach 时注册正确的能力实例
// 而不是将 BaseAbility 自身注册到实体。
type BaseAbility struct {
	self  facade.Ability
	owner facade.Entity
}

// NewBaseAbility 创建基础能力并绑定真实能力实例
func NewBaseAbility(self facade.Ability) BaseAbility {
	return BaseAbility{self: self}
}

// Bind 绑定真实能力实例（建议在构造函数中调用）
func (b *BaseAbility) Bind(self facade.Ability) { b.self = self }

// Name 默认返回组合结构体的类型名
func (b *BaseAbility) Name() string {
	if b.self != nil {
		t := reflect.TypeOf(b.self)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return t.Name()
	}
	return "ability"
}

// Attach 记录 owner 并将真实能力实例注册到实体
func (b *BaseAbility) Attach(ctx context.Context, owner facade.Entity) error {
	b.owner = owner
	if b.self != nil {
		owner.AddAbility(b.self)
	} else {
		owner.AddAbility(b)
	}
	return nil
}

// Detach 清理 owner 引用
func (b *BaseAbility) Detach(ctx context.Context) error {
	b.owner = nil
	return nil
}
