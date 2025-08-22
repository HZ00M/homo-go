package base

import (
	"context"
	"testing"

	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

type mmEntity struct {
	BaseEntity
}

func (m *mmEntity) Init(ctx context.Context) error {
	m.SetTypeName("user")
	return m.BaseEntity.Init(ctx)
}

// 实现 Saveable 以测试 IsAllLanded
func (m *mmEntity) Save(ctx context.Context) error { m.SetDirty(false); return nil }
func (m *mmEntity) AutoSetDirty() bool             { return true }

func TestMemoryManager_Basic(t *testing.T) {
	mgr := NewMemoryManager()
	ctx := context.Background()

	// 不存在
	exists, _ := mgr.Exists(ctx, "user", "U1")
	if exists {
		t.Fatalf("should not exist")
	}

	// Create
	_, err := mgr.Create(ctx, "user", "U1", func() facade.Entity { return &mmEntity{} })
	if err != nil {
		t.Fatalf("create err: %v", err)
	}

	// Exists
	exists, _ = mgr.Exists(ctx, "user", "U1")
	if !exists {
		t.Fatalf("should exist")
	}

	// Get
	e, err := mgr.Get(ctx, "user", "U1")
	if err != nil || e.ID() != "U1" {
		t.Fatalf("get err: %v", err)
	}

	// Remove
	_ = mgr.Remove(ctx, "user", "U1")
	exists, _ = mgr.Exists(ctx, "user", "U1")
	if exists {
		t.Fatalf("should not exist after remove")
	}
}

func TestMemoryManager_NotFoundHook(t *testing.T) {
	mgr := NewMemoryManager()
	ctx := context.Background()
	mgr.RegisterNotFoundHook("user", func(ctx context.Context, id string) (facade.Entity, error) {
		return &mmEntity{}, nil
	})
	_, err := mgr.Get(ctx, "user", "U2")
	if err != nil {
		t.Fatalf("notfound hook err: %v", err)
	}
}

func TestMemoryManager_ReleaseAll_IsAllLanded(t *testing.T) {
	mgr := NewMemoryManager()
	ctx := context.Background()
	e, _ := mgr.Create(ctx, "user", "U3", func() facade.Entity { return &mmEntity{} })
	if s, ok := e.(interface{ SetDirty(bool) }); ok {
		s.SetDirty(true)
	}
	ok, _ := mgr.IsAllLanded(ctx)
	if ok {
		t.Fatalf("should not landed when dirty")
	}
	// 清空
	_ = mgr.ReleaseAll(ctx)
	ok, _ = mgr.IsAllLanded(ctx)
	if !ok {
		t.Fatalf("should landed after release all")
	}
}
