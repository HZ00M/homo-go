package storage

import (
	"context"
	"sync"
)

// MemoryDriver 是最小可用的内存版存储驱动实现
// 仅用于本地开发与单测
// 数据结构：typeName -> id -> {payload, schema}

type MemoryDriver struct {
	mu   sync.RWMutex
	data map[string]map[string]struct {
		payload []byte
		schema  int
	}
}

// NewMemoryDriver 创建内存驱动实例
func NewMemoryDriver() *MemoryDriver {
	return &MemoryDriver{data: make(map[string]map[string]struct {
		payload []byte
		schema  int
	})}
}

// Get 按类型与ID读取序列化数据
func (m *MemoryDriver) Get(ctx context.Context, typeName, id string) (payload []byte, schema int, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mm := m.data[typeName]
	if mm == nil {
		return nil, 0, nil
	}
	v, ok := mm[id]
	if !ok {
		return nil, 0, nil
	}
	return v.payload, v.schema, nil
}

// Put 写入序列化数据
func (m *MemoryDriver) Put(ctx context.Context, typeName, id string, payload []byte, schema int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mm := m.data[typeName]
	if mm == nil {
		mm = make(map[string]struct {
			payload []byte
			schema  int
		})
		m.data[typeName] = mm
	}
	mm[id] = struct {
		payload []byte
		schema  int
	}{payload: append([]byte(nil), payload...), schema: schema}
	return nil
}

// Exists 判断是否存在
func (m *MemoryDriver) Exists(ctx context.Context, typeName, id string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mm := m.data[typeName]
	if mm == nil {
		return false, nil
	}
	_, ok := mm[id]
	return ok, nil
}
