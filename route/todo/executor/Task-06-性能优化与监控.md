## 6A 任务卡：性能优化与监控

- 编号: Task-06
- 模块: route/executor
- 责任人: 待分配
- 优先级: 🟢 低
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）
为有状态服务执行器实现性能优化和监控功能，提升系统性能和可观测性，支持性能指标收集、性能分析和性能调优，为生产环境提供可靠的性能保障。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：核心功能已在 `stateful_executor.go` 中完成
  - ✅ 已完成：基础性能优化已集成到实现中
  - ❌ 未实现：完整的性能监控系统
  - ❌ 未实现：性能指标收集和分析
- **差距**：
  - 需要实现性能监控指标
  - 需要实现性能分析工具
  - 需要实现性能调优机制
  - 需要实现性能报告生成
- **约束**：
  - 必须保持与Java版本功能的兼容性
  - 必须符合Go语言的性能优化规范
  - 必须支持Kratos框架的监控
- **风险**：
  - 性能优化可能影响功能正确性
  - 监控系统可能增加系统开销

### A3 设计（Architect）
- **接口契约**：
  - **核心功能**：性能监控、性能分析、性能调优
  - **核心特性**：
    - 支持性能指标收集
    - 支持性能分析报告
    - 支持性能调优建议
    - 支持性能基准测试

- **架构设计**：
  - 采用单文件架构，所有性能优化和监控功能集成在 `StatefulExecutorImpl` 中
  - 使用Go标准库进行性能分析
  - 支持Prometheus指标导出
  - 支持性能基准测试

- **核心功能模块**：
  - `PerformanceMonitor`: 性能监控器
  - `MetricsCollector`: 指标收集器
  - `PerformanceAnalyzer`: 性能分析器
  - `OptimizationEngine`: 优化引擎

- **极小任务拆分**：
  - T06-01：实现性能监控指标
  - T06-02：实现性能分析工具
  - T06-03：实现性能调优机制
  - T06-04：实现性能报告生成
  - T06-05：集成Prometheus监控
  - T06-06：实现性能基准测试

### A4 行动（Act）
#### T06-01：实现性能监控指标
```go
// route/executor/performance_monitor.go
package executor

import (
    "sync"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
    // 操作计数
    operationCounter *prometheus.CounterVec
    // 操作延迟
    operationDuration *prometheus.HistogramVec
    // 错误计数
    errorCounter *prometheus.CounterVec
    // 缓存命中率
    cacheHitRatio *prometheus.GaugeVec
    // Redis连接状态
    redisConnectionStatus *prometheus.GaugeVec
}

// NewPerformanceMetrics 创建新的性能指标
func NewPerformanceMetrics() *PerformanceMetrics {
    return &PerformanceMetrics{
        operationCounter: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "stateful_executor_operations_total",
                Help: "Total number of operations",
            },
            []string{"operation", "status"},
        ),
        operationDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "stateful_executor_operation_duration_seconds",
                Help:    "Operation duration in seconds",
                Buckets: prometheus.DefBuckets,
            },
            []string{"operation"},
        ),
        errorCounter: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "stateful_executor_errors_total",
                Help: "Total number of errors",
            },
            []string{"operation", "error_type"},
        ),
        cacheHitRatio: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "stateful_executor_cache_hit_ratio",
                Help: "Cache hit ratio",
            },
            []string{"cache_type"},
        ),
        redisConnectionStatus: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "stateful_executor_redis_connection_status",
                Help: "Redis connection status",
            },
            []string{"status"},
        ),
    }
}

// RecordOperation 记录操作
func (pm *PerformanceMetrics) RecordOperation(operation, status string, duration time.Duration) {
    pm.operationCounter.WithLabelValues(operation, status).Inc()
    pm.operationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordError 记录错误
func (pm *PerformanceMetrics) RecordError(operation, errorType string) {
    pm.errorCounter.WithLabelValues(operation, errorType).Inc()
}

// UpdateCacheHitRatio 更新缓存命中率
func (pm *PerformanceMetrics) UpdateCacheHitRatio(cacheType string, ratio float64) {
    pm.cacheHitRatio.WithLabelValues(cacheType).Set(ratio)
}

// UpdateRedisConnectionStatus 更新Redis连接状态
func (pm *PerformanceMetrics) UpdateRedisConnectionStatus(status string, value float64) {
    pm.redisConnectionStatus.WithLabelValues(status).Set(value)
}
```

