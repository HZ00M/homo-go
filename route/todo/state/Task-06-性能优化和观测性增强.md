## 6A 任务卡：性能优化和观测性增强

- 编号: Task-06
- 模块: stateful-route
- 责任人: 待分配
- 优先级: 🟢 低
- 状态: ❌ 未开始
- 预计完成时间: 待定
- 实际完成时间: -

### A1 目标（Aim）
对有状态路由模块进行性能优化和观测性增强，提升系统性能、可观测性和可维护性，包括缓存优化、算法优化、监控指标、链路追踪等。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：基础功能实现和单元测试
  - 🔄 部分实现：基本的性能特性
  - ❌ 未实现：高级性能优化和观测性功能
- **差距**：
  - 需要优化负载均衡算法性能
  - 需要增强缓存策略和命中率
  - 需要添加详细的监控指标和链路追踪
- **约束**：
  - 必须保持与Java版本功能一致
  - 符合Go语言和Kratos框架最佳实践
  - 不影响现有功能的稳定性
- **风险**：
  - 技术风险：优化可能引入新的bug
  - 业务风险：性能提升效果不明显
  - 依赖风险：监控组件依赖变更

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：性能监控和链路追踪接口
  - **核心方法**：
    - 性能监控：`RecordMetrics`、`GetMetrics`、`ExportMetrics`
    - 链路追踪：`StartSpan`、`EndSpan`、`AddEvent`
    - 缓存优化：`CacheStats`、`CacheOptimize`、`CacheEvict`
  - **输入输出参数及错误码**：
    - 使用标准监控指标格式（Prometheus、OpenTelemetry）
    - 支持配置驱动的性能调优
    - 提供性能分析报告

- **架构设计**：
  - 采用观察者模式，支持性能事件通知
  - 使用装饰器模式，支持性能监控包装
  - 支持配置驱动的性能调优策略

- **核心功能模块**：
  - 性能监控器：指标收集、性能分析
  - 链路追踪器：请求链路、性能分析
  - 缓存优化器：策略优化、命中率提升
  - 性能分析器：瓶颈识别、优化建议

- **极小任务拆分**：
  - T06-01：实现性能监控器
  - T06-02：实现链路追踪器
  - T06-03：优化缓存策略
  - T06-04：优化负载均衡算法
  - T06-05：实现性能分析报告

