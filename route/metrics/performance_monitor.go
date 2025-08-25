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
	RequestTotal    prometheus.Counter
	RequestDuration prometheus.Histogram
	RequestErrors   prometheus.Counter

	// 缓存相关指标
	CacheHits   prometheus.Counter
	CacheMisses prometheus.Counter
	CacheSize   prometheus.Gauge

	// 负载均衡相关指标
	LoadBalancerRequests prometheus.Counter
	PodSelectionDuration prometheus.Histogram

	// 健康检查相关指标
	HealthCheckTotal    prometheus.Counter
	HealthCheckDuration prometheus.Histogram
	HealthCheckFailures prometheus.Counter
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
		"request_total":  pm.metrics.RequestTotal.Get(),
		"request_errors": pm.metrics.RequestErrors.Get(),
		"cache_size":     pm.metrics.CacheSize.Get(),
	}
}

// AddCustomMetric 添加自定义指标
func (pm *PerformanceMonitor) AddCustomMetric(name string, collector prometheus.Collector) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.customMetrics[name]; exists {
		return prometheus.AlreadyRegisteredError{
			ExistingCollector: pm.customMetrics[name],
			NewCollector:      collector,
		}
	}

	pm.customMetrics[name] = collector
	pm.logger.Log(log.LevelInfo, "Added custom metric", "name", name)
	return nil
}

// GetCustomMetrics 获取自定义指标
func (pm *PerformanceMonitor) GetCustomMetrics() map[string]prometheus.Collector {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]prometheus.Collector)
	for k, v := range pm.customMetrics {
		result[k] = v
	}
	return result
}
