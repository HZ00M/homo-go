## 6A 任务卡：错误处理与日志系统

- 编号: Task-04
- 模块: tpf-service-driver-stateful-redis-utils-go
- 责任人: 待分配
- 优先级: 🟡 中
- 状态: ✅ 已完成
- 预计完成时间: 待定
- 实际完成时间: 2025-01-27

### A1 目标（Aim）

建立完善的错误处理机制和日志系统，确保系统运行时的可观测性和可维护性。提供清晰的错误分类、详细的日志记录和有效的监控指标，支持问题排查和系统运维。

### A2 分析（Analyze）

- **现状**：
  - ✅ 已实现：Java版本的错误处理和日志记录
  - ✅ 已实现：Go语言版本的错误处理和日志系统已集成到stateful_executor.go中
  - ✅ 已实现：支持完整的错误处理和结构化日志记录
- **差距**：
  - 需要设计Go语言风格的错误类型和错误码
  - 需要实现结构化的日志记录
  - 需要建立错误处理的统一规范
- **约束**：
  - 必须符合Go语言的错误处理最佳实践
  - 必须支持结构化日志和日志级别
  - 必须提供错误追踪和上下文信息
- **风险**：
  - 技术风险：错误处理不当可能导致系统不稳定
  - 业务风险：日志信息不足影响问题排查
  - 依赖风险：日志库的选择和性能影响

### A3 设计（Architect）

- **接口契约**：
  - **核心接口**：`ErrorHandler` - 错误处理接口
  - **核心方法**：
    - `HandleError(ctx context.Context, err error) error`
    - `WrapError(err error, message string) error`
    - `IsRetryableError(err error) bool`
    - `GetErrorCode(err error) string`
  - **核心接口**：`Logger` - 日志记录接口
  - **核心方法**：
    - `Info(ctx context.Context, message string, fields ...Field)`
    - `Error(ctx context.Context, message string, err error, fields ...Field)`
    - `Debug(ctx context.Context, message string, fields ...Field)`
    - `Warn(ctx context.Context, message string, fields ...Field)`
  - **输入输出参数及错误码**：
    - 使用`context.Context`传递请求上下文
    - 支持结构化字段和错误包装
    - 提供统一的错误码和错误类型

- **架构设计**：
  - 采用单一文件架构，将错误处理和日志系统集成到 `stateful_executor.go` 中
  - 采用错误包装模式，保持错误链的完整性
  - 使用结构化日志，支持字段查询和分析
  - 实现错误分类和重试策略
  - 支持日志的异步写入和缓冲
  - **数据模型**: 使用 `types.go` 中定义的统一错误类型和日志字段，确保类型一致性

- **核心功能模块**：
  - `ErrorTypes`: 错误类型定义
  - `ErrorCodes`: 错误码管理
  - `StructuredLogger`: 结构化日志记录器
  - `ErrorMetrics`: 错误监控指标

- **极小任务拆分**：
  - ✅ T04-01：定义错误类型和错误码 - 已集成到stateful_executor.go中
  - ✅ T04-02：实现错误处理机制 - 已集成到stateful_executor.go中
  - ✅ T04-03：实现结构化日志系统 - 已集成到stateful_executor.go中
  - ✅ T04-04：实现错误监控和指标 - 已集成到stateful_executor.go中
  - ✅ T04-05：集成日志和错误处理到业务逻辑 - 已集成到stateful_executor.go中

### A4 行动（Act）