### A4 行动（Act）
#### T06-01：实现性能监控器
```go
// route/state/metrics/performance_monitor.go
package metrics

import (
    "context"
    "sync"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
    // 请求相关指标
    RequestTotal     prometheus.Counter
    RequestDuration  prometheus.Histogram
    RequestErrors    prometheus.Counter
    
    // 缓存相关指标
    CacheHits        prometheus.Counter
    CacheMisses      prometheus.Counter
    CacheSize        prometheus.Gauge
    
    // 负载均衡相关指标
    LoadBalancerRequests prometheus.Counter
    PodSelectionDuration prometheus.Histogram
    
    // 健康检查相关指标
    HealthCheckTotal     prometheus.Counter
    HealthCheckDuration  prometheus.Histogram
    HealthCheckFailures  prometheus.Counter
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
    mu sync.RWMutex
    
    metrics *PerformanceMetrics
    logger  log.Logger
    
    // 自定义指标
    customMetrics map[string]prometheus.Collector
}

// NewPerformanceMonitor 创建新的性能监控器
func NewPerformanceMonitor(logger log.Logger) *PerformanceMonitor {
    metrics := &PerformanceMetrics{
        RequestTotal: promauto.NewCounter(prometheus.CounterOpts{
            Name: "stateful_route_requests_total",
            Help: "Total number of routing requests",
        }),
        RequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
            Name:    "stateful_route_request_duration_seconds",
            Help:    "Duration of routing requests",
            Buckets: prometheus.DefBuckets,
        }),
        RequestErrors: promauto.NewCounter(prometheus.CounterOpts{
            Name: "stateful_route_request_errors_total",
            Help: "Total number of routing request errors",
        }),
        CacheHits: promauto.NewCounter(prometheus.CounterOpts{
            Name: "stateful_route_cache_hits_total",
            Help: "Total number of cache hits",
        }),
        CacheMisses: promauto.NewCounter(prometheus.CounterOpts{
            Name: "stateful_route_cache_misses_total",
            Help: "Total number of cache misses",
        }),
        CacheSize: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "stateful_route_cache_size",
            Help: "Current cache size",
        }),
        LoadBalancerRequests: promauto.NewCounter(prometheus.CounterOpts{
            Name: "stateful_route_load_balancer_requests_total",
            Help: "Total number of load balancer requests",
        }),
        PodSelectionDuration: promauto.NewHistogram(prometheus.HistogramOpts{
            Name:    "stateful_route_pod_selection_duration_seconds",
            Help:    "Duration of pod selection",
            Buckets: prometheus.DefBuckets,
        }),
        HealthCheckTotal: promauto.NewCounter(prometheus.CounterOpts{
            Name: "stateful_route_health_check_total",
            Help: "Total number of health checks",
        }),
        HealthCheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
            Name:    "stateful_route_health_check_duration_seconds",
            Help:    "Duration of health checks",
            Buckets: prometheus.DefBuckets,
        }),
        HealthCheckFailures: promauto.NewCounter(prometheus.CounterOpts{
            Name: "stateful_route_health_check_failures_total",
            Help: "Total number of health check failures",
        }),
    }
    
    return &PerformanceMonitor{
        metrics:       metrics,
        logger:        logger,
        customMetrics: make(map[string]prometheus.Collector),
    }
}

// RecordRequest 记录请求指标
func (pm *PerformanceMonitor) RecordRequest(ctx context.Context, duration time.Duration, err error) {
    pm.metrics.RequestTotal.Inc()
    pm.metrics.RequestDuration.Observe(duration.Seconds())
    
    if err != nil {
        pm.metrics.RequestErrors.Inc()
    }
}

// RecordCacheHit 记录缓存命中
func (pm *PerformanceMonitor) RecordCacheHit(ctx context.Context) {
    pm.metrics.CacheHits.Inc()
}

// RecordCacheMiss 记录缓存未命中
func (pm *PerformanceMonitor) RecordCacheMiss(ctx context.Context) {
    pm.metrics.CacheMisses.Inc()
}

// UpdateCacheSize 更新缓存大小
func (pm *PerformanceMonitor) UpdateCacheSize(ctx context.Context, size int) {
    pm.metrics.CacheSize.Set(float64(size))
}

// RecordLoadBalancerRequest 记录负载均衡请求
func (pm *PerformanceMonitor) RecordLoadBalancerRequest(ctx context.Context, duration time.Duration) {
    pm.metrics.LoadBalancerRequests.Inc()
    pm.metrics.PodSelectionDuration.Observe(duration.Seconds())
}

// RecordHealthCheck 记录健康检查
func (pm *PerformanceMonitor) RecordHealthCheck(ctx context.Context, duration time.Duration, success bool) {
    pm.metrics.HealthCheckTotal.Inc()
    pm.metrics.HealthCheckDuration.Observe(duration.Seconds())
    
    if !success {
        pm.metrics.HealthCheckFailures.Inc()
    }
}

// GetCacheHitRate 获取缓存命中率
func (pm *PerformanceMonitor) GetCacheHitRate(ctx context.Context) float64 {
    hits := pm.metrics.CacheHits.Get()
    misses := pm.metrics.CacheMisses.Get()
    total := hits + misses
    
    if total == 0 {
        return 0.0
    }
    
    return hits / total
}

// ExportMetrics 导出指标
func (pm *PerformanceMonitor) ExportMetrics(ctx context.Context) map[string]interface{} {
    return map[string]interface{}{
        "cache_hit_rate": pm.GetCacheHitRate(ctx),
        "request_total":   pm.metrics.RequestTotal.Get(),
        "request_errors":  pm.metrics.RequestErrors.Get(),
        "cache_size":      pm.metrics.CacheSize.Get(),
    }
}
```

