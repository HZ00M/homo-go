package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/route"
)

// ==================== 缓存配置 ====================

// CacheConfig 缓存配置
type CacheConfig struct {
	// 缓存更新间隔（秒）
	UpdateInterval time.Duration `json:"updateInterval"`
	// 缓存过期时间（秒）
	ExpirationTime time.Duration `json:"expirationTime"`
	// 最大缓存条目数
	MaxCacheEntries int `json:"maxCacheEntries"`
	// 启用缓存预热
	EnableWarmup bool `json:"enableWarmup"`
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		UpdateInterval:  30 * time.Second,
		ExpirationTime:  300 * time.Second,
		MaxCacheEntries: 10000,
		EnableWarmup:    true,
	}
}

// ==================== 缓存条目 ====================

// CacheEntry 缓存条目
type CacheEntry struct {
	Data        interface{} `json:"data"`
	ExpireAt    time.Time   `json:"expireAt"`
	LastAccess  time.Time   `json:"lastAccess"`
	AccessCount int64       `json:"accessCount"`
}

// IsExpired 检查缓存条目是否过期
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpireAt)
}

// Touch 更新访问时间和计数
func (e *CacheEntry) Touch() {
	e.LastAccess = time.Now()
	e.AccessCount++
}

// ==================== 服务状态缓存实现 ====================

// ServiceStateCacheImpl 服务状态缓存实现
type ServiceStateCacheImpl struct {
	mu sync.RWMutex

	// 配置
	config *route.StatefulBaseConfig

	// 日志记录器
	logger log.Logger

	// 路由信息驱动
	routeInfoDriver route.RouteInfoDriver

	// 缓存存储
	routableStateCache map[string]map[string][]*route.StatefulServiceState       // namespace -> serviceName -> states
	allServices        map[string]map[string]map[int]*route.StatefulServiceState // namespace -> serviceName -> podId -> state

	// 缓存统计
	stats *CacheStats

	// 控制通道
	stopCh chan struct{}
	wg     sync.WaitGroup

	// 初始化状态
	initialized bool
}

// CacheStats 缓存统计信息
type CacheStats struct {
	TotalEntries      int64         `json:"totalEntries"`
	HitCount          int64         `json:"hitCount"`
	MissCount         int64         `json:"missCount"`
	UpdateCount       int64         `json:"updateCount"`
	LastUpdateTime    time.Time     `json:"lastUpdateTime"`
	AverageUpdateTime time.Duration `json:"averageUpdateTime"`
}

// NewServiceStateCache 创建新的服务状态缓存
func NewServiceStateCache(config *route.StatefulBaseConfig, logger log.Logger, routeInfoDriver route.RouteInfoDriver) *ServiceStateCacheImpl {
	if config == nil {
		config = &route.StatefulBaseConfig{
			StateCacheExpireSecs: 300,
		}
	}

	return &ServiceStateCacheImpl{
		config:             config,
		logger:             logger,
		routeInfoDriver:    routeInfoDriver,
		routableStateCache: make(map[string]map[string][]*route.StatefulServiceState),
		allServices:        make(map[string]map[string]map[int]*route.StatefulServiceState),
		stats:              &CacheStats{},
		stopCh:             make(chan struct{}),
	}
}

// ==================== 接口实现 ====================

// GetOrder 获取初始化顺序
func (c *ServiceStateCacheImpl) GetOrder() int {
	return 1 // 缓存组件优先初始化
}

// OnInitModule 初始化模块
func (c *ServiceStateCacheImpl) OnInitModule() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	c.logger.Log(log.LevelInfo, "初始化服务状态缓存组件")

	// 启动缓存更新协程
	c.startUpdateRoutine()

	// 如果启用预热，执行缓存预热
	if c.config.StateCacheExpireSecs > 0 {
		go c.warmupCache()
	}

	c.initialized = true
	c.logger.Log(log.LevelInfo, "服务状态缓存组件初始化完成")
	return nil
}

