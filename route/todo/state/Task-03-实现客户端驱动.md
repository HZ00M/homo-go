## 6A 任务卡：实现客户端驱动（StatefulRouteForClientDriver）

- 编号: Task-03
- 模块: stateful-route
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: 待定
- 实际完成时间: -

### A1 目标（Aim）
实现有状态路由模块的客户端驱动，负责计算最佳pod、管理用户链接缓存、处理跨命名空间路由等核心功能，确保客户端能够高效地获取和缓存路由信息。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Java版本的完整客户端驱动实现
  - 🔄 部分实现：接口已定义，需要实现具体逻辑
  - ❌ 未实现：Go版本的客户端驱动实现
- **差距**：
  - Java实现逻辑需要转换为Go实现
  - 缓存管理、异步操作需要适配Go生态
  - 跨命名空间通信需要适配gRPC调用
- **约束**：
  - 必须保持与Java版本功能逻辑一致
  - 使用现有的routeInfoDriver和StatefulRedisExecutor
  - 符合Go语言和Kratos框架设计规范
- **风险**：
  - 技术风险：Go缓存实现可能与Java不同
  - 业务风险：路由逻辑转换可能遗漏
  - 依赖风险：gRPC客户端依赖变更影响

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`StatefulRouteForClientDriver` - 客户端驱动接口
  - **核心方法**：
    - Pod计算：`ComputeLinkedPod`、`SetLinkedPodIfAbsent`
    - 链接获取：`GetLinkedPod`、`GetLinkedPodNotCache`
    - 缓存管理：`GetLinkedInCache`、`CacheExpired`、`SetCache`
    - 批量操作：`GetLinkService`、`BatchGetLinkedPod`
  - **输入输出参数及错误码**：
    - 使用`context.Context`进行超时和取消控制
    - 返回`error`接口类型，支持错误包装
    - 使用Go的多返回值特性

- **架构设计**：
  - 采用组合模式，集成现有组件
  - 使用Go的缓存库管理本地缓存
  - 支持跨命名空间的gRPC通信

- **核心功能模块**：
  - `StatefulRouteForClientDriverImpl`: 客户端驱动实现
  - 缓存管理：本地缓存、过期策略
  - 路由计算：最佳pod选择、负载均衡
  - 跨命名空间：gRPC客户端、远程调用

- **极小任务拆分**：
  - T03-01：实现基础结构和依赖注入
  - T03-02：实现缓存管理功能
  - T03-03：实现pod计算和链接管理
  - T03-04：实现跨命名空间通信
  - T03-05：实现批量操作和性能优化

### A4 行动（Act）
#### T03-01：实现基础结构和依赖注入
```go
// route/driver/client_driver.go
package driver

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "github.com/patrickmn/go-cache"
    
    "tpf-service-go/stateful-route/interfaces"
    "tpf-service-go/stateful-route/types"
)

// StatefulRouteForClientDriverImpl 有状态客户端驱动实现
type StatefulRouteForClientDriverImpl struct {
    mu sync.RWMutex
    
    // 依赖组件
    statefulRedisExecutor interfaces.StatefulRedisExecutor
    routeInfoDriver       interfaces.RouteInfoDriver
    routeGrpcClient       interfaces.RouteGrpcClient
    
    // 缓存管理
    userPodIndexCache *cache.Cache
    
    // 配置信息
    serverInfo *ServerInfo
    
    // 常量
    userLinkInfoCacheTime int
    cacheExpiredValue     int
    
    logger log.Logger
}

// NewStatefulRouteForClientDriverImpl 创建新的客户端驱动实例
func NewStatefulRouteForClientDriverImpl(
    statefulRedisExecutor interfaces.StatefulRedisExecutor,
    routeInfoDriver interfaces.RouteInfoDriver,
    routeGrpcClient interfaces.RouteGrpcClient,
    serverInfo *ServerInfo,
    logger log.Logger,
) *StatefulRouteForClientDriverImpl {
    // 缓存过期时间比Redis临时链接信息缓存时间短60秒
    cacheTime := 300 - 60 // 假设TEMP_LINK_INFO_CACHE_TIME_SECS = 300
    
    return &StatefulRouteForClientDriverImpl{
        statefulRedisExecutor:  statefulRedisExecutor,
        routeInfoDriver:        routeInfoDriver,
        routeGrpcClient:        routeGrpcClient,
        serverInfo:             serverInfo,
        userLinkInfoCacheTime:  cacheTime,
        cacheExpiredValue:      -100, // 缓存过期标识值
        logger:                 logger,
    }
}

// Init 初始化客户端驱动
func (c *StatefulRouteForClientDriverImpl) Init() error {
    // 初始化缓存，设置过期时间
    c.userPodIndexCache = cache.New(
        time.Duration(c.userLinkInfoCacheTime)*time.Second,
        5*time.Minute, // 清理间隔
    )
    
    c.logger.Info("Stateful route client driver initialized")
    return nil
}
```