#### T06-02：实现链路追踪器
```go
// route/state/tracing/request_tracer.go
package tracing

import (
    "context"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

// RequestTracer 请求链路追踪器
type RequestTracer struct {
    tracer trace.Tracer
    logger log.Logger
}

// TraceInfo 追踪信息
type TraceInfo struct {
    TraceID    string
    SpanID     string
    ParentID   string
    Operation  string
    StartTime  time.Time
    EndTime    time.Time
    Attributes map[string]string
}

// NewRequestTracer 创建新的请求追踪器
func NewRequestTracer(logger log.Logger) *RequestTracer {
    return &RequestTracer{
        tracer: otel.Tracer("stateful-route"),
        logger: logger,
    }
}

// StartSpan 开始追踪span
func (rt *RequestTracer) StartSpan(ctx context.Context, operation string, attributes map[string]string) (context.Context, trace.Span) {
    // 转换属性
    attrs := make([]attribute.KeyValue, 0, len(attributes))
    for k, v := range attributes {
        attrs = append(attrs, attribute.String(k, v))
    }
    
    // 创建span
    ctx, span := rt.tracer.Start(ctx, operation, trace.WithAttributes(attrs...))
    
    // 记录开始时间
    span.SetAttributes(attribute.Int64("start_time", time.Now().UnixNano()))
    
    return ctx, span
}

// EndSpan 结束追踪span
func (rt *RequestTracer) EndSpan(span trace.Span, err error) {
    if err != nil {
        span.SetStatus(trace.StatusError, err.Error())
        span.RecordError(err)
    } else {
        span.SetStatus(trace.StatusOK, "")
    }
    
    // 记录结束时间
    span.SetAttributes(attribute.Int64("end_time", time.Now().UnixNano()))
    
    span.End()
}

// AddEvent 添加事件
func (rt *RequestTracer) AddEvent(span trace.Span, name string, attributes map[string]string) {
    attrs := make([]attribute.KeyValue, 0, len(attributes))
    for k, v := range attributes {
        attrs = append(attrs, attribute.String(k, v))
    }
    
    span.AddEvent(name, trace.WithAttributes(attrs...))
}

// TraceRequest 追踪请求
func (rt *RequestTracer) TraceRequest(ctx context.Context, operation string, fn func(context.Context) error) error {
    ctx, span := rt.StartSpan(ctx, operation, nil)
    defer rt.EndSpan(span, nil)
    
    startTime := time.Now()
    err := fn(ctx)
    duration := time.Since(startTime)
    
    // 记录请求持续时间
    span.SetAttributes(attribute.Duration("duration", duration))
    
    if err != nil {
        rt.EndSpan(span, err)
    }
    
    return err
}

// GetTraceInfo 获取追踪信息
func (rt *RequestTracer) GetTraceInfo(ctx context.Context) *TraceInfo {
    span := trace.SpanFromContext(ctx)
    if span == nil {
        return nil
    }
    
    spanCtx := span.SpanContext()
    
    return &TraceInfo{
        TraceID:    spanCtx.TraceID().String(),
        SpanID:     spanCtx.SpanID().String(),
        ParentID:   "", // 需要从上下文获取
        Operation:  span.Name(),
        StartTime:  time.Now(), // 需要从span获取
        EndTime:    time.Now(), // 需要从span获取
        Attributes: make(map[string]string),
    }
}
```