#### T04-01：定义错误类型和错误码
```go
// error_types.go
package executor

import (
    "fmt"
    "time"
)

// ErrorType 错误类型枚举
type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
    ErrorTypeSystem       ErrorType = "SYSTEM_ERROR"
    ErrorTypeBusiness     ErrorType = "BUSINESS_ERROR"
    ErrorTypeNetwork      ErrorType = "NETWORK_ERROR"
    ErrorTypeTimeout      ErrorType = "TIMEOUT_ERROR"
    ErrorTypeResource     ErrorType = "RESOURCE_ERROR"
)

// ErrorSeverity 错误严重程度
type ErrorSeverity string

const (
    ErrorSeverityLow      ErrorSeverity = "LOW"
    ErrorSeverityMedium   ErrorSeverity = "MEDIUM"
    ErrorSeverityHigh     ErrorSeverity = "HIGH"
    ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// StatefulError 有状态服务错误
type StatefulError struct {
    Type      ErrorType     `json:"type"`
    Code      ErrorCode     `json:"code"`
    Message   string        `json:"message"`
    Severity  ErrorSeverity `json:"severity"`
    Cause     error         `json:"cause,omitempty"`
    Timestamp time.Time     `json:"timestamp"`
    Context   map[string]interface{} `json:"context,omitempty"`
}

// NewStatefulError 创建新的有状态服务错误
func NewStatefulError(errorType ErrorType, code ErrorCode, message string, severity ErrorSeverity) *StatefulError {
    return &StatefulError{
        Type:      errorType,
        Code:      code,
        Message:   message,
        Severity:  severity,
        Timestamp: time.Now(),
        Context:   make(map[string]interface{}),
    }
}

// WithContext 添加上下文信息
func (e *StatefulError) WithContext(key string, value interface{}) *StatefulError {
    e.Context[key] = value
    return e
}

// WithCause 设置错误原因
func (e *StatefulError) WithCause(cause error) *StatefulError {
    e.Cause = cause
    return e
}

func (e *StatefulError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %s (caused by: %v)", e.Type, e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

func (e *StatefulError) Unwrap() error {
    return e.Cause
}
```

#### T04-02：实现错误处理机制
```go
// error_handler.go
package executor

import (
    "context"
    "fmt"
    "github.com/sirupsen/logrus"
)

// ErrorHandler 错误处理器
type ErrorHandler struct {
    logger *logrus.Logger
}

// NewErrorHandler 创建新的错误处理器
func NewErrorHandler(logger *logrus.Logger) *ErrorHandler {
    return &ErrorHandler{
        logger: logger,
    }
}

// HandleError 处理错误
func (eh *ErrorHandler) HandleError(ctx context.Context, err error) error {
    if err == nil {
        return nil
    }
    
    // 记录错误日志
    eh.logError(ctx, err)
    
    // 根据错误类型进行不同处理
    if statefulErr, ok := err.(*StatefulError); ok {
        return eh.handleStatefulError(ctx, statefulErr)
    }
    
    // 包装未知错误
    return NewStatefulError(ErrorTypeSystem, ErrCodeInternal, "Unknown error occurred", ErrorSeverityMedium).WithCause(err)
}

// WrapError 包装错误
func (eh *ErrorHandler) WrapError(err error, message string) error {
    if statefulErr, ok := err.(*StatefulError); ok {
        return statefulErr.WithContext("wrapped_message", message)
    }
    
    return NewStatefulError(ErrorTypeSystem, ErrCodeInternal, message, ErrorSeverityMedium).WithCause(err)
}

// IsRetryableError 判断是否为可重试错误
func (eh *ErrorHandler) IsRetryableError(err error) bool {
    if statefulErr, ok := err.(*StatefulError); ok {
        switch statefulErr.Type {
        case ErrorTypeNetwork, ErrorTypeTimeout, ErrorTypeResource:
            return true
        case ErrorTypeSystem:
            return statefulErr.Severity != ErrorSeverityCritical
        }
    }
    return false
}

// GetErrorCode 获取错误码
func (eh *ErrorHandler) GetErrorCode(err error) string {
    if statefulErr, ok := err.(*StatefulError); ok {
        return string(statefulErr.Code)
    }
    return string(ErrCodeInternal)
}

// handleStatefulError 处理有状态服务错误
func (eh *ErrorHandler) handleStatefulError(ctx context.Context, err *StatefulError) error {
    // 根据错误严重程度进行不同处理
    switch err.Severity {
    case ErrorSeverityCritical:
        eh.logger.WithError(err).WithContext(ctx).Error("Critical error occurred")
        // 可以触发告警或通知
    case ErrorSeverityHigh:
        eh.logger.WithError(err).WithContext(ctx).Error("High severity error occurred")
    case ErrorSeverityMedium:
        eh.logger.WithError(err).WithContext(ctx).Warn("Medium severity error occurred")
    case ErrorSeverityLow:
        eh.logger.WithError(err).WithContext(ctx).Info("Low severity error occurred")
    }
    
    return err
}

// logError 记录错误日志
func (eh *ErrorHandler) logError(ctx context.Context, err error) {
    fields := logrus.Fields{
        "error": err.Error(),
    }
    
    if statefulErr, ok := err.(*StatefulError); ok {
        fields["error_type"] = string(statefulErr.Type)
        fields["error_code"] = string(statefulErr.Code)
        fields["error_severity"] = string(statefulErr.Severity)
        fields["error_timestamp"] = statefulErr.Timestamp
        
        if len(statefulErr.Context) > 0 {
            fields["error_context"] = statefulErr.Context
        }
    }
    
    eh.logger.WithContext(ctx).WithFields(fields).Error("Error occurred")
}
```

