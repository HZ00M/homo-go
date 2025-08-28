// Package module 提供模块生命周期管理功能
package module

import (
	"context"
	"sort"
	"sync"
	"time"
)

// ModuleManager 模块管理器
type ModuleManager struct {
	mu        sync.RWMutex // 读写锁，保护并发访问
	modules   []Module     // 模块列表
	startTime time.Time    // 启动时间
}

// NewModuleManager 创建新的 ModuleManager 实例
func NewModuleManager(modules []Module) *ModuleManager {
	return &ModuleManager{
		modules:   modules,
		startTime: time.Now(),
	}
}

// SetModules 设置模块列表
func (mm *ModuleManager) SetModules(modules []Module) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.modules = make([]Module, len(modules))
	copy(mm.modules, modules)

	// 按 Order 排序
	mm.sortModules()
}

// GetModules 获取模块列表（已排序）
func (mm *ModuleManager) GetModules() []Module {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	modules := make([]Module, len(mm.modules))
	copy(modules, mm.modules)
	return modules
}

// sortModules 按 Order 排序模块
func (mm *ModuleManager) sortModules() {
	sort.Slice(mm.modules, func(i, j int) bool {
		return mm.modules[i].Order() < mm.modules[j].Order()
	})
}

// InitAll 初始化所有模块
func (mm *ModuleManager) InitAll(ctx context.Context) error {
	mm.mu.Lock()
	mm.startTime = time.Now()
	mm.mu.Unlock()

	if len(mm.modules) == 0 {
		return nil
	}

	// 按 Order 分组初始化
	orderGroups := make(map[int][]Module)
	for _, module := range mm.modules {
		order := module.Order()
		orderGroups[order] = append(orderGroups[order], module)
	}

	// 按 Order 顺序初始化
	orders := make([]int, 0, len(orderGroups))
	for order := range orderGroups {
		orders = append(orders, order)
	}
	sort.Ints(orders)

	for _, order := range orders {
		modules := orderGroups[order]

		if err := mm.initModulesByOrder(ctx, order, modules); err != nil {
			return err
		}
	}

	return nil
}

// initModulesByOrder 初始化指定 Order 的模块组
func (mm *ModuleManager) initModulesByOrder(ctx context.Context, order int, modules []Module) error {
	// 并发初始化相同 Order 的模块
	type initResult struct {
		module Module
		err    error
	}

	resultChan := make(chan initResult, len(modules))

	for _, module := range modules {
		go func(m Module) {
			err := m.Init(ctx)
			resultChan <- initResult{module: m, err: err}
		}(module)
	}

	// 收集结果
	for i := 0; i < len(modules); i++ {
		result := <-resultChan
		if result.err != nil {
			return result.err
		}
	}

	return nil
}

// AfterAllInit 调用所有模块的 AfterAllInit
func (mm *ModuleManager) AfterAllInit(ctx context.Context) {
	mm.mu.RLock()
	modules := make([]Module, len(mm.modules))
	copy(modules, mm.modules)
	mm.mu.RUnlock()

	for _, module := range modules {
		module.AfterAllInit(ctx)
	}
}

// AfterStart 调用所有模块的 AfterStart
func (mm *ModuleManager) AfterStart(ctx context.Context) {
	mm.mu.RLock()
	modules := make([]Module, len(mm.modules))
	copy(modules, mm.modules)
	mm.mu.RUnlock()

	for _, module := range modules {
		module.AfterStart(ctx)
	}
}

// CloseAll 关闭所有模块
func (mm *ModuleManager) CloseAll(ctx context.Context) error {
	mm.mu.RLock()
	modules := make([]Module, len(mm.modules))
	copy(modules, mm.modules)
	mm.mu.RUnlock()

	if len(modules) == 0 {
		return nil
	}

	// 按 Order 反序关闭
	for i := len(modules) - 1; i >= 0; i-- {
		module := modules[i]

		if err := module.BeforeClose(ctx); err != nil {
			return err
		}
	}

	return nil
}

// AfterStop 调用所有模块的 AfterStop
func (mm *ModuleManager) AfterStop(ctx context.Context) {
	mm.mu.RLock()
	modules := make([]Module, len(mm.modules))
	copy(modules, mm.modules)
	mm.mu.RUnlock()

	for _, module := range modules {
		module.AfterStop(ctx)
	}
}

// RegisterToApp 将模块生命周期注册到 Kratos 应用
func (mm *ModuleManager) RegisterToApp(app interface{}) error {
	// 这里需要根据具体的应用类型进行适配
	// 暂时返回 nil，具体实现需要根据 Kratos 应用类型
	return nil
}
