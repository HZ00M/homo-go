package base

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

type ttlSaveEntity struct {
	BaseEntity
	saves atomic.Int32
}

func (m *ttlSaveEntity) Init(ctx context.Context) error {
	m.SetTypeName("user")
	return m.BaseEntity.Init(ctx)
}

func (m *ttlSaveEntity) Save(ctx context.Context) error {
	m.saves.Add(1)
	m.SetDirty(false)
	return nil
}

func (m *ttlSaveEntity) AutoSetDirty() bool { return false }

// 确保实现Saveable接口
var _ facade.Saveable = (*ttlSaveEntity)(nil)

func TestMemoryManager_TTLUnload(t *testing.T) {
	mgr := NewMemoryManager(
		WithCacheTTLMillis(200),
		WithSavePeriodMillis(0),
	)
	ctx := context.Background()

	// 添加调试信息
	t.Logf("MemoryManager created with TTL: %dms, SavePeriod: %dms",
		mgr.opts.CacheTTLMillis, mgr.opts.SavePeriodMillis)

	_, err := mgr.Create(ctx, "user", "U1", func() facade.Entity { return &ttlSaveEntity{} })
	if err != nil {
		t.Fatalf("create err: %v", err)
	}

	// 检查实体是否创建成功
	exists, _ := mgr.Exists(ctx, "user", "U1")
	t.Logf("Entity created, exists: %v", exists)

	// 等待一小段时间让后台任务有机会处理
	time.Sleep(100 * time.Millisecond)

	// 检查分桶状态
	t.Logf("TTL buckets count: %d, current index: %d", len(mgr.ttlBuckets), mgr.ttlIdx)
	t.Logf("TTL index by key count: %d", len(mgr.ttlIndexByKey))

	// 检查实体的具体分桶位置
	key := "user/U1"
	if slot, ok := mgr.ttlIndexByKey[key]; ok {
		t.Logf("Entity %s is in TTL bucket %d", key, slot)
		// 检查该槽位是否有实体
		bucket := mgr.ttlBuckets[slot]
		t.Logf("Bucket %d contains %d entities", slot, len(bucket))
	}

	// 检查实体的元数据
	if meta, ok := mgr.meta["user"]["U1"]; ok {
		now := time.Now().UnixMilli()
		t.Logf("Entity meta: lastAccess=%d, now=%d, diff=%dms",
			meta.lastAccessMs, now, now-meta.lastAccessMs)
	}

	// 监控时间轮推进和实体状态
	deadline := time.Now().Add(3 * time.Second)
	lastIdx := mgr.ttlIdx
	for time.Now().Before(deadline) {
		exists, _ := mgr.Exists(ctx, "user", "U1")
		currentIdx := mgr.ttlIdx

		// 如果时间轮推进了，记录日志
		if currentIdx != lastIdx {
			t.Logf("Time wheel advanced: %d -> %d at %v", lastIdx, currentIdx, time.Now())
			lastIdx = currentIdx
		}

		if !exists {
			t.Logf("Entity unloaded at: %v", time.Now())
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("entity not unloaded by TTL")
}

func TestMemoryManager_SavePeriodAndDirty(t *testing.T) {
	mgr := NewMemoryManager(
		WithCacheTTLMillis(0),
		WithSavePeriodMillis(150),
	)
	ctx := context.Background()

	// 添加调试信息
	t.Logf("MemoryManager created with TTL: %dms, SavePeriod: %dms",
		mgr.opts.CacheTTLMillis, mgr.opts.SavePeriodMillis)

	ent, err := mgr.Create(ctx, "user", "U2", func() facade.Entity { return &ttlSaveEntity{} })
	if err != nil {
		t.Fatalf("create err: %v", err)
	}
	se := ent.(*ttlSaveEntity)

	// 设置实体为脏状态以触发保存
	se.SetDirty(true)
	t.Logf("Entity created and set dirty, saves count: %d", se.saves.Load())

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		saveCount := se.saves.Load()
		if saveCount >= 1 {
			t.Logf("Save triggered, saves count: %d", saveCount)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("periodic save not triggered, final saves count: %d", se.saves.Load())
}

func TestMemoryManager_ApplyOptions_Dynamic(t *testing.T) {
	mgr := NewMemoryManager(
		WithCacheTTLMillis(0),
		WithSavePeriodMillis(0),
	)
	ctx := context.Background()

	// 添加调试信息
	t.Logf("MemoryManager created with TTL: %dms, SavePeriod: %dms",
		mgr.opts.CacheTTLMillis, mgr.opts.SavePeriodMillis)

	_, err := mgr.Create(ctx, "user", "U3", func() facade.Entity { return &ttlSaveEntity{} })
	if err != nil {
		t.Fatalf("create err: %v", err)
	}

	// 检查实体是否创建成功
	exists, _ := mgr.Exists(ctx, "user", "U3")
	t.Logf("Entity created, exists: %v", exists)

	mgr.ApplyOptions(WithCacheTTLMillis(150))
	t.Logf("Applied options, new TTL: %dms", mgr.opts.CacheTTLMillis)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		exists, _ := mgr.Exists(ctx, "user", "U3")
		if !exists {
			t.Logf("Entity unloaded after ApplyOptions at: %v", time.Now())
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("TTL unload not active after ApplyOptions")
}

func TestMemoryManager_BackgroundTasks(t *testing.T) {
	// 测试后台任务是否正确启动
	mgr := NewMemoryManager(
		WithCacheTTLMillis(1000),
		WithSavePeriodMillis(1000),
	)

	// 检查配置是否正确应用
	if mgr.opts.CacheTTLMillis != 1000 {
		t.Errorf("Expected CacheTTLMillis 1000, got %d", mgr.opts.CacheTTLMillis)
	}
	if mgr.opts.SavePeriodMillis != 1000 {
		t.Errorf("Expected SavePeriodMillis 1000, got %d", mgr.opts.SavePeriodMillis)
	}

	// 检查时间轮是否正确初始化
	if len(mgr.ttlBuckets) == 0 {
		t.Error("TTL buckets not initialized")
	}
	if len(mgr.saveBuckets) == 0 {
		t.Error("Save buckets not initialized")
	}

	// 检查ticker是否启动
	if mgr.ttlTicker == nil {
		t.Error("TTL ticker not started")
	}
	if mgr.saveTicker == nil {
		t.Error("Save ticker not started")
	}

	t.Logf("MemoryManager background tasks status: TTL=%v, Save=%v",
		mgr.ttlTicker != nil, mgr.saveTicker != nil)
}