#### T06-03：优化缓存策略
```go
// route/state/cache/cache_optimizer.go
package cache

import (
    "context"
    "sync"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "github.com/patrickmn/go-cache"
)

// CacheStrategy 缓存策略
type CacheStrategy int

const (
    StrategyLRU CacheStrategy = iota
    StrategyLFU
    StrategyTTL
    StrategyAdaptive
)

// CacheOptimizer 缓存优化器
type CacheOptimizer struct {
    mu sync.RWMutex
    
    // 缓存统计
    stats *CacheStats
    
    // 缓存策略
    strategy CacheStrategy
    
    // 自适应参数
    adaptiveParams *AdaptiveParams
    
    logger log.Logger
}

// CacheStats 缓存统计
type CacheStats struct {
    Hits        int64
    Misses      int64
    Evictions   int64
    Expirations int64
    Size        int
    MaxSize     int
}

// AdaptiveParams 自适应参数
type AdaptiveParams struct {
    HitRateThreshold float64
    EvictionRate    float64
    TTLAdjustment   time.Duration
}

// NewCacheOptimizer 创建新的缓存优化器
func NewCacheOptimizer(strategy CacheStrategy, logger log.Logger) *CacheOptimizer {
    return &CacheOptimizer{
        stats: &CacheStats{},
        strategy: strategy,
        adaptiveParams: &AdaptiveParams{
            HitRateThreshold: 0.8,
            EvictionRate:     0.1,
            TTLAdjustment:    5 * time.Minute,
        },
        logger: logger,
    }
}

// OptimizeCache 优化缓存
func (co *CacheOptimizer) OptimizeCache(ctx context.Context, c *cache.Cache) error {
    co.mu.Lock()
    defer co.mu.Unlock()
    
    // 更新统计信息
    co.updateStats(c)
    
    // 根据策略进行优化
    switch co.strategy {
    case StrategyLRU:
        return co.optimizeLRU(ctx, c)
    case StrategyLFU:
        return co.optimizeLFU(ctx, c)
    case StrategyTTL:
        return co.optimizeTTL(ctx, c)
    case StrategyAdaptive:
        return co.optimizeAdaptive(ctx, c)
    default:
        return co.optimizeLRU(ctx, c)
    }
}

// optimizeLRU LRU优化策略
func (co *CacheOptimizer) optimizeLRU(ctx context.Context, c *cache.Cache) error {
    // 如果缓存大小超过阈值，进行LRU清理
    if co.stats.Size > co.stats.MaxSize*8/10 {
        co.logger.Infof("LRU optimization: cache size %d exceeds threshold %d", co.stats.Size, co.stats.MaxSize*8/10)
        
        // 清理最久未使用的项
        c.DeleteExpired()
        co.stats.Evictions++
    }
    
    return nil
}

// optimizeLFU LFU优化策略
func (co *CacheOptimizer) optimizeLFU(ctx context.Context, c *cache.Cache) error {
    // LFU策略需要记录访问频率
    // 这里简化实现，实际应该使用更复杂的LFU算法
    return co.optimizeLRU(ctx, c)
}

// optimizeTTL TTL优化策略
func (co *CacheOptimizer) optimizeTTL(ctx context.Context, c *cache.Cache) error {
    // 根据命中率调整TTL
    hitRate := co.getHitRate()
    
    if hitRate < co.adaptiveParams.HitRateThreshold {
        // 命中率低，增加TTL
        co.logger.Infof("TTL optimization: hit rate %.2f below threshold %.2f, increasing TTL", hitRate, co.adaptiveParams.HitRateThreshold)
        // 这里应该调整缓存的TTL设置
    }
    
    return nil
}

// optimizeAdaptive 自适应优化策略
func (co *CacheOptimizer) optimizeAdaptive(ctx context.Context, c *cache.Cache) error {
    hitRate := co.getHitRate()
    evictionRate := co.getEvictionRate()
    
    // 根据命中率和驱逐率进行自适应调整
    if hitRate < co.adaptiveParams.HitRateThreshold {
        // 命中率低，增加缓存大小或TTL
        co.logger.Infof("Adaptive optimization: hit rate %.2f below threshold, adjusting cache parameters", hitRate)
    }
    
    if evictionRate > co.adaptiveParams.EvictionRate {
        // 驱逐率高，减少缓存大小或TTL
        co.logger.Infof("Adaptive optimization: eviction rate %.2f above threshold, adjusting cache parameters", evictionRate)
    }
    
    return nil
}

// updateStats 更新统计信息
func (co *CacheOptimizer) updateStats(c *cache.Cache) {
    // 获取缓存统计信息
    stats := c.Stats()
    
    co.stats.Hits = stats.Hits
    co.stats.Misses = stats.Misses
    co.stats.Evictions = stats.Evictions
    co.stats.Expirations = stats.Expirations
    co.stats.Size = stats.Size
    co.stats.MaxSize = stats.MaxSize
}

// getHitRate 获取命中率
func (co *CacheOptimizer) getHitRate() float64 {
    total := co.stats.Hits + co.stats.Misses
    if total == 0 {
        return 0.0
    }
    return float64(co.stats.Hits) / float64(total)
}

// getEvictionRate 获取驱逐率
func (co *CacheOptimizer) getEvictionRate() float64 {
    total := co.stats.Hits + co.stats.Misses
    if total == 0 {
        return 0.0
    }
    return float64(co.stats.Evictions) / float64(total)
}

// GetStats 获取缓存统计
func (co *CacheOptimizer) GetStats(ctx context.Context) *CacheStats {
    co.mu.RLock()
    defer co.mu.RUnlock()
    
    return &CacheStats{
        Hits:        co.stats.Hits,
        Misses:      co.stats.Misses,
        Evictions:   co.stats.Evictions,
        Expirations: co.stats.Expirations,
        Size:        co.stats.Size,
        MaxSize:     co.stats.MaxSize,
    }
}
```

