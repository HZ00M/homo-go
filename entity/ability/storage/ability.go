package storage

import (
	"context"

	"github.com/go-kratos/kratos/v2/drivers/storage"
	"github.com/go-kratos/kratos/v2/entity/base"
	"github.com/go-kratos/kratos/v2/entity/facade"
)

// StorageAbility 提供最小持久化能力：将实体序列化后写入驱动
// 方法 Save(ctx) ([]byte, error) 可通过 Dispatcher 分发

type StorageAbility struct {
	base.BaseAbility
	driver   storage.Driver
	typeName string
	owner    facade.Entity
}

func NewStorageAbility(driver storage.Driver, typeName string) *StorageAbility {
	// 绑定自身以便 BaseAbility.Attach 能将正确实例注册到实体
	a := &StorageAbility{driver: driver, typeName: typeName}
	a.BaseAbility.Bind(a)
	return a
}

func (a *StorageAbility) Name() string { return "storage" }

// Save 将实体以 SaveObject 形式序列化并写入底层存储
// 返回固定 "ok" 字节以便测试
func (a *StorageAbility) Save(ctx context.Context) ([]byte, error) {
	if a.owner == nil {
		return nil, facade.ErrNotFound
	}
	so, ok := a.owner.(facade.SaveObject)
	if !ok {
		return nil, facade.ErrEncode
	}
	payload, err := so.MarshalBinary()
	if err != nil {
		return nil, err
	}
	schema := so.GetSchemaVersion()
	if err := a.driver.Put(ctx, a.typeName, a.owner.ID(), payload, schema); err != nil {
		return nil, err
	}
	return []byte("ok"), nil
}
