package base

import (
	"context"
	"testing"
)

type dummyAbility struct{ BaseAbility }

func (d *dummyAbility) Name() string { return "dummy" }

func TestBaseEntity_AbilityAndDirty(t *testing.T) {
	ctx := context.Background()
	be := &BaseEntity{}
	be.SetTypeName("user")
	be.SetID("U1")
	if err := be.Init(ctx); err != nil {
		t.Fatalf("init err: %v", err)
	}
	if be.LoadTime() == 0 {
		t.Fatalf("load time should be set")
	}
	// attach ability via ability.Attach flow
	d := &dummyAbility{}
	// bind self so BaseAbility.Attach registers correct instance
	d.BaseAbility.Bind(d)
	_ = d.Attach(ctx, be)
	if be.GetAbility("dummy") == nil {
		t.Fatalf("ability not registered")
	}
	// dirty flag
	if be.IsDirty() {
		t.Fatalf("should be clean initially")
	}
	be.SetDirty(true)
	if !be.IsDirty() || be.DirtyTimeUnixMilli() == 0 {
		t.Fatalf("dirty or dirty time not set")
	}
	be.SetDirty(false)
	if be.IsDirty() {
		t.Fatalf("should be clean after reset")
	}
}
