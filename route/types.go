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

// LoadState 负载状态数值（数值越小表示负载越低）
type LoadState int

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

	// 根据负载状态数值调整分数（数值越小表示负载越低，分数越高）
	// 直接使用数值比较，不需要枚举常量
	switch {
	case state.LoadState <= 1:
		score += 50 // 空闲状态，最高分
	case state.LoadState == 2:
		score += 30 // 低负载，高分
	case state.LoadState == 3:
		score += 10 // 中等负载，中等分
	case state.LoadState == 4:
		score -= 20 // 高负载，减分
	case state.LoadState >= 5:
		score -= 50 // 过载状态，大幅减分
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

// ==================== State模块类型定义 ====================

// ServerInfo 服务器信息
type ServerInfo struct {
	Namespace   string // 命名空间
	ServiceName string // 服务名称
	PodIndex    int    // Pod索引
}

// BaseConfig 基础配置
type BaseConfig struct {
	// 服务状态更新周期（秒）
	ServiceStateUpdatePeriodSeconds int
	// CPU因子（用于计算加权状态）
	CpuFactor float64
	// 临时链接信息缓存时间（秒）
	TempLinkInfoCacheTimeSeconds int
	// 用户链接信息缓存时间（秒）
	UserLinkInfoCacheTimeSeconds int
}

// StateObj 状态对象
type StateObj struct {
	State        int       // 状态值
	UpdateTime   time.Time // 更新时间
	RoutingState string    // 路由状态
}

// ToStateString 转换为状态字符串
func (s *StateObj) ToStateString() string {
	// 这里可以根据需要实现状态序列化逻辑
	// 暂时返回简单的字符串表示
	return "state"
}

// DefaultBaseConfig 默认基础配置
func DefaultBaseConfig() *BaseConfig {
	return &BaseConfig{
		ServiceStateUpdatePeriodSeconds: 30,
		CpuFactor:                       0.7,
		TempLinkInfoCacheTimeSeconds:    300,
		UserLinkInfoCacheTimeSeconds:    240, // 比临时缓存短60秒
	}
}