#### T06-02：实现性能分析工具
```go
// route/executor/performance_analyzer.go
package executor

import (
    "context"
    "fmt"
    "runtime"
    "runtime/pprof"
    "time"
)

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
    metrics *PerformanceMetrics
    logger  log.Logger
}

// NewPerformanceAnalyzer 创建新的性能分析器
func NewPerformanceAnalyzer(metrics *PerformanceMetrics, logger log.Logger) *PerformanceAnalyzer {
    return &PerformanceAnalyzer{
        metrics: metrics,
        logger:  logger,
    }
}

// AnalyzeOperation 分析操作性能
func (pa *PerformanceAnalyzer) AnalyzeOperation(ctx context.Context, operation string, fn func() error) error {
    start := time.Now()
    
    // 记录内存使用
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    initialAlloc := m.Alloc
    
    // 执行操作
    err := fn()
    
    // 计算性能指标
    duration := time.Since(start)
    runtime.ReadMemStats(&m)
    memoryUsed := m.Alloc - initialAlloc
    
    // 记录指标
    status := "success"
    if err != nil {
        status = "error"
        pa.metrics.RecordError(operation, "execution_error")
    }
    
    pa.metrics.RecordOperation(operation, status, duration)
    
    // 记录性能日志
    pa.logger.Log(log.LevelDebug, "msg", "操作性能分析",
        "operation", operation,
        "duration", duration,
        "memory_used", memoryUsed,
        "status", status)
    
    return err
}

// GeneratePerformanceReport 生成性能报告
func (pa *PerformanceAnalyzer) GeneratePerformanceReport() *PerformanceReport {
    return &PerformanceReport{
        Timestamp: time.Now(),
        Metrics:   pa.metrics,
        // 其他性能数据
    }
}

// PerformanceReport 性能报告
type PerformanceReport struct {
    Timestamp time.Time
    Metrics   *PerformanceMetrics
    // 其他性能数据字段
}
```

#### T06-03：实现性能调优机制
```go
// route/executor/optimization_engine.go
package executor

import (
    "context"
    "time"
)

// OptimizationEngine 优化引擎
type OptimizationEngine struct {
    analyzer *PerformanceAnalyzer
    logger   log.Logger
}

// NewOptimizationEngine 创建新的优化引擎
func NewOptimizationEngine(analyzer *PerformanceAnalyzer, logger log.Logger) *OptimizationEngine {
    return &OptimizationEngine{
        analyzer: analyzer,
        logger:   logger,
    }
}

// OptimizeScriptCache 优化脚本缓存
func (oe *OptimizationEngine) OptimizeScriptCache(ctx context.Context, executor *StatefulExecutorImpl) {
    // 分析脚本缓存使用情况
    cacheStats := executor.GetScriptCacheStats()
    
    // 根据使用频率优化缓存策略
    if cacheStats.HitRatio < 0.8 {
        oe.logger.Log(log.LevelInfo, "msg", "脚本缓存命中率较低，建议优化",
            "hit_ratio", cacheStats.HitRatio)
        
        // 预加载常用脚本
        if err := executor.preloadScripts(ctx); err != nil {
            oe.logger.Log(log.LevelError, "msg", "脚本预加载失败", "error", err)
        }
    }
}

// OptimizeRedisConnections 优化Redis连接
func (oe *OptimizationEngine) OptimizeRedisConnections(ctx context.Context, executor *StatefulExecutorImpl) {
    // 检查Redis连接池状态
    poolStats := executor.GetRedisPoolStats()
    
    if poolStats.ActiveConnections > poolStats.MaxConnections*0.8 {
        oe.logger.Log(log.LevelWarn, "msg", "Redis连接池使用率较高",
            "active_connections", poolStats.ActiveConnections,
            "max_connections", poolStats.MaxConnections)
    }
}

// GetOptimizationSuggestions 获取优化建议
func (oe *OptimizationEngine) GetOptimizationSuggestions() []string {
    suggestions := []string{
        "启用脚本缓存预加载",
        "优化Redis连接池配置",
        "启用批量操作",
        "优化Lua脚本执行",
    }
    
    return suggestions
}
```

