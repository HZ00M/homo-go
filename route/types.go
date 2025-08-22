package route

import (
	"time"
)

// ==================== 基础枚举定义 ====================

// ServiceState 服务状态枚举
type ServiceState int

const (
	ServiceStateUnknown     ServiceState = iota
	ServiceStateReady                    // 服务就绪
	ServiceStateNotReady                 // 服务未就绪
	ServiceStateFailed                   // 服务失败
	ServiceStatePending                  // 服务等待中
	ServiceStateTerminating              // 服务终止中
)

// String 返回服务状态的字符串表示
func (s ServiceState) String() string {
	switch s {
	case ServiceStateReady:
		return "Ready"
	case ServiceStateNotReady:
		return "NotReady"
	case ServiceStateFailed:
		return "Failed"
	case ServiceStatePending:
		return "Pending"
	case ServiceStateTerminating:
		return "Terminating"
	default:
		return "Unknown"
	}
}

// LoadState 负载状态枚举
type LoadState int

const (
	LoadStateUnknown    LoadState = iota
	LoadStateIdle                 // 空闲
	LoadStateLow                  // 低负载
	LoadStateMedium               // 中等负载
	LoadStateHigh                 // 高负载
	LoadStateOverloaded           // 过载
)

// String 返回负载状态的字符串表示
func (l LoadState) String() string {
	switch l {
	case LoadStateIdle:
		return "Idle"
	case LoadStateLow:
		return "Low"
	case LoadStateMedium:
		return "Medium"
	case LoadStateHigh:
		return "High"
	case LoadStateOverloaded:
		return "Overloaded"
	default:
		return "Unknown"
	}
}

// RoutingState 路由状态枚举
type RoutingState int

const (
	RoutingStateUnknown     RoutingState = iota
	RoutingStateReady                    // 可以接收新请求
	RoutingStateDenyNewCall              // 拒绝新请求
	RoutingStateShutdown                 // 关闭状态
)

// String 返回路由状态的字符串表示
func (r RoutingState) String() string {
	switch r {
	case RoutingStateReady:
		return "Ready"
	case RoutingStateDenyNewCall:
		return "DenyNewCall"
	case RoutingStateShutdown:
		return "Shutdown"
	default:
		return "Unknown"
	}
}

// ==================== 核心数据结构 ====================

// StatefulServiceState 有状态服务状态
type StatefulServiceState struct {
	PodID        int               `json:"podId"`        // Pod索引
	State        ServiceState      `json:"state"`        // 服务状态
	LoadState    LoadState         `json:"loadState"`    // 负载状态
	RoutingState RoutingState      `json:"routingState"` // 路由状态
	UpdateTime   int64             `json:"updateTime"`   // 更新时间戳
	Metadata     map[string]string `json:"metadata"`     // 元数据
}

// ServiceStateInfo 服务状态信息
type ServiceStateInfo struct {
	Namespace   string                        `json:"namespace"`   // 命名空间
	ServiceName string                        `json:"serviceName"` // 服务名称
	PodStates   map[int]*StatefulServiceState `json:"podStates"`   // Pod状态映射
	UpdateTime  int64                         `json:"updateTime"`  // 更新时间戳
}

// PodLinkInfo Pod链接信息
type PodLinkInfo struct {
	UID            string `json:"uid"`            // 用户ID
	Namespace      string `json:"namespace"`      // 命名空间
	ServiceName    string `json:"serviceName"`    // 服务名称
	PodID          int    `json:"podId"`          // Pod索引
	PersistTime    int64  `json:"persistTime"`    // 持久化时间
	CreateTime     int64  `json:"createTime"`     // 创建时间
	LastAccessTime int64  `json:"lastAccessTime"` // 最后访问时间
}

// WorkloadState 工作负载状态
type WorkloadState struct {
	Namespace     string            `json:"namespace"`     // 命名空间
	ServiceName   string            `json:"serviceName"`   // 服务名称
	State         ServiceState      `json:"state"`         // 工作负载状态
	Replicas      int               `json:"replicas"`      // 副本数
	ReadyReplicas int               `json:"readyReplicas"` // 就绪副本数
	UpdateTime    int64             `json:"updateTime"`    // 更新时间戳
	Metadata      map[string]string `json:"metadata"`      // 元数据
}

// ServiceTag 服务标签
type ServiceTag struct {
	Namespace   string `json:"namespace"`   // 命名空间
	Tag         string `json:"tag"`         // 标签名
	ServiceName string `json:"serviceName"` // 服务名称
	CreateTime  int64  `json:"createTime"`  // 创建时间
	UpdateTime  int64  `json:"updateTime"`  // 更新时间
}

// ==================== 基础配置 ====================

