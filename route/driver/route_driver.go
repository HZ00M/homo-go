package driver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/route"
)

// RouteInfoDriverImpl 有状态服务路由驱动实现
type RouteInfoDriverImpl struct {
	mu  sync.RWMutex
	log log.Logger

	// 配置
	baseConfig *route.StatefulBaseConfig

	// 依赖组件
	stateCache       route.ServiceStateCache
	statefulExecutor route.StatefulExecutor

	// 事件管理
	onStateChange map[string]route.StateChanged

	// 链接信息缓存时间
	linkInfoCacheTimeSecs int

	// 服务标签缓存
	serviceTagCache map[string]map[string]string // namespace -> tag -> serviceName
}

// NewRouteInfoDriverImpl 创建新的路由驱动实现
func NewRouteInfoDriverImpl(
	baseConfig *route.StatefulBaseConfig,
	stateCache route.ServiceStateCache,
	statefulExecutor route.StatefulExecutor,
	logger log.Logger,
) *RouteInfoDriverImpl {
	return &RouteInfoDriverImpl{
		baseConfig:            baseConfig,
		stateCache:            stateCache,
		statefulExecutor:      statefulExecutor,
		log:                   logger,
		onStateChange:         make(map[string]route.StateChanged),
		linkInfoCacheTimeSecs: 300, // 默认5分钟
		serviceTagCache:       make(map[string]map[string]string),
	}
}

// GetLinkInfoCacheTimeSecs 获取链接信息缓存时间
func (r *RouteInfoDriverImpl) GetLinkInfoCacheTimeSecs() int {
	return r.linkInfoCacheTimeSecs
}

// RegisterRoutingStateChangedEvent 注册路由状态变更事件
func (r *RouteInfoDriverImpl) RegisterRoutingStateChangedEvent(namespace string, stateChanged route.StateChanged) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.onStateChange[namespace] = stateChanged
	r.log.Log(log.LevelInfo, "Routing state changed event registered", "namespace", namespace)
}

// OnRoutingStateChanged 处理路由状态变更
func (r *RouteInfoDriverImpl) OnRoutingStateChanged(namespace, serviceName string, podIndex int, pre, now *route.StatefulServiceState) {
	r.mu.RLock()
	stateChanged, exists := r.onStateChange[namespace]
	r.mu.RUnlock()

	if exists {
		logger := r.log // 在goroutine外部获取logger引用
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// 使用外部的logger记录panic
					logger.Log(log.LevelError, "Panic in state changed callback", "recover", r)
				}
			}()

			stateChanged.OnStateChanged(namespace, serviceName, podIndex, pre, now)
		}()
	}

	r.log.Log(log.LevelInfo, "Routing state changed",
		"namespace", namespace, "service", serviceName, "pod", podIndex,
		"pre", pre.State, "now", now.State)
}

// GetReadyServiceState 获取就绪服务状态
func (r *RouteInfoDriverImpl) GetReadyServiceState(ctx context.Context, namespace, serviceName string) (map[int]*route.StatefulServiceState, error) {
	r.log.Log(log.LevelDebug, "Getting ready service state", "namespace", namespace, "service", serviceName)

	// 先尝试从缓存获取
	if r.stateCache != nil {
		routablePods := r.stateCache.RoutablePods(namespace, serviceName)
		if len(routablePods) > 0 {
			return routablePods, nil
		}
	}

	// 从Redis获取所有服务状态
	allServices, err := r.GetAllServiceState(ctx, namespace, serviceName)
	if err != nil {
		return nil, err
	}

	// 过滤就绪状态
	readyServices := make(map[int]*route.StatefulServiceState)
	for podIndex, service := range allServices {
		if r.IsPodRoutable(namespace, serviceName, podIndex) {
			readyServices[podIndex] = service
		}
	}

	return readyServices, nil
}

// GetAllServiceState 获取所有服务状态
func (r *RouteInfoDriverImpl) GetAllServiceState(ctx context.Context, namespace, serviceName string) (map[int]*route.StatefulServiceState, error) {
	r.log.Log(log.LevelDebug, "Getting all service state", "namespace", namespace, "service", serviceName)

	// 从StatefulExecutor获取服务状态
	podStates, err := r.statefulExecutor.GetServiceState(ctx, namespace, serviceName)
	if err != nil {
		return nil, err
	}

	// 转换为StatefulServiceState格式
	services := make(map[int]*route.StatefulServiceState)
	for podID, state := range podStates {
		// 将字符串状态转换为ServiceState枚举
		var serviceState route.ServiceState
		switch state {
		case "READY":
			serviceState = route.ServiceStateReady
		case "NOT_READY":
			serviceState = route.ServiceStateNotReady
		case "FAILED":
			serviceState = route.ServiceStateFailed
		case "PENDING":
			serviceState = route.ServiceStatePending
		case "TERMINATING":
			serviceState = route.ServiceStateTerminating
		default:
			serviceState = route.ServiceStateUnknown
		}

		services[podID] = &route.StatefulServiceState{
			PodID:      podID,
			State:      serviceState,
			UpdateTime: time.Now().Unix(),
			Metadata:   make(map[string]string),
		}
	}

	return services, nil
}

// GetServiceNameByTag 根据标签获取服务名
func (r *RouteInfoDriverImpl) GetServiceNameByTag(ctx context.Context, namespace, tag string) (string, error) {
	r.log.Log(log.LevelDebug, "Getting service name by tag", "namespace", namespace, "tag", tag)

	// 先检查本地缓存
	r.mu.RLock()
	if tagMap, exists := r.serviceTagCache[namespace]; exists {
		if serviceName, exists := tagMap[tag]; exists {
			r.mu.RUnlock()
			return serviceName, nil
		}
	}
	r.mu.RUnlock()

	// 由于StatefulExecutor接口没有直接暴露Redis客户端，我们需要通过其他方式获取标签
	// 这里暂时返回错误，需要完善接口设计
	return "", fmt.Errorf("GetServiceNameByTag not implemented yet - need Redis client access")
}

