package route

import (
	"context"
)

// ==================== 核心接口定义 ====================

// RouteInfoDriver 路由信息驱动接口
// 定义所有有状态操作的核心接口
type RouteInfoDriver interface {
	// 获取链接信息缓存时间（秒）
	GetLinkInfoCacheTimeSecs() int

	// 注册路由状态变更事件
	RegisterRoutingStateChangedEvent(namespace string, stateChanged StateChanged)

	// 路由状态变更通知
	OnRoutingStateChanged(namespace, serviceName string, podIndex int, pre, now *StatefulServiceState)

	// 获取就绪的服务状态
	GetReadyServiceState(ctx context.Context, namespace, serviceName string) (map[int]*StatefulServiceState, error)

	// 获取所有服务状态
	GetAllServiceState(ctx context.Context, namespace, serviceName string) (map[int]*StatefulServiceState, error)

	// 根据标签获取服务名称
	GetServiceNameByTag(ctx context.Context, namespace, tag string) (string, error)

	// 设置全局服务名称标签
	SetGlobalServiceNameTag(ctx context.Context, namespace, tag, serviceName string) (bool, error)

	// 检查Pod是否可路由
	IsPodRoutable(namespace, serviceName string, podIndex int) bool

	// 获取存活的Pod列表
	AlivePods(namespace, serviceName string) map[int]*StatefulServiceState

	// 获取可路由的Pod列表
	RoutablePods(namespace, serviceName string) map[int]*StatefulServiceState

	// 获取服务最佳Pod
	GetServiceBestPod(ctx context.Context, namespace, serviceName string) (int, error)

	// 检查Pod是否可用
	IsPodAvailable(namespace, serviceName string, podIndex int) bool

	// 检查工作负载是否就绪
	IsWorkloadReady(ctx context.Context, namespace, serviceName string) (bool, error)
}

// ==================== 执行器接口定义 ====================

// StatefulExecutor 有状态执行器接口
// 定义有状态服务的核心操作接口
// 与Java版本的StatefulRedisExecutor功能完全一致
type StatefulExecutor interface {
	// 服务状态管理
	SetServiceState(ctx context.Context, namespace, serviceName string, podID int, state string) error
	GetServiceState(ctx context.Context, namespace, serviceName string) (map[int]string, error)

	// 工作负载状态管理
	SetWorkloadState(ctx context.Context, namespace, serviceName, state string) error
	GetWorkloadState(ctx context.Context, namespace, serviceName string) (string, error)
	GetWorkloadStateBatch(ctx context.Context, namespace string, serviceNames []string) (map[string]string, error)

	// Pod链接管理
	SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (int, error)
	TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (bool, int, error)
	SetLinkedPodIfAbsent(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (int, error)
	GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)
	GetLinkedPodIfPersist(ctx context.Context, namespace, uid, serviceName string) (int, error)
	BatchGetLinkedPod(ctx context.Context, namespace string, keys []string, serviceName string) (map[int][]string, error)
	RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string, persistSeconds int) (bool, error)
	RemoveLinkedPodWithId(ctx context.Context, namespace, uid, serviceName string, persistSeconds, podID int) (bool, error)
	GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error)
}

// ==================== 缓存接口定义 ====================

// ServiceStateCache 服务状态缓存接口
// 管理服务状态的缓存和更新
type ServiceStateCache interface {
	// 获取初始化顺序
	GetOrder() int

	// 初始化模块
	OnInitModule() error

	// 获取服务最佳Pod
	GetServiceBestPod(ctx context.Context, namespace, serviceName string) (int, error)

	// 检查Pod是否可用
	IsPodAvailable(namespace, serviceName string, podIndex int) bool

	// 检查Pod是否可路由
	IsPodRoutable(namespace, serviceName string, podIndex int) bool

	// 获取存活的Pod列表
	AlivePods(namespace, serviceName string) map[int]*StatefulServiceState

	// 获取可路由的Pod列表
	RoutablePods(namespace, serviceName string) map[int]*StatefulServiceState

	// 运行缓存更新
	Run()
}

// ==================== 模块初始化接口 ====================

// ModuleInitializer 模块初始化接口
// 定义模块的初始化生命周期
type ModuleInitializer interface {
	// 获取初始化顺序（数字越小优先级越高）
	GetOrder() int

	// 初始化模块
	OnInitModule() error
}

// ==================== 工厂接口定义 ====================

// RouteDriverFactory 路由驱动工厂接口
// 创建和管理路由驱动实例
type RouteDriverFactory interface {
	// 创建路由驱动
	CreateDriver(config *StatefulBaseConfig) (RouteInfoDriver, error)

	// 获取默认驱动
	GetDefaultDriver() RouteInfoDriver

	// 关闭所有驱动
	CloseAll() error
}