// GetServiceBestPod 获取服务最佳Pod
func (c *ServiceStateCacheImpl) GetServiceBestPod(ctx context.Context, namespace, serviceName string) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 从缓存获取可路由的Pod
	routablePods := c.getRoutablePodsFromCache(namespace, serviceName)
	if len(routablePods) == 0 {
		return -1, fmt.Errorf("没有可路由的Pod: %s/%s", namespace, serviceName)
	}

	// 选择最佳Pod（这里使用简单的负载状态选择策略）
	bestPod := c.selectBestPod(routablePods)
	if bestPod == nil {
		return -1, fmt.Errorf("无法选择最佳Pod: %s/%s", namespace, serviceName)
	}

	c.stats.HitCount++
	return bestPod.PodID, nil
}

// IsPodAvailable 检查Pod是否可用
func (c *ServiceStateCacheImpl) IsPodAvailable(namespace, serviceName string, podIndex int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	serviceStates, exists := c.allServices[namespace]
	if !exists {
		return false
	}

	podStates, exists := serviceStates[serviceName]
	if !exists {
		return false
	}

	podState, exists := podStates[podIndex]
	if !exists {
		return false
	}

	return podState.State == route.ServiceStateReady
}

// IsPodRoutable 检查Pod是否可路由
func (c *ServiceStateCacheImpl) IsPodRoutable(namespace, serviceName string, podIndex int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	routablePods := c.getRoutablePodsFromCache(namespace, serviceName)
	for _, pod := range routablePods {
		if pod.PodID == podIndex {
			return true
		}
	}
	return false
}

// AlivePods 获取存活的Pod列表
func (c *ServiceStateCacheImpl) AlivePods(namespace, serviceName string) map[int]*route.StatefulServiceState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	serviceStates, exists := c.allServices[namespace]
	if !exists {
		return make(map[int]*route.StatefulServiceState)
	}

	podStates, exists := serviceStates[serviceName]
	if !exists {
		return make(map[int]*route.StatefulServiceState)
	}

	// 过滤存活的Pod（状态不是Terminating和Failed）
	alivePods := make(map[int]*route.StatefulServiceState)
	for podId, state := range podStates {
		if state.State != route.ServiceStateTerminating && state.State != route.ServiceStateFailed {
			alivePods[podId] = state
		}
	}

	return alivePods
}

// RoutablePods 获取可路由的Pod列表
func (c *ServiceStateCacheImpl) RoutablePods(namespace, serviceName string) map[int]*route.StatefulServiceState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	routablePods := c.getRoutablePodsFromCache(namespace, serviceName)
	result := make(map[int]*route.StatefulServiceState)

	for _, pod := range routablePods {
		result[pod.PodID] = pod
	}

	return result
}

// Run 运行缓存组件
func (c *ServiceStateCacheImpl) Run() {
	c.logger.Log(log.LevelInfo, "启动服务状态缓存组件")

	// 等待停止信号
	<-c.stopCh
	c.logger.Log(log.LevelInfo, "服务状态缓存组件已停止")
}

// Stop 停止缓存组件
func (c *ServiceStateCacheImpl) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return
	}

	c.logger.Log(log.LevelInfo, "停止服务状态缓存组件")
	close(c.stopCh)
	c.wg.Wait()
	c.initialized = false
}

// ==================== 私有方法 ====================

// startUpdateRoutine 启动缓存更新协程
func (c *ServiceStateCacheImpl) startUpdateRoutine() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		updateInterval := time.Duration(c.config.StateCacheExpireSecs) * time.Second
		if updateInterval == 0 {
			updateInterval = 30 * time.Second // 默认30秒
		}

		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.updateCache()
			case <-c.stopCh:
				return
			}
		}
	}()
}

// updateCache 更新缓存
func (c *ServiceStateCacheImpl) updateCache() {
	startTime := time.Now()

	c.logger.Log(log.LevelDebug, "开始更新服务状态缓存")

	// 这里应该从实际的数据源更新缓存
	// 目前使用模拟数据，实际实现中应该调用StatefulRedisExecutor
	c.updateCacheFromSource()

	updateTime := time.Since(startTime)
	c.stats.UpdateCount++
	c.stats.LastUpdateTime = time.Now()
	c.stats.AverageUpdateTime = (c.stats.AverageUpdateTime*time.Duration(c.stats.UpdateCount-1) + updateTime) / time.Duration(c.stats.UpdateCount)

	c.logger.Log(log.LevelDebug, "服务状态缓存更新完成", "耗时", updateTime)
}