// SetGlobalServiceNameTag 设置全局服务名标签
func (r *RouteInfoDriverImpl) SetGlobalServiceNameTag(ctx context.Context, namespace, tag, serviceName string) (bool, error) {
	r.log.Log(log.LevelDebug, "Setting global service name tag", "namespace", namespace, "tag", tag, "service", serviceName)

	// 由于StatefulExecutor接口没有直接暴露Redis客户端，我们需要通过其他方式设置标签
	// 这里暂时返回错误，需要完善接口设计
	return false, fmt.Errorf("SetGlobalServiceNameTag not implemented yet - need Redis client access")
}

// IsPodRoutable 检查Pod是否可路由
func (r *RouteInfoDriverImpl) IsPodRoutable(namespace, serviceName string, podIndex int) bool {
	// 优先从缓存检查
	if r.stateCache != nil {
		return r.stateCache.IsPodRoutable(namespace, serviceName, podIndex)
	}

	// 从Redis检查
	ctx := context.Background()
	podStates, err := r.statefulExecutor.GetServiceState(ctx, namespace, serviceName)
	if err != nil {
		return false
	}

	if state, exists := podStates[podIndex]; exists {
		return state == "READY"
	}

	return false
}

// AlivePods 获取存活Pod
func (r *RouteInfoDriverImpl) AlivePods(namespace, serviceName string) map[int]*route.StatefulServiceState {
	if r.stateCache != nil {
		return r.stateCache.AlivePods(namespace, serviceName)
	}

	// 从Redis获取
	ctx := context.Background()
	podStates, err := r.statefulExecutor.GetServiceState(ctx, namespace, serviceName)
	if err != nil {
		return make(map[int]*route.StatefulServiceState)
	}

	services := make(map[int]*route.StatefulServiceState)
	for podID, state := range podStates {
		// 将字符串状态转换为ServiceState枚举
		var serviceState route.ServiceState
		switch state {
		case "READY":
			serviceState = route.ServiceStateReady
		case "NOT_READY":
			serviceState = route.ServiceStateNotReady
		case "FAILED":
			serviceState = route.ServiceStateFailed
		case "PENDING":
			serviceState = route.ServiceStatePending
		case "TERMINATING":
			serviceState = route.ServiceStateTerminating
		default:
			serviceState = route.ServiceStateUnknown
		}

		services[podID] = &route.StatefulServiceState{
			PodID:      podID,
			State:      serviceState,
			UpdateTime: time.Now().Unix(),
			Metadata:   make(map[string]string),
		}
	}

	return services
}

// RoutablePods 获取可路由Pod
func (r *RouteInfoDriverImpl) RoutablePods(namespace, serviceName string) map[int]*route.StatefulServiceState {
	if r.stateCache != nil {
		return r.stateCache.RoutablePods(namespace, serviceName)
	}

	// 从Redis获取并过滤
	alivePods := r.AlivePods(namespace, serviceName)
	routablePods := make(map[int]*route.StatefulServiceState)

	for podIndex, pod := range alivePods {
		if r.IsPodRoutable(namespace, serviceName, podIndex) {
			routablePods[podIndex] = pod
		}
	}

	return routablePods
}

// GetServiceBestPod 获取服务最佳Pod
func (r *RouteInfoDriverImpl) GetServiceBestPod(ctx context.Context, namespace, serviceName string) (int, error) {
	r.log.Log(log.LevelDebug, "Getting service best pod", "namespace", namespace, "service", serviceName)

	// 优先从缓存获取
	if r.stateCache != nil {
		bestPod, err := r.stateCache.GetServiceBestPod(ctx, namespace, serviceName)
		if err == nil && bestPod > 0 {
			return bestPod, nil
		}
	}

	// 从Redis获取并选择最佳Pod
	readyServices, err := r.GetReadyServiceState(ctx, namespace, serviceName)
	if err != nil {
		return 0, err
	}

	if len(readyServices) == 0 {
		return 0, fmt.Errorf("service not ready")
	}

	// 选择负载最低的Pod
	var bestPod int
	var minLoad route.LoadState = 5 // 过载状态作为初始值

	for podIndex, service := range readyServices {
		if service.LoadState < minLoad {
			bestPod = podIndex
			minLoad = service.LoadState
		}
	}

	return bestPod, nil
}

// IsPodAvailable 检查Pod是否可用
func (r *RouteInfoDriverImpl) IsPodAvailable(namespace, serviceName string, podIndex int) bool {
	// 优先从缓存检查
	if r.stateCache != nil {
		return r.stateCache.IsPodAvailable(namespace, serviceName, podIndex)
	}

	// 从Redis检查
	ctx := context.Background()
	podStates, err := r.statefulExecutor.GetServiceState(ctx, namespace, serviceName)
	if err != nil {
		return false
	}

	_, exists := podStates[podIndex]
	return exists
}

// IsWorkloadReady 检查工作负载是否就绪
func (r *RouteInfoDriverImpl) IsWorkloadReady(ctx context.Context, namespace, serviceName string) (bool, error) {
	// 获取工作负载状态
	workloadState, err := r.statefulExecutor.GetWorkloadState(ctx, namespace, serviceName)
	if err != nil {
		return false, err
	}

	return workloadState == "READY", nil
}

// 辅助方法
func (r *RouteInfoDriverImpl) formatServiceTagRedisKey(namespace, tag string) string {
	return fmt.Sprintf("sf:{%s:tag:%s}", namespace, tag)
}