#### T03-02：实现缓存管理功能
```go
// route/driver/client_driver.go
package driver

// GetLinkedInCache 获取缓存中的连接信息
func (c *StatefulRouteForClientDriverImpl) GetLinkedInCache(
    ctx context.Context, 
    namespace, uid, serviceName string,
) (int, error) {
    key := c.formatCacheKey(namespace, uid, serviceName)
    
    if podIndex, found := c.userPodIndexCache.Get(key); found {
        if c.isCacheValid(podIndex.(int)) {
            c.logger.Debugf("Cache hit: userPodIndex=%d, uid=%s", podIndex, uid)
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

// isCacheValid 检查缓存值是否有效
func (c *StatefulRouteForClientDriverImpl) isCacheValid(podIndex int) bool {
    return podIndex != 0 && podIndex != c.cacheExpiredValue
}

// formatCacheKey 格式化缓存键
func (c *StatefulRouteForClientDriverImpl) formatCacheKey(namespace, uid, serviceName string) string {
    return fmt.Sprintf("%s-%s-%s", namespace, uid, serviceName)
}
```

#### T03-03：实现pod计算和链接管理
```go
// route/driver/client_driver.go
package driver

// ComputeLinkedPod 计算一个最佳podId
func (c *StatefulRouteForClientDriverImpl) ComputeLinkedPod(
    ctx context.Context, 
    namespace, uid, serviceName string,
) (int, error) {
    tag := c.formatCacheKey(namespace, uid, serviceName)
    
    // 首先尝试从缓存获取
    if userPodIndex, err := c.GetLinkedInCache(ctx, namespace, uid, serviceName); err == nil {
        // 检查pod是否仍然可路由
        if c.routeInfoDriver.IsPodRoutable(namespace, serviceName, userPodIndex) {
            c.logger.Debugf("Cache hit: userPodIndex=%d, uid=%s", userPodIndex, uid)
            return userPodIndex, nil
        } else {
            c.logger.Errorf("Pod %d is not routable for uid %s, service %s", userPodIndex, uid, serviceName)
            c.CacheExpired(ctx, namespace, uid, serviceName)
        }
    }
    
    // 获取最佳可用pod
    bestPodIndex, err := c.routeInfoDriver.GetServiceBestPod(ctx, namespace, serviceName)
    if err != nil {
        return 0, fmt.Errorf("failed to get best pod: %w", err)
    }
    
    if bestPodIndex != 0 {
        // 有可用pod，尝试设置链接
        var linkedPod int
        if c.isLocalNamespace(namespace) {
            linkedPod, err = c.SetLinkedPodIfAbsent(ctx, namespace, uid, serviceName, bestPodIndex)
        } else {
            linkedPod, err = c.routeGrpcClient.SetLinkedPodIfAbsent(ctx, namespace, uid, serviceName, bestPodIndex)
        }
        
        if err != nil {
            return 0, fmt.Errorf("failed to set linked pod: %w", err)
        }
        
        // 验证返回的pod是否可用
        if linkedPod != 0 && c.routeInfoDriver.IsPodAvailable(ctx, namespace, serviceName, linkedPod) {
            c.logger.Infof("Cache updated: namespace=%s, uid=%s, service=%s, podIndex=%d", 
                namespace, uid, serviceName, linkedPod)
            c.SetCache(ctx, namespace, uid, serviceName, linkedPod)
            return linkedPod, nil
        }
    } else {
        // 没有可用pod，尝试从Redis获取现有链接
        c.logger.Warnf("No routable pods available, trying to get existing link for uid=%s, service=%s", uid, serviceName)
        linkedPod, err := c.GetLinkedPod(ctx, namespace, uid, serviceName)
        if err != nil {
            return 0, fmt.Errorf("failed to get linked pod: %w", err)
        }
        
        if linkedPod != 0 && c.routeInfoDriver.IsPodAvailable(ctx, namespace, serviceName, linkedPod) {
            c.logger.Infof("Found existing link: namespace=%s, uid=%s, service=%s, podIndex=%d", 
                namespace, uid, serviceName, linkedPod)
            c.SetCache(ctx, namespace, uid, serviceName, linkedPod)
            return linkedPod, nil
        }
    }
    
    return 0, fmt.Errorf("no available pod found")
}

// SetLinkedPodIfAbsent 设置链接pod（如果不存在）
func (c *StatefulRouteForClientDriverImpl) SetLinkedPodIfAbsent(
    ctx context.Context, 
    namespace, uid, serviceName string, 
    podIndex int,
) (int, error) {
    if c.isLocalNamespace(namespace) {
        return c.statefulRedisExecutor.SetLinkedPodIfAbsent(
            ctx, namespace, uid, serviceName, podIndex, 300, // TEMP_LINK_INFO_CACHE_TIME_SECS
        )
    }
    
    return 0, fmt.Errorf("cross-namespace operation not supported for SetLinkedPodIfAbsent")
}
```

