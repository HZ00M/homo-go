package driver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/patrickmn/go-cache"

	"github.com/go-kratos/kratos/v2/route"
)

// StatefulRouteForClientDriverImpl 有状态客户端驱动实现
type StatefulRouteForClientDriverImpl struct {
	mu sync.RWMutex

	// 依赖组件
	statefulExecutor route.StatefulExecutor
	routeInfoDriver  route.RouteInfoDriver

	// 缓存管理
	userPodIndexCache *cache.Cache

	// 配置信息
	serverInfo *route.ServerInfo

	// 常量
	userLinkInfoCacheTime int
	cacheExpiredValue     int

	logger log.Logger
}

// NewStatefulRouteForClientDriverImpl 创建新的客户端驱动实例
func NewStatefulRouteForClientDriverImpl(
	statefulExecutor route.StatefulExecutor,
	routeInfoDriver route.RouteInfoDriver,
	serverInfo *route.ServerInfo,
	logger log.Logger,
) *StatefulRouteForClientDriverImpl {
	// 缓存过期时间比Redis临时链接信息缓存时间短60秒
	cacheTime := 300 - 60 // 假设TEMP_LINK_INFO_CACHE_TIME_SECS = 300

	return &StatefulRouteForClientDriverImpl{
		statefulExecutor:      statefulExecutor,
		routeInfoDriver:       routeInfoDriver,
		serverInfo:            serverInfo,
		userLinkInfoCacheTime: cacheTime,
		cacheExpiredValue:     -100, // 缓存过期标识值
		logger:                logger,
	}
}

// Init 初始化客户端驱动
func (c *StatefulRouteForClientDriverImpl) Init() error {
	// 初始化缓存，设置过期时间
	c.userPodIndexCache = cache.New(
		time.Duration(c.userLinkInfoCacheTime)*time.Second,
		5*time.Minute, // 清理间隔
	)

	c.logger.Log(log.LevelInfo, "Stateful route client driver initialized")
	return nil
}

// ComputeLinkedPod 计算一个最佳podId
func (c *StatefulRouteForClientDriverImpl) ComputeLinkedPod(
	ctx context.Context,
	namespace, uid, serviceName string,
) (int, error) {
	// 首先尝试从缓存获取
	if userPodIndex, err := c.GetLinkedInCache(ctx, namespace, uid, serviceName); err == nil {
		// 检查pod是否仍然可路由
		if c.routeInfoDriver.IsPodRoutable(namespace, serviceName, userPodIndex) {
			c.logger.Log(log.LevelDebug, "Cache hit", "userPodIndex", userPodIndex, "uid", uid)
			return userPodIndex, nil
		} else {
			c.logger.Log(log.LevelError, "Pod is not routable", "podIndex", userPodIndex, "uid", uid, "service", serviceName)
			c.CacheExpired(ctx, namespace, uid, serviceName)
		}
	}

	// 获取最佳可用pod
	bestPodIndex, err := c.routeInfoDriver.GetServiceBestPod(ctx, namespace, serviceName)
	if err != nil {
		return 0, fmt.Errorf("failed to get best pod: %w", err)
	}

	if bestPodIndex != 0 {
		// 有可用pod，尝试设置链接（只支持本地命名空间）
		linkedPod, err := c.SetLinkedPodIfAbsent(ctx, namespace, uid, serviceName, bestPodIndex)

		if err != nil {
			return 0, fmt.Errorf("failed to set linked pod: %w", err)
		}

		// 验证返回的pod是否可用
		if linkedPod != 0 && c.routeInfoDriver.IsPodAvailable(namespace, serviceName, linkedPod) {
			c.logger.Log(log.LevelInfo, "Cache updated", "namespace", namespace, "uid", uid, "service", serviceName, "podIndex", linkedPod)
			c.SetCache(ctx, namespace, uid, serviceName, linkedPod)
			return linkedPod, nil
		}
	} else {
		// 没有可用pod，尝试从Redis获取现有链接
		c.logger.Log(log.LevelWarn, "No routable pods available, trying to get existing link", "uid", uid, "service", serviceName)
		linkedPod, err := c.GetLinkedPod(ctx, namespace, uid, serviceName)
		if err != nil {
			return 0, fmt.Errorf("failed to get linked pod: %w", err)
		}

		if linkedPod != 0 && c.routeInfoDriver.IsPodAvailable(namespace, serviceName, linkedPod) {
			c.logger.Log(log.LevelInfo, "Found existing link", "namespace", namespace, "uid", uid, "service", serviceName, "podIndex", linkedPod)
			c.SetCache(ctx, namespace, uid, serviceName, linkedPod)
			return linkedPod, nil
		}
	}

	return 0, fmt.Errorf("no available pod found")
}

// GetLinkedPod 获取uid下的连接信息
func (c *StatefulRouteForClientDriverImpl) GetLinkedPod(
	ctx context.Context,
	namespace, uid, serviceName string,
) (int, error) {
	// 首先尝试从缓存获取
	if userPodIndex, err := c.GetLinkedInCache(ctx, namespace, uid, serviceName); err == nil {
		return userPodIndex, nil
	}

	// 从Redis或远程获取
	return c.GetLinkedPodNotCache(ctx, namespace, uid, serviceName)
}