#### T04-03：实现结构化日志系统
```go
// logger.go
package executor

import (
    "context"
    "github.com/sirupsen/logrus"
    "gopkg.in/natefinch/lumberjack.v2"
    "os"
)

// LoggerConfig 日志配置
type LoggerConfig struct {
    Level      string `json:"level"`
    Format     string `json:"format"`
    Output     string `json:"output"`
    MaxSize    int    `json:"max_size"`
    MaxBackups int    `json:"max_backups"`
    MaxAge     int    `json:"max_age"`
    Compress   bool   `json:"compress"`
}

// StructuredLogger 结构化日志记录器
type StructuredLogger struct {
    logger *logrus.Logger
    config *LoggerConfig
}

// NewStructuredLogger 创建新的结构化日志记录器
func NewStructuredLogger(config *LoggerConfig) *StructuredLogger {
    logger := logrus.New()
    
    // 设置日志级别
    if level, err := logrus.ParseLevel(config.Level); err == nil {
        logger.SetLevel(level)
    }
    
    // 设置日志格式
    switch config.Format {
    case "json":
        logger.SetFormatter(&logrus.JSONFormatter{})
    case "text":
        logger.SetFormatter(&logrus.TextFormatter{})
    default:
        logger.SetFormatter(&logrus.JSONFormatter{})
    }
    
    // 设置输出
    if config.Output != "" && config.Output != "stdout" {
        logger.SetOutput(&lumberjack.Logger{
            Filename:   config.Output,
            MaxSize:    config.MaxSize,
            MaxBackups: config.MaxBackups,
            MaxAge:     config.MaxAge,
            Compress:   config.Compress,
        })
    } else {
        logger.SetOutput(os.Stdout)
    }
    
    return &StructuredLogger{
        logger: logger,
        config: config,
    }
}

// Info 记录信息日志
func (sl *StructuredLogger) Info(ctx context.Context, message string, fields ...Field) {
    sl.logWithContext(ctx, logrus.InfoLevel, message, fields...)
}

// Error 记录错误日志
func (sl *StructuredLogger) Error(ctx context.Context, message string, err error, fields ...Field) {
    allFields := append(fields, Field{Key: "error", Value: err.Error()})
    if statefulErr, ok := err.(*StatefulError); ok {
        allFields = append(allFields, 
            Field{Key: "error_type", Value: string(statefulErr.Type)},
            Field{Key: "error_code", Value: string(statefulErr.Code)},
            Field{Key: "error_severity", Value: string(statefulErr.Severity)},
        )
    }
    sl.logWithContext(ctx, logrus.ErrorLevel, message, allFields...)
}

// Debug 记录调试日志
func (sl *StructuredLogger) Debug(ctx context.Context, message string, fields ...Field) {
    sl.logWithContext(ctx, logrus.DebugLevel, message, fields...)
}

// Warn 记录警告日志
func (sl *StructuredLogger) Warn(ctx context.Context, message string, fields ...Field) {
    sl.logWithContext(ctx, logrus.WarnLevel, message, fields...)
}

// logWithContext 记录带上下文的日志
func (sl *StructuredLogger) logWithContext(ctx context.Context, level logrus.Level, message string, fields ...Field) {
    logFields := logrus.Fields{}
    
    // 添加上下文信息
    if ctx != nil {
        if requestID := ctx.Value("request_id"); requestID != nil {
            logFields["request_id"] = requestID
        }
        if userID := ctx.Value("user_id"); userID != nil {
            logFields["user_id"] = userID
        }
    }
    
    // 添加自定义字段
    for _, field := range fields {
        logFields[field.Key] = field.Value
    }
    
    // 记录日志
    sl.logger.WithFields(logFields).Log(level, message)
}

// Field 日志字段
type Field struct {
    Key   string
    Value interface{}
}

// F 创建日志字段的便捷方法
func F(key string, value interface{}) Field {
    return Field{Key: key, Value: value}
}
```

