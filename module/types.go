// Package module 提供模块生命周期管理功能
package module

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
)

// LifecycleStage 生命周期阶段
type LifecycleStage int

const (
	StageUnknown LifecycleStage = iota
	StageInit
	StageAfterAllInit
	StageAfterStart
	StageRunning
	StageBeforeClose
	StageAfterStop
)

// String 返回生命周期阶段的字符串表示
func (s LifecycleStage) String() string {
	switch s {
	case StageUnknown:
		return "unknown"
	case StageInit:
		return "init"
	case StageAfterAllInit:
		return "after_all_init"
	case StageAfterStart:
		return "after_start"
	case StageRunning:
		return "running"
	case StageBeforeClose:
		return "before_close"
	case StageAfterStop:
		return "after_stop"
	default:
		return "unknown"
	}
}

// ModuleStatus 模块状态
type ModuleStatus struct {
	Name          string         `json:"name"`
	Order         int            `json:"order"`
	Stage         LifecycleStage `json:"stage"`
	InitTime      time.Time      `json:"init_time,omitempty"`
	StartTime     time.Time      `json:"start_time,omitempty"`
	LastUpdate    time.Time      `json:"last_update,omitempty"`
	LastError     error          `json:"last_error,omitempty"`
	IsInitialized bool           `json:"is_initialized"`
	IsStarted     bool           `json:"is_started"`
	IsRunning     bool           `json:"is_running"`
	IsClosed      bool           `json:"is_closed"`
}

// SetError 设置模块错误
func (ms *ModuleStatus) SetError(err error) {
	ms.LastError = err
	ms.LastUpdate = time.Now()
}

// ClearError 清除模块错误
func (ms *ModuleStatus) ClearError() {
	ms.LastError = nil
	ms.LastUpdate = time.Now()
}

// IsHealthy 检查模块是否健康
func (ms *ModuleStatus) IsHealthy() bool {
	return ms.LastError == nil && !ms.IsClosed
}

// ModuleSummary 模块汇总信息
type ModuleSummary struct {
	TotalCount  int            `json:"total_count"`
	Initialized int            `json:"initialized"`
	Started     int            `json:"started"`
	Running     int            `json:"running"`
	Closed      int            `json:"closed"`
	Failed      int            `json:"failed"`
	StartTime   time.Time      `json:"start_time"`
	Modules     []ModuleStatus `json:"modules"`
}

// 错误码定义
const (
	// ModuleInitFailed 模块初始化失败
	ModuleInitFailed = "MODULE_INIT_FAILED"
	// ModuleCloseFailed 模块关闭失败
	ModuleCloseFailed = "MODULE_CLOSE_FAILED"
	// ModuleInvalid 模块无效
	ModuleInvalid = "MODULE_INVALID"
	// ModuleOrderInvalid 模块优先级无效
	ModuleOrderInvalid = "MODULE_ORDER_INVALID"
)

// NewModuleInitError 创建模块初始化错误
func NewModuleInitError(moduleName string, reason string) error {
	return errors.New(
		500,
		ModuleInitFailed,
		fmt.Sprintf("module %s init failed: %s", moduleName, reason),
	)
}

// NewModuleCloseError 创建模块关闭错误
func NewModuleCloseError(moduleName string, reason string) error {
	return errors.New(
		500,
		ModuleCloseFailed,
		fmt.Sprintf("module %s close failed: %s", moduleName, reason),
	)
}

// NewModuleInvalidError 创建模块无效错误
func NewModuleInvalidError(moduleName string, reason string) error {
	return errors.New(
		400,
		ModuleInvalid,
		fmt.Sprintf("module %s is invalid: %s", moduleName, reason),
	)
}

// NewModuleOrderInvalidError 创建模块优先级无效错误
func NewModuleOrderInvalidError(moduleName string, order int) error {
	return errors.New(
		400,
		ModuleOrderInvalid,
		fmt.Sprintf("module %s has invalid order: %d", moduleName, order),
	)
}

// IsModuleError 判断是否为模块相关错误
func IsModuleError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为 Kratos 错误
	if kratosErr, ok := err.(*errors.Error); ok {
		reason := kratosErr.Reason
		return reason == ModuleInitFailed ||
			reason == ModuleCloseFailed ||
			reason == ModuleInvalid ||
			reason == ModuleOrderInvalid
	}

	return false
}
