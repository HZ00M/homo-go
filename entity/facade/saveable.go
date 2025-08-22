package facade

import "context"

type StorageSystem interface {
	Save(ctx context.Context, so SaveObject) error
}

// Storage 存储抽象（实现可为内存、Redis、MySQL 组合）
type Storage interface {
	Load(ctx context.Context, typeName, id string) (Entity, error)
	Save(ctx context.Context, so SaveObject) error
	Exists(ctx context.Context, typeName, id string) (bool, error)
}
type LType int

// SaveObject 序列化契约（由实体实现）
type SaveObject interface {
	LogicType() LType
	OwnerID() string
	MarshalBinary() ([]byte, error)
	UnmarshalBinary([]byte) error
	GetSchemaVersion() int
}