// updateCacheFromSource 从数据源更新缓存
func (c *ServiceStateCacheImpl) updateCacheFromSource() {
	// 这里应该实现从StatefulRedisExecutor获取数据的逻辑
	// 目前使用模拟数据
	c.mu.Lock()
	defer c.mu.Unlock()

	// 模拟数据更新
	c.updateMockData()
}

// updateMockData 更新模拟数据
func (c *ServiceStateCacheImpl) updateMockData() {
	// 模拟命名空间
	namespaces := []string{"default", "kube-system"}

	for _, namespace := range namespaces {
		if c.allServices[namespace] == nil {
			c.allServices[namespace] = make(map[string]map[int]*route.StatefulServiceState)
		}

		// 模拟服务
		services := []string{"web-service", "api-service", "db-service"}
		for _, serviceName := range services {
			if c.allServices[namespace][serviceName] == nil {
				c.allServices[namespace][serviceName] = make(map[int]*route.StatefulServiceState)
			}

			// 模拟Pod状态
			for podId := 0; podId < 3; podId++ {
				state := &route.StatefulServiceState{
					PodID:        podId,
					State:        route.ServiceStateReady,
					LoadState:    2, // 低负载状态
					RoutingState: route.RoutingStateReady,
					UpdateTime:   time.Now().Unix(),
				}

				c.allServices[namespace][serviceName][podId] = state
			}
		}
	}

	// 更新可路由状态缓存
	c.updateRoutableStateCache()
}

// updateRoutableStateCache 更新可路由状态缓存
func (c *ServiceStateCacheImpl) updateRoutableStateCache() {
	c.routableStateCache = make(map[string]map[string][]*route.StatefulServiceState)

	for namespace, services := range c.allServices {
		if c.routableStateCache[namespace] == nil {
			c.routableStateCache[namespace] = make(map[string][]*route.StatefulServiceState)
		}

		for serviceName, podStates := range services {
			var routablePods []*route.StatefulServiceState

			for _, podState := range podStates {
				if c.isPodRoutable(podState) {
					routablePods = append(routablePods, podState)
				}
			}

			c.routableStateCache[namespace][serviceName] = routablePods
		}
	}
}

// isPodRoutable 检查Pod是否可路由
func (c *ServiceStateCacheImpl) isPodRoutable(podState *route.StatefulServiceState) bool {
	return podState.State == route.ServiceStateReady &&
		podState.RoutingState == route.RoutingStateReady &&
		podState.LoadState != 5 // 过载状态
}

// getRoutablePodsFromCache 从缓存获取可路由的Pod
func (c *ServiceStateCacheImpl) getRoutablePodsFromCache(namespace, serviceName string) []*route.StatefulServiceState {
	namespaceCache, exists := c.routableStateCache[namespace]
	if !exists {
		return nil
	}

	serviceCache, exists := namespaceCache[serviceName]
	if !exists {
		return nil
	}

	return serviceCache
}

// selectBestPod 选择最佳Pod
func (c *ServiceStateCacheImpl) selectBestPod(pods []*route.StatefulServiceState) *route.StatefulServiceState {
	if len(pods) == 0 {
		return nil
	}

	// 简单的选择策略：优先选择负载最低的Pod
	bestPod := pods[0]
	bestLoad := bestPod.LoadState

	for _, pod := range pods[1:] {
		if pod.LoadState < bestLoad {
			bestPod = pod
			bestLoad = pod.LoadState
		}
	}

	return bestPod
}

// warmupCache 缓存预热
func (c *ServiceStateCacheImpl) warmupCache() {
	c.logger.Log(log.LevelInfo, "开始缓存预热")

	// 执行一次完整的缓存更新
	c.updateCache()

	c.logger.Log(log.LevelInfo, "缓存预热完成")
}

// GetStats 获取缓存统计信息
func (c *ServiceStateCacheImpl) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := *c.stats
	stats.TotalEntries = int64(len(c.allServices))

	return &stats
}

// ClearCache 清空缓存
func (c *ServiceStateCacheImpl) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.routableStateCache = make(map[string]map[string][]*route.StatefulServiceState)
	c.allServices = make(map[string]map[string]map[int]*route.StatefulServiceState)

	c.logger.Log(log.LevelInfo, "缓存已清空")
}