// StatefulBaseConfig 有状态服务基础配置
type StatefulBaseConfig struct {
	LinkInfoCacheTimeSecs int    `json:"linkInfoCacheTimeSecs"` // 链接信息缓存时间（秒）
	StateCacheExpireSecs  int    `json:"stateCacheExpireSecs"`  // 状态缓存过期时间（秒）
	MaxRetryCount         int    `json:"maxRetryCount"`         // 最大重试次数
	RetryDelayMs          int    `json:"retryDelayMs"`          // 重试延迟（毫秒）
	HealthCheckIntervalMs int    `json:"healthCheckIntervalMs"` // 健康检查间隔（毫秒）
	LogLevel              string `json:"logLevel"`              // 日志级别
	EnableMetrics         bool   `json:"enableMetrics"`         // 是否启用指标
	EnableTracing         bool   `json:"enableTracing"`         // 是否启用追踪
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addresses    []string      `json:"addresses"`    // Redis地址列表
	Password     string        `json:"password"`     // Redis密码
	DB           int           `json:"db"`           // Redis数据库
	PoolSize     int           `json:"poolSize"`     // 连接池大小
	MinIdleConns int           `json:"minIdleConns"` // 最小空闲连接数
	MaxRetries   int           `json:"maxRetries"`   // 最大重试次数
	DialTimeout  time.Duration `json:"dialTimeout"`  // 连接超时
	ReadTimeout  time.Duration `json:"readTimeout"`  // 读取超时
	WriteTimeout time.Duration `json:"writeTimeout"` // 写入超时
	PoolTimeout  time.Duration `json:"poolTimeout"`  // 连接池超时
	IdleTimeout  time.Duration `json:"idleTimeout"`  // 空闲超时
}

// ==================== 回调接口 ====================

// StateChanged 状态变更回调接口
type StateChanged interface {
	OnStateChanged(namespace, serviceName string, podID int, pre, now *StatefulServiceState)
}

// StatefulObjCallback 有状态对象回调接口
type StatefulObjCallback interface {
	OnSuccess(result interface{})
	OnError(err error)
}

// ServerRedisCallback 服务端Redis回调接口
type ServerRedisCallback interface {
	OnSuccess(result interface{})
	OnError(err error)
}

// ==================== 工具函数 ====================

// IsPodAvailable 检查Pod是否可用
func IsPodAvailable(state *StatefulServiceState) bool {
	if state == nil {
		return false
	}
	return state.State == ServiceStateReady &&
		state.RoutingState == RoutingStateReady
}

// IsPodRoutable 检查Pod是否可路由
func IsPodRoutable(state *StatefulServiceState) bool {
	if state == nil {
		return false
	}
	return state.State == ServiceStateReady &&
		state.RoutingState != RoutingStateShutdown
}

// GetPodLoadScore 获取Pod负载评分（用于负载均衡）
func GetPodLoadScore(state *StatefulServiceState) int {
	if state == nil {
		return 1000 // 不可用Pod给最高分，优先被排除
	}

	// 基础分数
	score := 100

	// 根据负载状态调整分数
	switch state.LoadState {
	case LoadStateIdle:
		score += 50
	case LoadStateLow:
		score += 30
	case LoadStateMedium:
		score += 10
	case LoadStateHigh:
		score -= 20
	case LoadStateOverloaded:
		score -= 50
	}

	// 根据路由状态调整分数
	switch state.RoutingState {
	case RoutingStateReady:
		score += 20
	case RoutingStateDenyNewCall:
		score -= 30
	case RoutingStateShutdown:
		score -= 100
	}

	return score
}

// GetCurrentTimestamp 获取当前时间戳（毫秒）
func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// IsExpired 检查时间戳是否过期
func IsExpired(timestamp, expireSeconds int64) bool {
	now := GetCurrentTimestamp()
	return (now - timestamp) > (expireSeconds * 1000)
}

// ==================== 错误类型定义 ====================

// RouteError 路由错误
type RouteError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func (e *RouteError) Error() string {
	return e.Message
}

// 预定义错误码
const (
	ErrCodeServiceNotFound  = "SERVICE_NOT_FOUND"
	ErrCodePodNotFound      = "POD_NOT_FOUND"
	ErrCodePodNotAvailable  = "POD_NOT_AVAILABLE"
	ErrCodePodNotRoutable   = "POD_NOT_ROUTABLE"
	ErrCodeRedisError       = "REDIS_ERROR"
	ErrCodeInvalidParameter = "INVALID_PARAMETER"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeInternalError    = "INTERNAL_ERROR"
)

// NewRouteError 创建新的路由错误
func NewRouteError(code, message string) *RouteError {
	return &RouteError{
		Code:    code,
		Message: message,
	}
}

// NewRouteErrorWithDetails 创建带详细信息的路由错误
func NewRouteErrorWithDetails(code, message, details string) *RouteError {
	return &RouteError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
