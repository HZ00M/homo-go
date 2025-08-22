package storage

import "context"

// Driver 底层存储驱动接口（可插拔）
type Driver interface {
	Get(ctx context.Context, typeName, id string) (payload []byte, schema int, err error)
	Put(ctx context.Context, typeName, id string, payload []byte, schema int) error
	Exists(ctx context.Context, typeName, id string) (bool, error)
}
