package storage

import (
	"context"
	"github.com/go-kratos/kratos/v2/entity/facade"
)

type StorageSystemImpl struct {
	entityMgr facade.EntityMgr
}

func (a *StorageSystemImpl) Init(ctx context.Context, eMgr facade.EntityMgr) {
	// 为可保存对象添加能力
	// 为可保存对象添加定时器能力
	eMgr.RegisterAddProcess("storage", func(ctx context.Context, owner facade.Entity) {
		abi := &StorageAbility{}
		abi.Attach(ctx, owner)
	})
	// 可保存对象创建时保存数据库
	eMgr.RegisterCreateProcess("storage", func(ctx context.Context, owner facade.Entity) {
		abi := owner.GetAbility("storage").(*StorageAbility)
		abi.Save(ctx)
	})
}
