package driver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/route"
)

// StatefulRouteForServerDriverImpl 有状态服务端驱动实现
type StatefulRouteForServerDriverImpl struct {
	mu sync.RWMutex

	// 配置信息
	baseConfig *route.BaseConfig
	serverInfo *route.ServerInfo

	// 依赖组件
	statefulExecutor route.StatefulExecutor
	routeInfoDriver  route.RouteInfoDriver

	// 状态信息
	loadState    int
	routingState route.RoutingState
	podIndex     int

	// 定时器
	stateTimer    *time.Ticker
	workloadTimer *time.Ticker

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	logger log.Logger
}

// NewStatefulRouteForServerDriverImpl 创建新的服务端驱动实例
func NewStatefulRouteForServerDriverImpl(
	baseConfig *route.BaseConfig,
	statefulExecutor route.StatefulExecutor,
	routeInfoDriver route.RouteInfoDriver,
	serverInfo *route.ServerInfo,
	logger log.Logger,
) *StatefulRouteForServerDriverImpl {
	return &StatefulRouteForServerDriverImpl{
		baseConfig:       baseConfig,
		statefulExecutor: statefulExecutor,
		routeInfoDriver:  routeInfoDriver,
		serverInfo:       serverInfo,
		logger:           logger,
		routingState:     route.RoutingStateReady,
	}
}

// SetLoadState 设置服务器负载状态
func (s *StatefulRouteForServerDriverImpl) SetLoadState(ctx context.Context, loadState int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loadState = loadState
	s.logger.Log(log.LevelInfo, "Set load state", "loadState", loadState)
	return nil
}

// GetLoadState 获取服务器负载状态
func (s *StatefulRouteForServerDriverImpl) GetLoadState(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.loadState, nil
}

// SetRoutingState 设置路由状态
func (s *StatefulRouteForServerDriverImpl) SetRoutingState(ctx context.Context, state route.RoutingState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.routingState = state
	s.logger.Log(log.LevelInfo, "Set routing state", "state", state.String())
	return nil
}

// GetRoutingState 获取路由状态
func (s *StatefulRouteForServerDriverImpl) GetRoutingState(ctx context.Context) (route.RoutingState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.routingState, nil
}

// SetLinkedPod 设置连接信息
func (s *StatefulRouteForServerDriverImpl) SetLinkedPod(
	ctx context.Context,
	namespace, uid, serviceName string,
	podId int,
) (int, error) {
	// 验证命名空间（只支持本地命名空间）
	if namespace != "" && namespace != s.serverInfo.Namespace {
		return 0, fmt.Errorf("namespace %s is not local", namespace)
	}

	// 调用Redis执行器设置链接
	previousPodId, err := s.statefulExecutor.SetLinkedPod(
		ctx, namespace, uid, serviceName, podId, -1, // PERSIST_FOREVER
	)
	if err != nil {
		s.logger.Log(log.LevelError, "Failed to set linked pod", "error", err)
		return 0, err
	}

	s.logger.Log(log.LevelInfo, "Set linked pod", "namespace", namespace, "uid", uid, "service", serviceName, "podId", podId, "previous", previousPodId)

	return previousPodId, nil
}

// TrySetLinkedPod 尝试设置指定uid连接的podId
func (s *StatefulRouteForServerDriverImpl) TrySetLinkedPod(
	ctx context.Context,
	namespace, uid, serviceName string,
	podId int,
) (bool, int, error) {
	// 验证命名空间（只支持本地命名空间）
	if namespace != "" && namespace != s.serverInfo.Namespace {
		return false, 0, fmt.Errorf("namespace %s is not local", namespace)
	}

	// 调用Redis执行器尝试设置链接
	success, previousPodId, err := s.statefulExecutor.TrySetLinkedPod(
		ctx, namespace, uid, serviceName, podId, -1, // PERSIST_FOREVER
	)
	if err != nil {
		s.logger.Log(log.LevelError, "Failed to try set linked pod", "error", err)
		return false, 0, err
	}

	s.logger.Log(log.LevelInfo, "Try set linked pod", "namespace", namespace, "uid", uid, "service", serviceName, "podId", podId, "success", success, "previous", previousPodId)

	return success, previousPodId, nil
}

// RemoveLinkedPod 移除服务连接信息
func (s *StatefulRouteForServerDriverImpl) RemoveLinkedPod(
	ctx context.Context,
	namespace, uid, serviceName string,
) (bool, error) {
	// 验证命名空间（只支持本地命名空间）
	if namespace != "" && namespace != s.serverInfo.Namespace {
		return false, fmt.Errorf("namespace %s is not local", namespace)
	}

	// 调用Redis执行器移除链接
	removed, err := s.statefulExecutor.RemoveLinkedPod(
		ctx, namespace, uid, serviceName, 0, // 立即移除
	)
	if err != nil {
		s.logger.Log(log.LevelError, "Failed to remove linked pod", "error", err)
		return false, err
	}

	s.logger.Log(log.LevelInfo, "Remove linked pod", "namespace", namespace, "uid", uid, "service", serviceName, "removed", removed)

	return removed, nil
}