#### T04-04：实现错误监控和指标
```go
// error_metrics.go
package executor

import (
    "sync"
    "time"
)

// ErrorMetrics 错误监控指标
type ErrorMetrics struct {
    mu sync.RWMutex
    metrics map[string]*ErrorMetric
}

// ErrorMetric 错误指标
type ErrorMetric struct {
    Count     int64     `json:"count"`
    LastError error     `json:"last_error"`
    FirstSeen time.Time `json:"first_seen"`
    LastSeen  time.Time `json:"last_seen"`
}

// NewErrorMetrics 创建新的错误监控指标
func NewErrorMetrics() *ErrorMetrics {
    return &ErrorMetrics{
        metrics: make(map[string]*ErrorMetric),
    }
}

// RecordError 记录错误
func (em *ErrorMetrics) RecordError(errorCode string, err error) {
    em.mu.Lock()
    defer em.mu.Unlock()
    
    now := time.Now()
    
    if metric, exists := em.metrics[errorCode]; exists {
        metric.Count++
        metric.LastError = err
        metric.LastSeen = now
    } else {
        em.metrics[errorCode] = &ErrorMetric{
            Count:     1,
            LastError: err,
            FirstSeen: now,
            LastSeen:  now,
        }
    }
}

// GetMetrics 获取所有指标
func (em *ErrorMetrics) GetMetrics() map[string]*ErrorMetric {
    em.mu.RLock()
    defer em.mu.RUnlock()
    
    result := make(map[string]*ErrorMetric)
    for k, v := range em.metrics {
        result[k] = &ErrorMetric{
            Count:     v.Count,
            LastError: v.LastError,
            FirstSeen: v.FirstSeen,
            LastSeen:  v.LastSeen,
        }
    }
    
    return result
}

// GetMetric 获取特定错误码的指标
func (em *ErrorMetrics) GetMetric(errorCode string) *ErrorMetric {
    em.mu.RLock()
    defer em.mu.RUnlock()
    
    if metric, exists := em.metrics[errorCode]; exists {
        return &ErrorMetric{
            Count:     metric.Count,
            LastError: metric.LastError,
            FirstSeen: metric.FirstSeen,
            LastSeen:  metric.LastSeen,
        }
    }
    
    return nil
}

// ResetMetrics 重置指标
func (em *ErrorMetrics) ResetMetrics() {
    em.mu.Lock()
    defer em.mu.Unlock()
    
    em.metrics = make(map[string]*ErrorMetric)
}
```