#### T03-04：实现跨命名空间通信
```go
// route/driver/client_driver.go
package driver

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

// GetLinkedPodNotCache 直接从注册中心获取连接信息
func (c *StatefulRouteForClientDriverImpl) GetLinkedPodNotCache(
    ctx context.Context, 
    namespace, uid, serviceName string,
) (int, error) {
    if c.isLocalNamespace(namespace) {
        return c.statefulRedisExecutor.GetLinkedPod(ctx, namespace, uid, serviceName)
    }
    
    // 跨命名空间，调用gRPC
    return c.routeGrpcClient.GetLinkedPod(ctx, namespace, uid, serviceName)
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
    
    if c.isLocalNamespace(namespace) {
        // 本命名空间，从Redis获取持久化信息
        linkedPod, err := c.statefulRedisExecutor.GetLinkedPodIfPersist(ctx, namespace, uid, serviceName)
        if err != nil {
            return 0, err
        }
        
        if linkedPod != 0 {
            c.logger.Infof("Got linked pod from persist: namespace=%s, uid=%s, service=%s, podIndex=%d", 
                namespace, uid, serviceName, linkedPod)
            c.SetCache(ctx, namespace, uid, serviceName, linkedPod)
        }
        
        return linkedPod, nil
    }
    
    // 跨命名空间，调用gRPC
    return c.routeGrpcClient.GetLinkedPodInCacheOrIfPersist(ctx, namespace, uid, serviceName)
}

// isLocalNamespace 检查是否为本地命名空间
func (c *StatefulRouteForClientDriverImpl) isLocalNamespace(namespace string) bool {
    return namespace == "" || namespace == c.serverInfo.Namespace
}
```

#### T03-05：实现批量操作和性能优化
```go
// route/driver/client_driver.go
package driver

// GetLinkService 获取用户连接的所有服务
func (c *StatefulRouteForClientDriverImpl) GetLinkService(
    ctx context.Context, 
    namespace, uid string,
) (map[string]int, error) {
    if c.isLocalNamespace(namespace) {
        return c.statefulRedisExecutor.GetLinkService(ctx, namespace, uid)
    }
    
    // 跨命名空间，调用gRPC
    return c.routeGrpcClient.GetLinkService(ctx, namespace, uid)
}

// BatchGetLinkedPod 批量获取连接信息
func (c *StatefulRouteForClientDriverImpl) BatchGetLinkedPod(
    ctx context.Context, 
    namespace string, 
    keys []string, 
    serviceName string,
) (map[int][]string, error) {
    if c.isLocalNamespace(namespace) {
        return c.statefulRedisExecutor.BatchGetLinkedPod(ctx, namespace, keys, serviceName)
    }
    
    // 跨命名空间，调用gRPC
    return c.routeGrpcClient.BatchGetLinkedPod(ctx, namespace, keys, serviceName)
}

// Close 关闭客户端驱动
func (c *StatefulRouteForClientDriverImpl) Close() error {
    if c.userPodIndexCache != nil {
        c.userPodIndexCache.Flush()
    }
    
    c.logger.Info("Stateful route client driver closed")
    return nil
}
```

### A5 验证（Assure）
- **测试用例**：
  - 缓存管理功能测试
  - Pod计算功能测试
  - 跨命名空间通信测试
  - 批量操作功能测试
- **性能验证**：
  - 缓存命中率测试
  - 批量操作性能测试
  - 并发访问性能测试
- **回归测试**：
  - 确保与Java版本功能一致
- **测试结果**：
  - 待实现

### A6 迭代（Advance）
- 性能优化：缓存策略优化、批量操作优化
- 功能扩展：支持更多路由策略、监控指标
- 观测性增强：添加详细日志、性能监控、缓存统计
- 下一步任务链接：Task-04 实现路由管理器

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
[总结任务完成情况]
