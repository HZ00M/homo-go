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

// ==================== State模块接口定义 ====================

// StatefulRouteForClientDriver 有状态客户端驱动接口
type StatefulRouteForClientDriver interface {
	// 计算一个最佳podId
	ComputeLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)

	// 获取uid下的连接信息
	GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error)

	// 从缓存或注册中心获取连接信息
	GetLinkedPodInCacheOrIfPersist(ctx context.Context, namespace, uid, serviceName string) (int, error)

	// 直接从注册中心获取连接信息
	GetLinkedPodNotCache(ctx context.Context, namespace, uid, serviceName string) (int, error)

	// 获取缓存中的连接信息
	GetLinkedInCache(ctx context.Context, namespace, uid, serviceName string) (int, error)

	// 标识连接信息过期
	CacheExpired(ctx context.Context, namespace, uid, serviceName string) (int, error)

	// 设置连接信息到缓存
	SetCache(ctx context.Context, namespace, uid, serviceName string, podIndex int) (int, error)

	// 获取用户连接的所有服务
	GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error)

	// 批量获取连接信息
	BatchGetLinkedPod(ctx context.Context, namespace string, keys []string, serviceName string) (map[int][]string, error)

	// 设置链接pod（如果不存在）
	SetLinkedPodIfAbsent(ctx context.Context, namespace, uid, serviceName string, podIndex int) (int, error)

	// 初始化
	Init() error

	// 关闭
	Close() error
}

// StatefulRouteForServerDriver 有状态服务端驱动接口
type StatefulRouteForServerDriver interface {
	// 设置服务器负载状态
	SetLoadState(ctx context.Context, loadState int) error

	// 获取服务器负载状态
	GetLoadState(ctx context.Context) (int, error)

	// 设置路由状态
	SetRoutingState(ctx context.Context, state RoutingState) error

	// 获取路由状态
	GetRoutingState(ctx context.Context) (RoutingState, error)

	// 设置连接信息
	SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId int) (int, error)

	// 尝试设置指定uid连接的podId
	TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId int) (bool, int, error)

	// 移除服务连接信息
	RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string) (bool, error)

	// 移除指定uid连接的podId
	RemoveLinkedPodWithId(ctx context.Context, namespace, uid, serviceName string, podId, persistSeconds int) (bool, error)

	// 启动服务端驱动
	Start(ctx context.Context) error

	// 停止服务端驱动
	Stop(ctx context.Context) error
}