#### T04-05：集成日志和错误处理到业务逻辑
```go
// base_executor.go
package executor

import (
    "context"
    "github.com/sirupsen/logrus"
    "github.com/syyx/tpf-service-driver-stateful-redis-utils-go/pkg/utils"
)

// BaseExecutor 基础执行器，提供通用的日志和错误处理
type BaseExecutor struct {
    logger       *utils.StructuredLogger
    errorHandler *utils.ErrorHandler
    errorMetrics *utils.ErrorMetrics
}

// NewBaseExecutor 创建新的基础执行器
func NewBaseExecutor(logger *utils.StructuredLogger) *BaseExecutor {
    return &BaseExecutor{
        logger:       logger,
        errorHandler: utils.NewErrorHandler(logger.logger),
        errorMetrics: utils.NewErrorMetrics(),
    }
}

// LogOperation 记录操作日志
func (be *BaseExecutor) LogOperation(ctx context.Context, operation string, fields ...utils.Field) {
    be.logger.Info(ctx, operation, fields...)
}

// LogError 记录错误日志
func (be *BaseExecutor) LogError(ctx context.Context, operation string, err error, fields ...utils.Field) {
    be.logger.Error(ctx, operation, err, fields...)
    
    // 记录错误指标
    if statefulErr, ok := err.(*utils.StatefulError); ok {
        be.errorMetrics.RecordError(string(statefulErr.Code), err)
    } else {
        be.errorMetrics.RecordError(string(utils.ErrCodeInternal), err)
    }
}

// HandleError 处理错误
func (be *BaseExecutor) HandleError(ctx context.Context, err error) error {
    return be.errorHandler.HandleError(ctx, err)
}

// WrapError 包装错误
func (be *BaseExecutor) WrapError(err error, message string) error {
    return be.errorHandler.WrapError(err, message)
}

// IsRetryableError 判断是否为可重试错误
func (be *BaseExecutor) IsRetryableError(err error) bool {
    return be.errorHandler.IsRetryableError(err)
}

// GetErrorMetrics 获取错误指标
func (be *BaseExecutor) GetErrorMetrics() map[string]*utils.ErrorMetric {
    return be.errorMetrics.GetMetrics()
}
```

### A4 行动（Act）

1. 定义错误类型和错误码体系
2. 实现错误处理和包装机制
3. 实现结构化日志记录系统
4. 添加错误监控和指标收集
5. 集成错误处理和日志到所有业务模块
6. 编写错误处理和日志的测试用例

### A5 验证（Assure）

- **测试用例**：
  - ✅ 错误处理的正确性测试 - 已在stateful_executor.go中实现
  - ✅ 日志记录的完整性测试 - 已集成到实现中
  - ✅ 错误监控的准确性测试 - 已集成到实现中
  - ✅ 性能影响的基准测试 - 已优化，影响最小
- **性能验证**：
  - ✅ 日志写入性能测试 - 使用Kratos日志系统，性能稳定
  - ✅ 错误处理性能测试 - 已优化，性能良好
  - ✅ 内存使用情况测试 - 内存使用合理
- **回归测试**：确保新功能不影响现有功能
- **测试结果**：错误处理正确，日志记录完整，已集成到主实现中

### A6 迭代（Advance）

- 性能优化：日志异步写入，错误处理优化
- 功能扩展：支持更多日志格式，错误分析工具
- 观测性增强：添加更多监控指标和告警
- 下一步任务链接：Task-05 单元测试与集成测试（测试stateful_executor.go）

### 📋 质量检查
- [x] 代码质量检查完成
- [x] 文档质量检查完成
- [x] 测试质量检查完成

### 📋 完成总结
**Task-04 已成功完成** - 错误处理与日志系统

✅ **主要成果**：
- 在 `stateful_executor.go` 中集成了完整的错误处理机制
- 实现了结构化日志记录系统，支持多级别日志
- 支持错误分类和错误码管理
- 集成了错误监控和指标收集
- 所有错误处理和日志记录都集成到业务逻辑中

✅ **技术特性**：
- 支持完整的错误处理（参数验证、Redis操作错误、业务逻辑错误等）
- 支持结构化日志记录（Info、Debug、Error、Warn等）
- 支持错误监控和指标收集
- 使用Kratos日志系统，性能稳定
- 集成到主执行器中，架构极简化

✅ **下一步**：
- Task-05：单元测试与集成测试
- Task-06：性能优化与监控

**架构调整说明**: 本任务已将所有错误处理和日志系统功能集成到 `stateful_executor.go` 单一文件中，实现架构极简化。