#### T06-04：优化负载均衡算法
```go
// route/state/loadbalancer/optimized_load_balancer.go
package loadbalancer

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
)

// OptimizedLoadBalancer 优化的负载均衡器
type OptimizedLoadBalancer struct {
    mu sync.RWMutex
    
    // 轮询计数器
    roundRobinCounter uint64
    
    // 加权轮询状态
    weightedState map[int]*WeightedState
    
    // 最少连接状态
    connectionState map[int]*ConnectionState
    
    // 性能统计
    stats *LoadBalancerStats
    
    logger log.Logger
}

// WeightedState 加权状态
type WeightedState struct {
    CurrentWeight int
    EffectiveWeight int
    PodIndex int
}

// ConnectionState 连接状态
type ConnectionState struct {
    ConnectionCount int64
    LastUpdate     time.Time
    PodIndex       int
}

// LoadBalancerStats 负载均衡器统计
type LoadBalancerStats struct {
    TotalRequests    int64
    StrategyRequests map[string]int64
    AverageLatency   time.Duration
    LastUpdate       time.Time
}

// NewOptimizedLoadBalancer 创建新的优化负载均衡器
func NewOptimizedLoadBalancer(logger log.Logger) *OptimizedLoadBalancer {
    return &OptimizedLoadBalancer{
        weightedState:    make(map[int]*WeightedState),
        connectionState:  make(map[int]*ConnectionState),
        stats: &LoadBalancerStats{
            StrategyRequests: make(map[string]int64),
        },
        logger: logger,
    }
}

// SelectPod 选择pod（优化版本）
func (olb *OptimizedLoadBalancer) SelectPod(ctx context.Context, pods []*PodInfo, strategy LoadBalancingStrategy) (int, error) {
    if len(pods) == 0 {
        return 0, fmt.Errorf("no available pods")
    }
    
    // 过滤健康的pod
    healthyPods := olb.filterHealthyPods(pods)
    if len(healthyPods) == 0 {
        return 0, fmt.Errorf("no healthy pods available")
    }
    
    startTime := time.Now()
    var podIndex int
    var err error
    
    // 根据策略选择pod
    switch strategy {
    case StrategyRoundRobin:
        podIndex, err = olb.optimizedRoundRobin(healthyPods)
    case StrategyLeastConnections:
        podIndex, err = olb.optimizedLeastConnections(healthyPods)
    case StrategyWeightedRoundRobin:
        podIndex, err = olb.optimizedWeightedRoundRobin(healthyPods)
    case StrategyLeastResponseTime:
        podIndex, err = olb.optimizedLeastResponseTime(healthyPods)
    default:
        podIndex, err = olb.optimizedRoundRobin(healthyPods)
    }
    
    // 更新统计信息
    olb.updateStats(strategy, time.Since(startTime))
    
    return podIndex, err
}

// optimizedRoundRobin 优化的轮询策略
func (olb *OptimizedLoadBalancer) optimizedRoundRobin(pods []*PodInfo) (int, error) {
    // 使用原子操作进行轮询
    counter := atomic.AddUint64(&olb.roundRobinCounter, 1)
    selectedIndex := int(counter % uint64(len(pods)))
    
    return pods[selectedIndex].Index, nil
}

// optimizedLeastConnections 优化的最少连接策略
func (olb *OptimizedLoadBalancer) optimizedLeastConnections(pods []*PodInfo) (int, error) {
    olb.mu.RLock()
    defer olb.mu.RUnlock()
    
    var minConnections int64 = 1<<63 - 1
    var selectedPod *PodInfo
    
    for _, pod := range pods {
        // 获取连接状态
        state, exists := olb.connectionState[pod.Index]
        if !exists {
            // 初始化连接状态
            olb.connectionState[pod.Index] = &ConnectionState{
                ConnectionCount: 0,
                LastUpdate:     time.Now(),
                PodIndex:       pod.Index,
            }
            state = olb.connectionState[pod.Index]
        }
        
        if state.ConnectionCount < minConnections {
            minConnections = state.ConnectionCount
            selectedPod = pod
        }
    }
    
    if selectedPod == nil {
        return 0, fmt.Errorf("no pod selected")
    }
    
    return selectedPod.Index, nil
}

// optimizedWeightedRoundRobin 优化的加权轮询策略
func (olb *OptimizedLoadBalancer) optimizedWeightedRoundRobin(pods []*PodInfo) (int, error) {
    olb.mu.Lock()
    defer olb.mu.Unlock()
    
    // 初始化加权状态
    for _, pod := range pods {
        if _, exists := olb.weightedState[pod.Index]; !exists {
            olb.weightedState[pod.Index] = &WeightedState{
                CurrentWeight:   pod.Weight,
                EffectiveWeight: pod.Weight,
                PodIndex:        pod.Index,
            }
        }
    }
    
    // 选择权重最高的pod
    var maxWeight int
    var selectedPod *PodInfo
    
    for _, pod := range pods {
        state := olb.weightedState[pod.Index]
        if state.CurrentWeight > maxWeight {
            maxWeight = state.CurrentWeight
            selectedPod = pod
        }
    }
    
    if selectedPod == nil {
        return 0, fmt.Errorf("no pod selected")
    }
    
    // 更新权重
    state := olb.weightedState[selectedPod.Index]
    state.CurrentWeight -= olb.calculateTotalWeight(pods)
    
    // 如果权重为负，重置为有效权重
    if state.CurrentWeight <= 0 {
        state.CurrentWeight += state.EffectiveWeight
    }
    
    return selectedPod.Index, nil
}

// optimizedLeastResponseTime 优化的最少响应时间策略
func (olb *OptimizedLoadBalancer) optimizedLeastResponseTime(pods []*PodInfo) (int, error) {
    if len(pods) == 0 {
        return 0, fmt.Errorf("no pods available")
    }
    
    minResponseTime := pods[0].ResponseTime
    selectedPod := pods[0]
    
    for _, pod := range pods[1:] {
        if pod.ResponseTime < minResponseTime {
            minResponseTime = pod.ResponseTime
            selectedPod = pod
        }
    }
    
    return selectedPod.Index, nil
}

// filterHealthyPods 过滤健康的pod
func (olb *OptimizedLoadBalancer) filterHealthyPods(pods []*PodInfo) []*PodInfo {
    healthyPods := make([]*PodInfo, 0, len(pods))
    for _, pod := range pods {
        if pod.IsHealthy {
            healthyPods = append(healthyPods, pod)
        }
    }
    return healthyPods
}

// calculateTotalWeight 计算总权重
func (olb *OptimizedLoadBalancer) calculateTotalWeight(pods []*PodInfo) int {
    total := 0
    for _, pod := range pods {
        total += pod.Weight
    }
    return total
}

// updateStats 更新统计信息
func (olb *OptimizedLoadBalancer) updateStats(strategy LoadBalancingStrategy, latency time.Duration) {
    olb.mu.Lock()
    defer olb.mu.Unlock()
    
    atomic.AddInt64(&olb.stats.TotalRequests, 1)
    
    strategyName := strategy.String()
    olb.stats.StrategyRequests[strategyName]++
    
    // 更新平均延迟
    if olb.stats.AverageLatency == 0 {
        olb.stats.AverageLatency = latency
    } else {
        olb.stats.AverageLatency = (olb.stats.AverageLatency + latency) / 2
    }
    
    olb.stats.LastUpdate = time.Now()
}

// GetStats 获取统计信息
func (olb *OptimizedLoadBalancer) GetStats(ctx context.Context) *LoadBalancerStats {
    olb.mu.RLock()
    defer olb.mu.RUnlock()
    
    // 返回副本
    stats := &LoadBalancerStats{
        TotalRequests:    atomic.LoadInt64(&olb.stats.TotalRequests),
        StrategyRequests: make(map[string]int64),
        AverageLatency:   olb.stats.AverageLatency,
        LastUpdate:       olb.stats.LastUpdate,
    }
    
    for k, v := range olb.stats.StrategyRequests {
        stats.StrategyRequests[k] = v
    }
    
    return stats
}
```