// GetLinkedPodInCacheOrIfPersist 从缓存或注册中心获取连接信息
func (c *StatefulRouteForClientDriverImpl) GetLinkedPodInCacheOrIfPersist(
	ctx context.Context,
	namespace, uid, serviceName string,
) (int, error) {
	// 首先尝试从缓存获取
	if userPodIndex, err := c.GetLinkedInCache(ctx, namespace, uid, serviceName); err == nil {
		return userPodIndex, nil
	}

	// 只支持本地命名空间
	if !c.isLocalNamespace(namespace) {
		return 0, fmt.Errorf("cross-namespace operation not supported")
	}

	// 从Redis获取持久化信息
	linkedPod, err := c.statefulExecutor.GetLinkedPodIfPersist(ctx, namespace, uid, serviceName)
	if err != nil {
		return 0, err
	}

	if linkedPod != 0 {
		c.logger.Log(log.LevelInfo, "Got linked pod from persist", "namespace", namespace, "uid", uid, "service", serviceName, "podIndex", linkedPod)
		c.SetCache(ctx, namespace, uid, serviceName, linkedPod)
	}

	return linkedPod, nil
}

// GetLinkedPodNotCache 直接从注册中心获取连接信息
func (c *StatefulRouteForClientDriverImpl) GetLinkedPodNotCache(
	ctx context.Context,
	namespace, uid, serviceName string,
) (int, error) {
	// 只支持本地命名空间
	if !c.isLocalNamespace(namespace) {
		return 0, fmt.Errorf("cross-namespace operation not supported")
	}

	return c.statefulExecutor.GetLinkedPod(ctx, namespace, uid, serviceName)
}

// GetLinkedInCache 获取缓存中的连接信息
func (c *StatefulRouteForClientDriverImpl) GetLinkedInCache(
	ctx context.Context,
	namespace, uid, serviceName string,
) (int, error) {
	key := c.formatCacheKey(namespace, uid, serviceName)

	if podIndex, found := c.userPodIndexCache.Get(key); found {
		if c.isCacheValid(podIndex.(int)) {
			c.logger.Log(log.LevelDebug, "Cache hit", "userPodIndex", podIndex, "uid", uid)
			return podIndex.(int), nil
		}
	}

	return 0, fmt.Errorf("cache miss or expired")
}

// CacheExpired 标识连接信息过期
func (c *StatefulRouteForClientDriverImpl) CacheExpired(
	ctx context.Context,
	namespace, uid, serviceName string,
) (int, error) {
	key := c.formatCacheKey(namespace, uid, serviceName)

	if previousPodIndex, found := c.userPodIndexCache.Get(key); found {
		if c.isCacheValid(previousPodIndex.(int)) {
			c.userPodIndexCache.Set(key, c.cacheExpiredValue, cache.DefaultExpiration)
			return previousPodIndex.(int), nil
		}
	}

	return 0, fmt.Errorf("no valid cache found")
}

// SetCache 设置连接信息到缓存
func (c *StatefulRouteForClientDriverImpl) SetCache(
	ctx context.Context,
	namespace, uid, serviceName string,
	podIndex int,
) (int, error) {
	key := c.formatCacheKey(namespace, uid, serviceName)

	var previousPodIndex int
	if previous, found := c.userPodIndexCache.Get(key); found {
		if c.isCacheValid(previous.(int)) {
			previousPodIndex = previous.(int)
		}
	}

	c.userPodIndexCache.Set(key, podIndex, cache.DefaultExpiration)

	if previousPodIndex != 0 {
		return previousPodIndex, nil
	}
	return 0, nil
}

// SetLinkedPodIfAbsent 设置链接pod（如果不存在）
func (c *StatefulRouteForClientDriverImpl) SetLinkedPodIfAbsent(
	ctx context.Context,
	namespace, uid, serviceName string,
	podIndex int,
) (int, error) {
	// 只支持本地命名空间
	if !c.isLocalNamespace(namespace) {
		return 0, fmt.Errorf("cross-namespace operation not supported")
	}

	return c.statefulExecutor.SetLinkedPodIfAbsent(
		ctx, namespace, uid, serviceName, podIndex, 300, // TEMP_LINK_INFO_CACHE_TIME_SECS
	)
}

// GetLinkService 获取用户连接的所有服务
func (c *StatefulRouteForClientDriverImpl) GetLinkService(
	ctx context.Context,
	namespace, uid string,
) (map[string]int, error) {
	// 只支持本地命名空间
	if !c.isLocalNamespace(namespace) {
		return nil, fmt.Errorf("cross-namespace operation not supported")
	}

	return c.statefulExecutor.GetLinkService(ctx, namespace, uid)
}

// BatchGetLinkedPod 批量获取连接信息
func (c *StatefulRouteForClientDriverImpl) BatchGetLinkedPod(
	ctx context.Context,
	namespace string,
	keys []string,
	serviceName string,
) (map[int][]string, error) {
	// 只支持本地命名空间
	if !c.isLocalNamespace(namespace) {
		return nil, fmt.Errorf("cross-namespace operation not supported")
	}

	return c.statefulExecutor.BatchGetLinkedPod(ctx, namespace, keys, serviceName)
}

// Close 关闭客户端驱动
func (c *StatefulRouteForClientDriverImpl) Close() error {
	if c.userPodIndexCache != nil {
		c.userPodIndexCache.Flush()
	}

	c.logger.Log(log.LevelInfo, "Stateful route client driver closed")
	return nil
}

// isCacheValid 检查缓存值是否有效
func (c *StatefulRouteForClientDriverImpl) isCacheValid(podIndex int) bool {
	return podIndex != 0 && podIndex != c.cacheExpiredValue
}

// formatCacheKey 格式化缓存键
func (c *StatefulRouteForClientDriverImpl) formatCacheKey(namespace, uid, serviceName string) string {
	return fmt.Sprintf("%s-%s-%s", namespace, uid, serviceName)
}

// isLocalNamespace 检查是否为本地命名空间
func (c *StatefulRouteForClientDriverImpl) isLocalNamespace(namespace string) bool {
	return namespace == "" || namespace == c.serverInfo.Namespace
}