#### T06-04：实现性能报告生成
```go
// route/executor/performance_reporter.go
package executor

import (
    "encoding/json"
    "fmt"
    "time"
)

// PerformanceReporter 性能报告器
type PerformanceReporter struct {
    analyzer *PerformanceAnalyzer
    engine   *OptimizationEngine
    logger   log.Logger
}

// NewPerformanceReporter 创建新的性能报告器
func NewPerformanceReporter(analyzer *PerformanceAnalyzer, engine *OptimizationEngine, logger log.Logger) *PerformanceReporter {
    return &PerformanceReporter{
        analyzer: analyzer,
        engine:   engine,
        logger:   logger,
    }
}

// GenerateReport 生成性能报告
func (pr *PerformanceReporter) GenerateReport() (*PerformanceReport, error) {
    report := pr.analyzer.GeneratePerformanceReport()
    
    // 添加优化建议
    suggestions := pr.engine.GetOptimizationSuggestions()
    report.Suggestions = suggestions
    
    return report, nil
}

// ExportReport 导出性能报告
func (pr *PerformanceReporter) ExportReport(format string) ([]byte, error) {
    report, err := pr.GenerateReport()
    if err != nil {
        return nil, err
    }
    
    switch format {
    case "json":
        return json.MarshalIndent(report, "", "  ")
    case "text":
        return pr.formatTextReport(report), nil
    default:
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
}

// formatTextReport 格式化文本报告
func (pr *PerformanceReporter) formatTextReport(report *PerformanceReport) []byte {
    text := fmt.Sprintf(`
性能报告
========
生成时间: %s

性能指标:
- 操作总数: %d
- 错误总数: %d
- 平均延迟: %v

优化建议:
`, report.Timestamp, report.TotalOperations, report.TotalErrors, report.AverageLatency)
    
    for i, suggestion := range report.Suggestions {
        text += fmt.Sprintf("%d. %s\n", i+1, suggestion)
    }
    
    return []byte(text)
}
```

#### T06-05：集成Prometheus监控
```go
// route/executor/prometheus_integration.go
package executor

import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusIntegration Prometheus集成
type PrometheusIntegration struct {
    metrics *PerformanceMetrics
    server  *http.Server
}

// NewPrometheusIntegration 创建新的Prometheus集成
func NewPrometheusIntegration(metrics *PerformanceMetrics, port string) *PrometheusIntegration {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    
    server := &http.Server{
        Addr:    ":" + port,
        Handler: mux,
    }
    
    return &PrometheusIntegration{
        metrics: metrics,
        server:  server,
    }
}

// Start 启动Prometheus监控服务
func (pi *PrometheusIntegration) Start() error {
    return pi.server.ListenAndServe()
}

// Stop 停止Prometheus监控服务
func (pi *PrometheusIntegration) Stop() error {
    return pi.server.Close()
}
```

#### T06-06：实现性能基准测试
```go
// route/executor/performance_benchmark.go
package executor

import (
    "context"
    "testing"
    "time"
)

// BenchmarkSetServiceState 基准测试设置服务状态
func BenchmarkSetServiceState(b *testing.B) {
    // 设置测试环境
    client := setupTestRedisClient()
    defer client.Close()
    
    logger := &MockLogger{}
    executor := NewStatefulExecutor(client, logger)
    
    ctx := context.Background()
    namespace := "benchmark"
    serviceName := "benchmark-service"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        podID := i % 10
        state := "running"
        err := executor.SetServiceState(ctx, namespace, serviceName, podID, state)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// BenchmarkGetServiceState 基准测试获取服务状态
func BenchmarkGetServiceState(b *testing.B) {
    // 设置测试环境
    client := setupTestRedisClient()
    defer client.Close()
    
    logger := &MockLogger{}
    executor := NewStatefulExecutor(client, logger)
    
    ctx := context.Background()
    namespace := "benchmark"
    serviceName := "benchmark-service"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := executor.GetServiceState(ctx, namespace, serviceName)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// setupTestRedisClient 设置测试Redis客户端
func setupTestRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   1,
    })
}
```

### A5 验证（Assure）
- **测试用例**：
  - [ ] 性能监控指标测试
  - [ ] 性能分析工具测试
  - [ ] 性能调优机制测试
  - [ ] 性能报告生成测试
  - [ ] Prometheus集成测试
  - [ ] 性能基准测试
- **性能验证**：
  - [ ] 性能提升 ≥20%
  - [ ] 监控开销 <5%
- **回归测试**：
  - [ ] 所有现有功能测试通过
- **测试结果**：
  - [ ] 待实现

### A6 迭代（Advance）
- 性能优化：基于监控数据进行持续优化
- 功能扩展：支持更多性能指标和分析工具
- 观测性增强：添加更多监控和告警功能
- 下一步任务链接：完成所有任务，进入生产环境部署

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
Task-06待开始，需要实现完整的性能优化和监控功能，包括：
1. 性能监控指标收集
2. 性能分析工具
3. 性能调优机制
4. 性能报告生成
5. Prometheus监控集成
6. 性能基准测试
7. 为生产环境提供性能保障

这是executor模块的最后一个任务，完成后将进入生产环境部署阶段。