#### T06-05：实现性能分析报告
```go
// route/state/analysis/performance_analyzer.go
package analysis

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
)

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
    mu sync.RWMutex
    
    // 性能数据收集器
    collectors map[string]PerformanceCollector
    
    // 分析规则
    rules []AnalysisRule
    
    // 分析结果
    results []AnalysisResult
    
    logger log.Logger
}

// PerformanceCollector 性能数据收集器接口
type PerformanceCollector interface {
    Collect(ctx context.Context) (*PerformanceData, error)
    GetName() string
}

// PerformanceData 性能数据
type PerformanceData struct {
    Timestamp   time.Time
    Collector   string
    Metrics     map[string]float64
    RawData     interface{}
}

// AnalysisRule 分析规则
type AnalysisRule struct {
    Name        string
    Description string
    Threshold   float64
    Severity    Severity
    Condition   func(*PerformanceData) bool
}

// AnalysisResult 分析结果
type AnalysisResult struct {
    Timestamp   time.Time
    Rule        string
    Severity    Severity
    Message     string
    Data        *PerformanceData
    Suggestions []string
}

// Severity 严重程度
type Severity int

const (
    SeverityLow Severity = iota
    SeverityMedium
    SeverityHigh
    SeverityCritical
)

// NewPerformanceAnalyzer 创建新的性能分析器
func NewPerformanceAnalyzer(logger log.Logger) *PerformanceAnalyzer {
    analyzer := &PerformanceAnalyzer{
        collectors: make(map[string]PerformanceCollector),
        rules:      make([]AnalysisRule, 0),
        results:    make([]AnalysisResult, 0),
        logger:     logger,
    }
    
    // 添加默认分析规则
    analyzer.addDefaultRules()
    
    return analyzer
}

// AddCollector 添加性能数据收集器
func (pa *PerformanceAnalyzer) AddCollector(collector PerformanceCollector) error {
    pa.mu.Lock()
    defer pa.mu.Unlock()
    
    name := collector.GetName()
    if _, exists := pa.collectors[name]; exists {
        return fmt.Errorf("collector %s already exists", name)
    }
    
    pa.collectors[name] = collector
    pa.logger.Infof("Added performance collector: %s", name)
    
    return nil
}

// AddRule 添加分析规则
func (pa *PerformanceAnalyzer) AddRule(rule AnalysisRule) error {
    pa.mu.Lock()
    defer pa.mu.Unlock()
    
    pa.rules = append(pa.rules, rule)
    pa.logger.Infof("Added analysis rule: %s", rule.Name)
    
    return nil
}

// Analyze 执行性能分析
func (pa *PerformanceAnalyzer) Analyze(ctx context.Context) ([]AnalysisResult, error) {
    pa.mu.Lock()
    defer pa.mu.Unlock()
    
    var allResults []AnalysisResult
    
    // 收集所有性能数据
    for name, collector := range pa.collectors {
        data, err := collector.Collect(ctx)
        if err != nil {
            pa.logger.Errorf("Failed to collect data from %s: %v", name, err)
            continue
        }
        
        // 应用分析规则
        results := pa.applyRules(data)
        allResults = append(allResults, results...)
    }
    
    // 保存分析结果
    pa.results = append(pa.results, allResults...)
    
    // 限制结果数量
    if len(pa.results) > 1000 {
        pa.results = pa.results[len(pa.results)-1000:]
    }
    
    pa.logger.Infof("Performance analysis completed: %d results generated", len(allResults))
    
    return allResults, nil
}

// applyRules 应用分析规则
func (pa *PerformanceAnalyzer) applyRules(data *PerformanceData) []AnalysisResult {
    var results []AnalysisResult
    
    for _, rule := range pa.rules {
        if rule.Condition(data) {
            result := AnalysisResult{
                Timestamp:   time.Now(),
                Rule:        rule.Name,
                Severity:    rule.Severity,
                Message:     rule.Description,
                Data:        data,
                Suggestions: pa.generateSuggestions(rule, data),
            }
            
            results = append(results, result)
        }
    }
    
    return results
}

// generateSuggestions 生成优化建议
func (pa *PerformanceAnalyzer) generateSuggestions(rule AnalysisRule, data *PerformanceData) []string {
    var suggestions []string
    
    switch rule.Name {
    case "cache_hit_rate_low":
        suggestions = append(suggestions, "增加缓存大小", "优化缓存策略", "检查缓存键设计")
    case "response_time_high":
        suggestions = append(suggestions, "优化算法实现", "增加缓存", "检查依赖服务性能")
    case "error_rate_high":
        suggestions = append(suggestions, "检查错误日志", "优化错误处理", "增加重试机制")
    case "memory_usage_high":
        suggestions = append(suggestions, "检查内存泄漏", "优化数据结构", "增加垃圾回收")
    default:
        suggestions = append(suggestions, "检查系统配置", "优化代码实现", "增加监控指标")
    }
    
    return suggestions
}

// addDefaultRules 添加默认分析规则
func (pa *PerformanceAnalyzer) addDefaultRules() {
    // 缓存命中率规则
    pa.AddRule(AnalysisRule{
        Name:        "cache_hit_rate_low",
        Description: "缓存命中率低于阈值",
        Threshold:   0.8,
        Severity:    SeverityMedium,
        Condition: func(data *PerformanceData) bool {
            if hitRate, exists := data.Metrics["cache_hit_rate"]; exists {
                return hitRate < 0.8
            }
            return false
        },
    })
    
    // 响应时间规则
    pa.AddRule(AnalysisRule{
        Name:        "response_time_high",
        Description: "响应时间过高",
        Threshold:   100, // 毫秒
        Severity:    SeverityHigh,
        Condition: func(data *PerformanceData) bool {
            if responseTime, exists := data.Metrics["response_time"]; exists {
                return responseTime > 100
            }
            return false
        },
    })
    
    // 错误率规则
    pa.AddRule(AnalysisRule{
        Name:        "error_rate_high",
        Description: "错误率过高",
        Threshold:   0.05, // 5%
        Severity:    SeverityCritical,
        Condition: func(data *PerformanceData) bool {
            if errorRate, exists := data.Metrics["error_rate"]; exists {
                return errorRate > 0.05
            }
            return false
        },
    })
}

// GetResults 获取分析结果
func (pa *PerformanceAnalyzer) GetResults(ctx context.Context, severity Severity) []AnalysisResult {
    pa.mu.RLock()
    defer pa.mu.RUnlock()
    
    var filteredResults []AnalysisResult
    
    for _, result := range pa.results {
        if result.Severity >= severity {
            filteredResults = append(filteredResults, result)
        }
    }
    
    return filteredResults
}

// GenerateReport 生成性能分析报告
func (pa *PerformanceAnalyzer) GenerateReport(ctx context.Context) (*PerformanceReport, error) {
    results, err := pa.Analyze(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze performance: %w", err)
    }
    
    report := &PerformanceReport{
        Timestamp:     time.Now(),
        Summary:       pa.generateSummary(results),
        CriticalIssues: pa.filterResultsBySeverity(results, SeverityCritical),
        HighIssues:    pa.filterResultsBySeverity(results, SeverityHigh),
        MediumIssues:  pa.filterResultsBySeverity(results, SeverityMedium),
        LowIssues:     pa.filterResultsBySeverity(results, SeverityLow),
        Suggestions:   pa.generateOverallSuggestions(results),
    }
    
    return report, nil
}

// PerformanceReport 性能分析报告
type PerformanceReport struct {
    Timestamp      time.Time
    Summary        string
    CriticalIssues []AnalysisResult
    HighIssues     []AnalysisResult
    MediumIssues   []AnalysisResult
    LowIssues      []AnalysisResult
    Suggestions    []string
}

// generateSummary 生成摘要
func (pa *PerformanceAnalyzer) generateSummary(results []AnalysisResult) string {
    total := len(results)
    critical := len(pa.filterResultsBySeverity(results, SeverityCritical))
    high := len(pa.filterResultsBySeverity(results, SeverityHigh))
    
    if critical > 0 {
        return fmt.Sprintf("发现 %d 个严重问题，%d 个高级问题，建议立即处理", critical, high)
    } else if high > 0 {
        return fmt.Sprintf("发现 %d 个高级问题，建议优先处理", high)
    } else {
        return fmt.Sprintf("系统性能正常，共分析 %d 个指标", total)
    }
}

// filterResultsBySeverity 按严重程度过滤结果
func (pa *PerformanceAnalyzer) filterResultsBySeverity(results []AnalysisResult, severity Severity) []AnalysisResult {
    var filtered []AnalysisResult
    for _, result := range results {
        if result.Severity == severity {
            filtered = append(filtered, result)
        }
    }
    return filtered
}

// generateOverallSuggestions 生成整体优化建议
func (pa *PerformanceAnalyzer) generateOverallSuggestions(results []AnalysisResult) []string {
    var suggestions []string
    
    // 根据问题类型生成建议
    hasCacheIssues := false
    hasPerformanceIssues := false
    hasErrorIssues := false
    
    for _, result := range results {
        switch result.Rule {
        case "cache_hit_rate_low":
            hasCacheIssues = true
        case "response_time_high":
            hasPerformanceIssues = true
        case "error_rate_high":
            hasErrorIssues = true
        }
    }
    
    if hasCacheIssues {
        suggestions = append(suggestions, "优化缓存策略和配置")
    }
    
    if hasPerformanceIssues {
        suggestions = append(suggestions, "优化算法实现和数据结构")
    }
    
    if hasErrorIssues {
        suggestions = append(suggestions, "加强错误处理和监控")
    }
    
    if len(suggestions) == 0 {
        suggestions = append(suggestions, "系统运行良好，继续保持")
    }
    
    return suggestions
}
```

### A5 验证（Assure）
- **测试用例**：
  - 性能监控功能测试
  - 链路追踪功能测试
  - 缓存优化功能测试
  - 负载均衡优化测试
  - 性能分析功能测试
- **性能验证**：
  - 性能提升基准测试
  - 监控开销测试
  - 缓存命中率测试
- **回归测试**：
  - 确保优化不影响现有功能
- **测试结果**：
  - 性能提升 ≥ 20%
  - 监控开销 < 5%
  - 缓存命中率提升 ≥ 10%

### A6 迭代（Advance）
- 性能优化：进一步算法优化、缓存策略优化
- 功能扩展：支持更多监控指标、分析规则
- 观测性增强：添加告警机制、自动优化建议
- 下一步任务链接：完成整个模块的6A迁移

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
[总结任务完成情况]