// RemoveLinkedPodWithId 移除指定uid连接的podId
func (s *StatefulRouteForServerDriverImpl) RemoveLinkedPodWithId(
	ctx context.Context,
	namespace, uid, serviceName string,
	podId, persistSeconds int,
) (bool, error) {
	// 验证命名空间（只支持本地命名空间）
	if namespace != "" && namespace != s.serverInfo.Namespace {
		return false, fmt.Errorf("namespace %s is not local", namespace)
	}

	// 调用Redis执行器移除指定pod的链接
	removed, err := s.statefulExecutor.RemoveLinkedPodWithId(
		ctx, namespace, uid, serviceName, persistSeconds, podId,
	)
	if err != nil {
		s.logger.Log(log.LevelError, "Failed to remove linked pod with id", "error", err)
		return false, err
	}

	s.logger.Log(log.LevelInfo, "Remove linked pod with id", "namespace", namespace, "uid", uid, "service", serviceName, "podId", podId, "removed", removed)

	return removed, nil
}

// Start 启动服务端驱动
func (s *StatefulRouteForServerDriverImpl) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	// 初始化pod信息
	if err := s.initializePodInfo(); err != nil {
		return fmt.Errorf("failed to initialize pod info: %w", err)
	}

	// 启动状态更新定时器
	s.startStateUpdateTimer()

	// 如果是主pod，启动工作负载状态更新
	if s.podIndex == 0 {
		s.startWorkloadStateUpdateTimer()
	}

	s.logger.Log(log.LevelInfo, "Stateful route server driver started", "service", s.serverInfo.ServiceName, "pod", s.podIndex)
	return nil
}

// Stop 停止服务端驱动
func (s *StatefulRouteForServerDriverImpl) Stop(ctx context.Context) error {
	s.logger.Log(log.LevelInfo, "Stopping stateful route server driver...")

	// 设置拒绝新调用状态
	s.SetRoutingState(ctx, route.RoutingStateDenyNewCall)

	// 等待一段时间确保状态传播
	waitTime := time.Duration(s.baseConfig.ServiceStateUpdatePeriodSeconds*5) * time.Second
	select {
	case <-time.After(waitTime):
	case <-ctx.Done():
		return ctx.Err()
	}

	// 取消上下文
	if s.cancel != nil {
		s.cancel()
	}

	// 停止定时器
	if s.stateTimer != nil {
		s.stateTimer.Stop()
	}
	if s.workloadTimer != nil {
		s.workloadTimer.Stop()
	}

	s.logger.Log(log.LevelInfo, "Stateful route server driver stopped")
	return nil
}

// initializePodInfo 初始化pod信息
func (s *StatefulRouteForServerDriverImpl) initializePodInfo() error {
	s.podIndex = s.serverInfo.PodIndex
	return nil
}

// startStateUpdateTimer 启动状态更新定时器
func (s *StatefulRouteForServerDriverImpl) startStateUpdateTimer() {
	interval := time.Duration(s.baseConfig.ServiceStateUpdatePeriodSeconds) * time.Second
	s.stateTimer = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-s.stateTimer.C:
				s.updateServiceState()
			}
		}
	}()
}

// startWorkloadStateUpdateTimer 启动工作负载状态更新定时器
func (s *StatefulRouteForServerDriverImpl) startWorkloadStateUpdateTimer() {
	// 工作负载状态更新间隔
	interval := time.Duration(s.baseConfig.ServiceStateUpdatePeriodSeconds*2) * time.Second
	s.workloadTimer = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-s.workloadTimer.C:
				s.updateWorkloadState()
			}
		}
	}()
}

// updateServiceState 更新服务状态
func (s *StatefulRouteForServerDriverImpl) updateServiceState() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// 计算当前状态
	currentState := s.calculateCurrentState()

	// 设置服务状态
	err := s.statefulExecutor.SetServiceState(
		ctx, s.serverInfo.Namespace, s.serverInfo.ServiceName, s.podIndex, currentState,
	)
	if err != nil {
		s.logger.Log(log.LevelError, "Failed to update service state", "error", err)
		return
	}

	s.logger.Log(log.LevelDebug, "Updated service state", "state", currentState)
}

// updateWorkloadState 更新工作负载状态
func (s *StatefulRouteForServerDriverImpl) updateWorkloadState() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// 计算工作负载状态
	workloadState := s.calculateWorkloadState()

	// 设置工作负载状态
	err := s.statefulExecutor.SetWorkloadState(
		ctx, s.serverInfo.Namespace, s.serverInfo.ServiceName, workloadState,
	)
	if err != nil {
		s.logger.Log(log.LevelError, "Failed to update workload state", "error", err)
		return
	}

	s.logger.Log(log.LevelDebug, "Updated workload state", "state", workloadState)
}

// calculateCurrentState 计算当前状态
func (s *StatefulRouteForServerDriverImpl) calculateCurrentState() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 计算加权状态
	cpuFactor := s.baseConfig.CpuFactor
	weightedState := int(cpuFactor*float64(s.getWaitingTasksNum()) + (1-cpuFactor)*float64(s.loadState))

	// 创建状态对象
	stateObj := &route.StateObj{
		State:        weightedState,
		UpdateTime:   time.Now(),
		RoutingState: s.routingState.String(),
	}

	return stateObj.ToStateString()
}

// calculateWorkloadState 计算工作负载状态
func (s *StatefulRouteForServerDriverImpl) calculateWorkloadState() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 根据负载状态和路由状态计算工作负载状态
	if s.routingState == route.RoutingStateShutdown {
		return "Shutdown"
	}

	if s.loadState >= 5 {
		return "Overloaded"
	}

	if s.loadState >= 3 {
		return "High"
	}

	if s.loadState >= 1 {
		return "Normal"
	}

	return "Idle"
}

// getWaitingTasksNum 获取等待任务数量（简化实现）
func (s *StatefulRouteForServerDriverImpl) getWaitingTasksNum() int {
	// 这里可以根据实际情况实现
	// 暂时返回负载状态的一半作为等待任务数
	return s.loadState / 2
}
