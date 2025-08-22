package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
)

// ==================== 错误处理器 ====================

// ErrorHandler 错误处理器
// 负责错误包装、分类和记录
type ErrorHandler struct {
	logger log.Logger
}

// NewErrorHandler 创建新的错误处理器
func NewErrorHandler(logger log.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError 处理错误
func (h *ErrorHandler) HandleError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// 记录错误
	log.NewHelper(h.logger).Error("错误处理", err)

	// 根据错误类型进行分类处理
	if h.IsRetryableError(err) {
		log.NewHelper(h.logger).Warn("可重试错误", err)
		// 这里可以添加重试逻辑
	}

	return err
}

// WrapError 包装错误
func (h *ErrorHandler) WrapError(err error, message string) error {
	if err == nil {
		return fmt.Errorf("%s", message)
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapErrorWithContext 带上下文的错误包装
func (h *ErrorHandler) WrapErrorWithContext(err error, message string, ctx map[string]interface{}) error {
	if err == nil {
		return fmt.Errorf("%s", message)
	}

	// 构建上下文信息
	contextStr := ""
	for k, v := range ctx {
		if contextStr != "" {
			contextStr += ", "
		}
		contextStr += fmt.Sprintf("%s=%v", k, v)
	}

	if contextStr != "" {
		return fmt.Errorf("%s [%s]: %w", message, contextStr, err)
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsRetryableError 检查是否为可重试错误
func (h *ErrorHandler) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// 可重试的错误类型
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"network unreachable",
		"temporary failure",
		"server overloaded",
		"rate limit exceeded",
		"service unavailable",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(errMsg, retryableErr) {
			return true
		}
	}

	return false
}

// GetErrorCode 获取错误码
func (h *ErrorHandler) GetErrorCode(err error) string {
	if err == nil {
		return "UNKNOWN_ERROR"
	}

	errMsg := strings.ToLower(err.Error())

	// 根据错误消息判断错误类型
	if strings.Contains(errMsg, "timeout") {
		return "TIMEOUT"
	}
	if strings.Contains(errMsg, "connection") {
		return "CONNECTION_ERROR"
	}
	if strings.Contains(errMsg, "validation") {
		return "VALIDATION_ERROR"
	}
	if strings.Contains(errMsg, "not found") {
		return "NOT_FOUND"
	}
	if strings.Contains(errMsg, "permission") {
		return "PERMISSION_DENIED"
	}
	if strings.Contains(errMsg, "redis") {
		return "REDIS_ERROR"
	}
	if strings.Contains(errMsg, "json") {
		return "JSON_ERROR"
	}

	return "INTERNAL_ERROR"
}

// ==================== 错误指标收集器 ====================

// ErrorMetrics 错误指标收集器
// 收集和统计各种类型的错误
type ErrorMetrics struct {
	errorCounts map[string]int64
	mutex       sync.RWMutex
}

// NewErrorMetrics 创建新的错误指标收集器
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		errorCounts: make(map[string]int64),
	}
}

// IncrementError 增加错误计数
func (m *ErrorMetrics) IncrementError(errorType string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.errorCounts[errorType]++
}

// GetErrorCount 获取错误计数
func (m *ErrorMetrics) GetErrorCount(errorType string) int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.errorCounts[errorType]
}

// GetAllErrorCounts 获取所有错误计数
func (m *ErrorMetrics) GetAllErrorCounts() map[string]int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]int64)
	for k, v := range m.errorCounts {
		result[k] = v
	}
	return result
}

// ResetErrorCounts 重置错误计数
func (m *ErrorMetrics) ResetErrorCounts() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.errorCounts = make(map[string]int64)
}

// GetTotalErrorCount 获取总错误数
func (m *ErrorMetrics) GetTotalErrorCount() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var total int64
	for _, count := range m.errorCounts {
		total += count
	}
	return total
}

// GetErrorRate 获取错误率（相对于总操作数）
func (m *ErrorMetrics) GetErrorRate(totalOperations int64) float64 {
	if totalOperations == 0 {
		return 0.0
	}

	totalErrors := m.GetTotalErrorCount()
	return float64(totalErrors) / float64(totalOperations)
}

// ==================== 错误类型常量 ====================

const (
	// 连接相关错误
	ErrorTypeRedisConnectionFailed = "redis_connection_failed"
	ErrorTypeRedisGetFailed        = "redis_get_failed"
	ErrorTypeRedisSetFailed        = "redis_set_failed"
	ErrorTypeRedisDelFailed        = "redis_del_failed"
	ErrorTypeRedisExistsFailed     = "redis_exists_failed"

	// 数据相关错误
	ErrorTypeJSONSerializationFailed   = "json_serialization_failed"
	ErrorTypeJSONDeserializationFailed = "json_deserialization_failed"
	ErrorTypeValidationFailed          = "validation_failed"

	// 系统相关错误
	ErrorTypeTimeout          = "timeout"
	ErrorTypeInternalError    = "internal_error"
	ErrorTypeResourceNotFound = "resource_not_found"
	ErrorTypePermissionDenied = "permission_denied"
)
