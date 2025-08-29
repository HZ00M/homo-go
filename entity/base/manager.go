package base

import (
	"context"
	"sync"
	"time"

	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

// MemoryManager 提供最小可用的内存版 EntityMgr 实现
// 功能点：
// - 内存索引：type -> id -> entity
// - NotFoundHook：miss 时按 type 构建
// - ReleaseAll：清空内存实体
// - IsAllLanded：若存在 SaveAble 实体未落地（脏），返回 false，否则 true
// - TTL 卸载与周期保存（可选，基于 Option 配置）

type MemoryManager struct {
	mu            sync.RWMutex
	entities      map[string]map[string]facade.Entity // type -> id -> entity
	notFoundHooks map[string]func(ctx context.Context, id string) (facade.Entity, error)

	// ability hooks
	createProcesses   map[string][]func(ctx context.Context, e facade.Entity)
	addProcesses      map[string][]func(ctx context.Context, e facade.Entity)
	getProcesses      map[string][]func(ctx context.Context, e facade.Entity)
	removeProcesses   map[string][]func(ctx context.Context, e facade.Entity)
	beforeAddHandlers map[string][]func(ctx context.Context, e facade.Entity) (facade.Ability, error)

	// on remove callback (optional)
	onEntityRemoved func(entityType, id string)

	// key-level locker to merge concurrent operations on same type/id
	keyMu    sync.Mutex
	keyLocks map[string]*sync.Mutex

	// --- 以下为 TTL/Save 配置与运行时元数据 ---
	opts mmOptions

	// meta: 记录每个实体的最近访问与保存时间
	meta map[string]map[string]entityMeta // type -> id -> meta

	// 分桶时间轮（固定槽位，键为 type/id）
	ttlBuckets     []map[string]struct{}
	saveBuckets    []map[string]struct{}
	ttlIndexByKey  map[string]int // key->slot
	saveIndexByKey map[string]int // key->slot
	ttlIdx         int
	saveIdx        int

	// 后台推进
	ttlTicker  *time.Ticker
	saveTicker *time.Ticker
	stopCh     chan struct{}
}

// --- Options 定义 ---

type mmOptions struct {
	KeepAliveOnGet   bool
	CacheTTLMillis   int64
	SavePeriodMillis int64
	DestroyOnUnload  bool
	BucketSlots      int
}

type Option func(*mmOptions)

func WithKeepAliveOnGet(v bool) Option    { return func(o *mmOptions) { o.KeepAliveOnGet = v } }
func WithCacheTTLMillis(v int64) Option   { return func(o *mmOptions) { o.CacheTTLMillis = v } }
func WithSavePeriodMillis(v int64) Option { return func(o *mmOptions) { o.SavePeriodMillis = v } }
func WithDestroyOnUnload(v bool) Option   { return func(o *mmOptions) { o.DestroyOnUnload = v } }
func WithBucketSlots(v int) Option        { return func(o *mmOptions) { o.BucketSlots = v } }

func defaultMMOptions() mmOptions {
	return mmOptions{
		KeepAliveOnGet:   true,
		CacheTTLMillis:   0,
		SavePeriodMillis: 0,
		DestroyOnUnload:  false,
		BucketSlots:      60,
	}
}

// NewMemoryManager 创建 MemoryManager（支持 Option）
func NewMemoryManager(opts ...Option) *MemoryManager {
	opt := defaultMMOptions()
	for _, fn := range opts {
		fn(&opt)
	}

	m := &MemoryManager{
		entities:          make(map[string]map[string]facade.Entity),
		notFoundHooks:     make(map[string]func(ctx context.Context, id string) (facade.Entity, error)),
		createProcesses:   make(map[string][]func(ctx context.Context, e facade.Entity)),
		addProcesses:      make(map[string][]func(ctx context.Context, e facade.Entity)),
		getProcesses:      make(map[string][]func(ctx context.Context, e facade.Entity)),
		removeProcesses:   make(map[string][]func(ctx context.Context, e facade.Entity)),
		beforeAddHandlers: make(map[string][]func(ctx context.Context, e facade.Entity) (facade.Ability, error)),
		keyLocks:          make(map[string]*sync.Mutex),
		opts:              opt,
		meta:              make(map[string]map[string]entityMeta),
		ttlIndexByKey:     make(map[string]int),
		saveIndexByKey:    make(map[string]int),
		stopCh:            make(chan struct{}),
	}
	m.initBucketsLocked()
	m.startBackgroundLocked()
	return m
}

// ApplyOptions 动态更新配置（线程安全）：启停后台任务并在必要时重建桶
func (m *MemoryManager) ApplyOptions(opts ...Option) {
	m.mu.Lock()
	defer m.mu.Unlock()
	opt := m.opts
	for _, fn := range opts {
		fn(&opt)
	}
	needRebuild := opt.BucketSlots <= 0 || opt.BucketSlots != m.opts.BucketSlots
	m.opts = opt
	if needRebuild {
		m.resetBucketsLocked()
	}
	m.restartBackgroundLocked()

	// 重新将已存在的实体入桶以应用新的配置
	m.rebucketExistingEntitiesLocked()
}

// SetOnEntityRemoved 注入实体移除回调（可选）
func (m *MemoryManager) SetOnEntityRemoved(fn func(entityType, id string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onEntityRemoved = fn
}

func makeKey(entityType, id string) string { return entityType + "/" + id }

func (m *MemoryManager) lockKey(entityType, id string) *sync.Mutex {
	key := makeKey(entityType, id)
	m.keyMu.Lock()
	lk, ok := m.keyLocks[key]
	if !ok {
		lk = &sync.Mutex{}
		m.keyLocks[key] = lk
	}
	m.keyMu.Unlock()
	lk.Lock()
	return lk
}

func (m *MemoryManager) unlockKey(lk *sync.Mutex) {
	lk.Unlock()
}

// Exists: 是否存在（内存或存储）。当前仅内存
func (m *MemoryManager) Exists(ctx context.Context, entityType, id string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if mm := m.entities[entityType]; mm != nil {
		_, ok := mm[id]
		return ok, nil
	}
	return false, nil
}

// Create: 创建并初始化实体（不落地）
func (m *MemoryManager) Create(ctx context.Context, entityType, id string, ctor func() facade.Entity) (facade.Entity, error) {
	// per-key lock to avoid duplicate concurrent create
	lk := m.lockKey(entityType, id)
	defer m.unlockKey(lk)

	inst := ctor()
	if inst == nil {
		return nil, facade.ErrNotFound
	}
	inst.SetID(id)
	if err := inst.Init(ctx); err != nil {
		return nil, err
	}
	// before-add hooks (lock-free)
	m.runBeforeAdd(ctx, inst)
	// add hooks (lock-free)
	m.runAdd(ctx, inst)

	m.mu.Lock()
	mm := m.entities[entityType]
	if mm == nil {
		mm = make(map[string]facade.Entity)
		m.entities[entityType] = mm
	}
	mm[id] = inst
	m.onCreatedLocked(entityType, id)
	m.mu.Unlock()

	// create hooks (lock-free)
	m.runCreate(ctx, inst)
	return inst, nil
}

// Get: 获取实体（未命中尝试 NotFoundHook）
func (m *MemoryManager) Get(ctx context.Context, entityType, id string) (facade.Entity, error) {
	m.mu.RLock()
	if mm := m.entities[entityType]; mm != nil {
		if e, ok := mm[id]; ok {
			m.mu.RUnlock()
			// get hooks on hit (lock-free)
			m.runGet(ctx, e)
			m.touchOnGet(entityType, id)
			return e, nil
		}
	}
	m.mu.RUnlock()
	//todo 从storage中加载
	// miss: merge concurrent load/create via key-level lock
	return m.loadOrCreateWithHook(ctx, entityType, id)
}

// GetNoKeepAlive: 获取实体（不续期、不置脏、不触发 get 流程钩子）
func (m *MemoryManager) GetNoKeepAlive(ctx context.Context, entityType, id string) (facade.Entity, error) {
	m.mu.RLock()
	if mm := m.entities[entityType]; mm != nil {
		if e, ok := mm[id]; ok {
			m.mu.RUnlock()
			return e, nil
		}
	}
	m.mu.RUnlock()
	// miss path same as Get but without get hook on hit, which we already avoided
	return m.loadOrCreateWithHook(ctx, entityType, id)
}

// GetOrCreate: 获取或创建
func (m *MemoryManager) GetOrCreate(ctx context.Context, entityType, id string, ctor func() facade.Entity) (facade.Entity, error) {
	if e, err := m.Get(ctx, entityType, id); err == nil {
		return e, nil
	}
	return m.Create(ctx, entityType, id, ctor)
}

// Remove: 仅移出内存（不删除持久化数据）
func (m *MemoryManager) Remove(ctx context.Context, entityType, id string) error {
	removed, onRemoved := m.removeInternal(ctx, entityType, id)
	if removed != nil {
		// remove hooks (lock-free)
		m.runRemove(ctx, removed)
		// upper-layer cleanup (optional)
		if onRemoved != nil {
			onRemoved(entityType, id)
		}
	}
	return nil
}

// DestroyAllType: 同一 id 的不同 type 都销毁（仅内存）
func (m *MemoryManager) DestroyAllType(ctx context.Context, id string) error {
	var removedList []struct {
		entityType string
		entity     facade.Entity
	}
	m.mu.Lock()
	for t, mm := range m.entities {
		if e, ok := mm[id]; ok {
			removedList = append(removedList, struct {
				entityType string
				entity     facade.Entity
			}{entityType: t, entity: e})
			delete(mm, id)
			m.entities[t] = mm
			m.onRemovedLocked(t, id)
		}
	}
	onRemoved := m.onEntityRemoved
	m.mu.Unlock()

	for _, it := range removedList {
		m.runRemove(ctx, it.entity)
		if onRemoved != nil {
			onRemoved(it.entityType, id)
		}
	}
	return nil
}

// ReleaseAll: 释放全部实体（等待落地与路由缓存过期）。当前清空内存并按默认仅回调策略清理
func (m *MemoryManager) ReleaseAll(ctx context.Context) error {
	m.mu.Lock()
	old := m.entities
	m.entities = make(map[string]map[string]facade.Entity)
	onRemoved := m.onEntityRemoved
	// 清理所有 meta 与桶
	m.meta = make(map[string]map[string]entityMeta)
	m.resetBucketsLocked()
	m.mu.Unlock()

	// 默认策略：不触发 remove hooks，不调用 Destroy，仅上层回调
	if onRemoved != nil {
		for t, mm := range old {
			for id := range mm {
				onRemoved(t, id)
			}
		}
	}
	return nil
}

// IsAllLanded: 是否所有可存储实体已落地。
// 当前遍历内存实体，若发现实现了 SaveAble 且 IsDirty() 为 true，则返回 false
func (m *MemoryManager) IsAllLanded(ctx context.Context) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, mm := range m.entities {
		for _, e := range mm {
			if s, ok := e.(facade.SaveAble); ok {
				if s.IsDirty() {
					return false, nil
				}
			}
		}
	}
	return true, nil
}

// RegisterNotFoundHook: miss 钩子
func (m *MemoryManager) RegisterNotFoundHook(entityType string, fn func(ctx context.Context, id string) (facade.Entity, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notFoundHooks[entityType] = fn
}

// RegisterCreateProcess: 注册按能力的创建流程钩子
func (m *MemoryManager) RegisterCreateProcess(abilityName string, consumer func(ctx context.Context, e facade.Entity)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createProcesses[abilityName] = append(m.createProcesses[abilityName], consumer)
}

// RegisterAddProcess: 注册按能力的添加流程钩子
func (m *MemoryManager) RegisterAddProcess(abilityName string, consumer func(ctx context.Context, e facade.Entity)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addProcesses[abilityName] = append(m.addProcesses[abilityName], consumer)
}

// RegisterGetProcess: 注册按能力的获取流程钩子
func (m *MemoryManager) RegisterGetProcess(abilityName string, consumer func(ctx context.Context, e facade.Entity)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getProcesses[abilityName] = append(m.getProcesses[abilityName], consumer)
}

// RegisterRemoveProcess: 注册按能力的移除流程钩子
func (m *MemoryManager) RegisterRemoveProcess(abilityName string, consumer func(ctx context.Context, e facade.Entity)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeProcesses[abilityName] = append(m.removeProcesses[abilityName], consumer)
}

// RegisterBeforeAddProcess: 注册按能力的前置添加流程钩子
func (m *MemoryManager) RegisterBeforeAddProcess(abilityName string, fun func(ctx context.Context, e facade.Entity) (facade.Ability, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.beforeAddHandlers[abilityName] = append(m.beforeAddHandlers[abilityName], fun)
}

// internal helper: miss path loader with NotFoundHook under key lock
func (m *MemoryManager) loadOrCreateWithHook(ctx context.Context, entityType, id string) (facade.Entity, error) {
	lk := m.lockKey(entityType, id)
	defer m.unlockKey(lk)

	// double-check after acquiring key lock
	m.mu.RLock()
	if mm := m.entities[entityType]; mm != nil {
		if e, ok := mm[id]; ok {
			m.mu.RUnlock()
			return e, nil
		}
	}
	hook := m.notFoundHooks[entityType]
	m.mu.RUnlock()
	if hook == nil {
		return nil, facade.ErrNotFound
	}

	inst, err := hook(ctx, id)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, facade.ErrNotFound
	}
	inst.SetID(id)
	if err := inst.Init(ctx); err != nil {
		return nil, err
	}
	// before-add hooks (lock-free)
	m.runBeforeAdd(ctx, inst)
	// add hooks (lock-free)
	m.runAdd(ctx, inst)

	m.mu.Lock()
	mm := m.entities[entityType]
	if mm == nil {
		mm = make(map[string]facade.Entity)
		m.entities[entityType] = mm
	}
	mm[id] = inst
	m.onCreatedLocked(entityType, id)
	m.mu.Unlock()
	return inst, nil
}

// internal hook runners
func (m *MemoryManager) runCreate(ctx context.Context, e facade.Entity) {
	m.mu.RLock()
	var fns []func(ctx context.Context, e facade.Entity)
	for _, list := range m.createProcesses {
		fns = append(fns, list...)
	}
	m.mu.RUnlock()
	for _, fn := range fns {
		fn(ctx, e)
	}
}

func (m *MemoryManager) runAdd(ctx context.Context, e facade.Entity) {
	m.mu.RLock()
	var fns []func(ctx context.Context, e facade.Entity)
	for _, list := range m.addProcesses {
		fns = append(fns, list...)
	}
	m.mu.RUnlock()
	for _, fn := range fns {
		fn(ctx, e)
	}
}

func (m *MemoryManager) runGet(ctx context.Context, e facade.Entity) {
	m.mu.RLock()
	var fns []func(ctx context.Context, e facade.Entity)
	for _, list := range m.getProcesses {
		fns = append(fns, list...)
	}
	m.mu.RUnlock()
	for _, fn := range fns {
		fn(ctx, e)
	}
}

func (m *MemoryManager) runRemove(ctx context.Context, e facade.Entity) {
	m.mu.RLock()
	var fns []func(ctx context.Context, e facade.Entity)
	for _, list := range m.removeProcesses {
		fns = append(fns, list...)
	}
	m.mu.RUnlock()
	for _, fn := range fns {
		fn(ctx, e)
	}
}

func (m *MemoryManager) runBeforeAdd(ctx context.Context, e facade.Entity) {
	m.mu.RLock()
	var fns []func(ctx context.Context, e facade.Entity) (facade.Ability, error)
	for _, list := range m.beforeAddHandlers {
		fns = append(fns, list...)
	}
	m.mu.RUnlock()
	for _, fn := range fns {
		abi, err := fn(ctx, e)
		if err != nil {
			continue
		}
		if abi != nil {
			_ = abi.Attach(ctx, e)
		}
	}
}

// --- 以下为 TTL/Save 内部实现 ---

type entityMeta struct {
	lastAccessMs int64
	lastSaveMs   int64
}

func (m *MemoryManager) initBucketsLocked() {
	if m.opts.BucketSlots <= 0 {
		m.opts.BucketSlots = 60
	}
	m.ttlBuckets = make([]map[string]struct{}, m.opts.BucketSlots)
	for i := range m.ttlBuckets {
		m.ttlBuckets[i] = make(map[string]struct{})
	}
	m.saveBuckets = make([]map[string]struct{}, m.opts.BucketSlots)
	for i := range m.saveBuckets {
		m.saveBuckets[i] = make(map[string]struct{})
	}
	m.ttlIdx = 0
	m.saveIdx = 0
	m.ttlIndexByKey = make(map[string]int)
	m.saveIndexByKey = make(map[string]int)
}

func (m *MemoryManager) resetBucketsLocked() {
	m.ttlBuckets = nil
	m.saveBuckets = nil
	m.ttlIndexByKey = make(map[string]int)
	m.saveIndexByKey = make(map[string]int)
	m.initBucketsLocked()
}

// 在创建实体后初始化 meta 并入桶
func (m *MemoryManager) onCreatedLocked(entityType, id string) {
	if m.meta[entityType] == nil {
		m.meta[entityType] = make(map[string]entityMeta)
	}
	now := time.Now().UnixMilli()
	m.meta[entityType][id] = entityMeta{lastAccessMs: now, lastSaveMs: 0}
	m.rebucketTTLLocked(entityType, id, now)
	m.rebucketSaveLocked(entityType, id, now)
}

// 在移除实体后清理 meta 与桶
func (m *MemoryManager) onRemovedLocked(entityType, id string) {
	if mp := m.meta[entityType]; mp != nil {
		delete(mp, id)
		if len(mp) == 0 {
			delete(m.meta, entityType)
		}
	}
	key := makeKey(entityType, id)
	if idx, ok := m.ttlIndexByKey[key]; ok {
		delete(m.ttlBuckets[idx], key)
		delete(m.ttlIndexByKey, key)
	}
	if idx, ok := m.saveIndexByKey[key]; ok {
		delete(m.saveBuckets[idx], key)
		delete(m.saveIndexByKey, key)
	}
}

func (m *MemoryManager) touchOnGet(entityType, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.opts.KeepAliveOnGet || m.opts.CacheTTLMillis == 0 {
		return
	}
	if _, ok := m.entities[entityType]; !ok {
		return
	}
	if _, ok := m.entities[entityType][id]; !ok {
		return
	}
	now := time.Now().UnixMilli()
	mp := m.meta[entityType]
	if mp == nil {
		mp = make(map[string]entityMeta)
		m.meta[entityType] = mp
	}
	meta := mp[id]
	meta.lastAccessMs = now
	mp[id] = meta
	m.rebucketTTLLocked(entityType, id, now)
}

func (m *MemoryManager) rebucketTTLLocked(entityType, id string, nowMs int64) {
	if m.opts.CacheTTLMillis == 0 || len(m.ttlBuckets) == 0 {
		return
	}
	key := makeKey(entityType, id)
	// 移出旧槽
	if old, ok := m.ttlIndexByKey[key]; ok {
		delete(m.ttlBuckets[old], key)
	}
	// 计算到期时间（毫秒）
	expireMs := nowMs + m.opts.CacheTTLMillis
	// 计算到期时间对应的秒数
	expireSec := expireMs / 1000
	// 计算当前时间对应的秒数
	nowSec := nowMs / 1000
	// 计算目标槽位（相对于当前时间的偏移，以秒为单位）
	offsetSec := expireSec - nowSec
	targetSlot := (m.ttlIdx + int(offsetSec)) % len(m.ttlBuckets)
	m.ttlBuckets[targetSlot][key] = struct{}{}
	m.ttlIndexByKey[key] = targetSlot
}

func (m *MemoryManager) rebucketSaveLocked(entityType, id string, nowMs int64) {
	if m.opts.SavePeriodMillis == 0 || len(m.saveBuckets) == 0 {
		return
	}
	key := makeKey(entityType, id)
	if old, ok := m.saveIndexByKey[key]; ok {
		delete(m.saveBuckets[old], key)
	}
	// 计算下次保存时间（毫秒）
	nextSaveMs := nowMs + m.opts.SavePeriodMillis
	// 计算下次保存时间对应的秒数
	nextSaveSec := nextSaveMs / 1000
	// 计算当前时间对应的秒数
	nowSec := nowMs / 1000
	// 计算目标槽位（相对于当前时间的偏移，以秒为单位）
	offsetSec := nextSaveSec - nowSec
	targetSlot := (m.saveIdx + int(offsetSec)) % len(m.saveBuckets)
	m.saveBuckets[targetSlot][key] = struct{}{}
	m.saveIndexByKey[key] = targetSlot
}

func (m *MemoryManager) startBackgroundLocked() {
	// TTL 扫描
	if m.opts.CacheTTLMillis > 0 && m.ttlTicker == nil {
		m.ttlTicker = time.NewTicker(time.Second)
		go m.runTTL()
	}
	// Save 扫描
	if m.opts.SavePeriodMillis > 0 && m.saveTicker == nil {
		m.saveTicker = time.NewTicker(time.Second)
		go m.runSave()
	}
}

func (m *MemoryManager) stopBackgroundLocked() {
	if m.ttlTicker != nil {
		m.ttlTicker.Stop()
		m.ttlTicker = nil
	}
	if m.saveTicker != nil {
		m.saveTicker.Stop()
		m.saveTicker = nil
	}
	select {
	case m.stopCh <- struct{}{}:
	default:
	}
}

func (m *MemoryManager) restartBackgroundLocked() {
	m.stopBackgroundLocked()
	m.startBackgroundLocked()
}

func (m *MemoryManager) runTTL() {
	for {
		select {
		case <-m.stopCh:
			return
		case <-m.ttlTicker.C:
			m.scanTTLOnce()
		}
	}
}

func (m *MemoryManager) runSave() {
	for {
		select {
		case <-m.stopCh:
			return
		case <-m.saveTicker.C:
			m.scanSaveOnce()
		}
	}
}

func (m *MemoryManager) scanTTLOnce() {
	now := time.Now().UnixMilli()
	var keys []string
	m.mu.Lock()
	if len(m.ttlBuckets) == 0 {
		m.mu.Unlock()
		return
	}
	slot := m.ttlIdx % len(m.ttlBuckets)
	for k := range m.ttlBuckets[slot] {
		keys = append(keys, k)
	}
	// 清空当前槽
	m.ttlBuckets[slot] = make(map[string]struct{})
	for _, k := range keys {
		delete(m.ttlIndexByKey, k)
	}
	m.ttlIdx = (m.ttlIdx + 1) % len(m.ttlBuckets)
	m.mu.Unlock()

	for _, key := range keys {
		// 解析 type/id
		// key 格式: type/id
		var entityType, id string
		for i := 0; i < len(key); i++ {
			if key[i] == '/' {
				entityType = key[:i]
				id = key[i+1:]
				break
			}
		}
		m.mu.RLock()
		meta, okMeta := m.meta[entityType][id]
		m.mu.RUnlock()
		if !okMeta {
			continue
		}
		if now-meta.lastAccessMs < m.opts.CacheTTLMillis {
			// 未到期，重排
			m.mu.Lock()
			m.rebucketTTLLocked(entityType, id, now)
			m.mu.Unlock()
			continue
		}
		// 到期：卸载（复用 Remove 流程），可选 Destroy
		ctx := context.Background()
		removed, onRemoved := m.removeInternal(ctx, entityType, id)
		if removed != nil {
			m.runRemove(ctx, removed)
			if onRemoved != nil {
				onRemoved(entityType, id)
			}
			if m.opts.DestroyOnUnload {
				_ = removed.Destroy(ctx)
			}
		}
	}
}

func (m *MemoryManager) scanSaveOnce() {
	now := time.Now().UnixMilli()
	var keys []string
	m.mu.Lock()
	if len(m.saveBuckets) == 0 {
		m.mu.Unlock()
		return
	}
	slot := m.saveIdx % len(m.saveBuckets)
	for k := range m.saveBuckets[slot] {
		keys = append(keys, k)
	}
	m.saveBuckets[slot] = make(map[string]struct{})
	for _, k := range keys {
		delete(m.saveIndexByKey, k)
	}
	m.saveIdx = (m.saveIdx + 1) % len(m.saveBuckets)
	m.mu.Unlock()

	for _, key := range keys {
		var entityType, id string
		for i := 0; i < len(key); i++ {
			if key[i] == '/' {
				entityType = key[:i]
				id = key[i+1:]
				break
			}
		}
		m.mu.RLock()
		var ent facade.Entity
		if mm := m.entities[entityType]; mm != nil {
			ent = mm[id]
		}
		meta := m.meta[entityType][id]
		m.mu.RUnlock()
		if ent == nil {
			continue
		}
		s, ok := ent.(facade.SaveAble)
		if !ok {
			// 非 SaveAble：仅重排
			m.mu.Lock()
			m.rebucketSaveLocked(entityType, id, now)
			m.mu.Unlock()
			continue
		}
		needSave := s.IsDirty() || (m.opts.SavePeriodMillis > 0 && now-meta.lastSaveMs >= m.opts.SavePeriodMillis)
		if !needSave {
			m.mu.Lock()
			m.rebucketSaveLocked(entityType, id, now)
			m.mu.Unlock()
			continue
		}
		if err := s.Save(context.Background()); err == nil {
			// 成功：清脏并更新 lastSaveMs，再重排
			s.SetDirty(false)
			m.mu.Lock()
			mp := m.meta[entityType]
			if mp == nil {
				mp = make(map[string]entityMeta)
				m.meta[entityType] = mp
			}
			md := mp[id]
			md.lastSaveMs = now
			mp[id] = md
			m.rebucketSaveLocked(entityType, id, now)
			m.mu.Unlock()
		} else {
			// 失败：下一轮重试（重排即可）
			m.mu.Lock()
			m.rebucketSaveLocked(entityType, id, now)
			m.mu.Unlock()
		}
	}
}

// removeInternal: 在锁内删除并返回被移除实体与上层回调
func (m *MemoryManager) removeInternal(ctx context.Context, entityType, id string) (facade.Entity, func(entityType, id string)) {
	var removed facade.Entity
	m.mu.Lock()
	if mm := m.entities[entityType]; mm != nil {
		if e, ok := mm[id]; ok {
			removed = e
		}
		delete(mm, id)
		m.entities[entityType] = mm
	}
	m.onRemovedLocked(entityType, id)
	onRemoved := m.onEntityRemoved
	m.mu.Unlock()
	return removed, onRemoved
}

// rebucketExistingEntitiesLocked 将已存在的实体重新入桶以应用新的TTL和保存配置
func (m *MemoryManager) rebucketExistingEntitiesLocked() {
	for entityType, mm := range m.entities {
		for id := range mm {
			if meta, ok := m.meta[entityType][id]; ok {
				// 重新入TTL桶
				if m.opts.CacheTTLMillis > 0 {
					m.rebucketTTLLocked(entityType, id, meta.lastAccessMs)
				}
				// 重新入保存桶
				if m.opts.SavePeriodMillis > 0 {
					m.rebucketSaveLocked(entityType, id, meta.lastSaveMs)
				}
			}
		}
	}
}

var _ facade.EntityMgr = (*MemoryManager)(nil)
