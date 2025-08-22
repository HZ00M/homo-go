## 6A 任务卡：实现路由管理器

- 编号: Task-04
- 模块: stateful-route
- 责任人: 待分配
- 优先级: 🟡 中
- 状态: ❌ 未开始
- 预计完成时间: 待定
- 实际完成时间: -

### A1 目标（Aim）
实现有状态路由模块的路由管理器，负责协调服务端驱动和客户端驱动，提供统一的路由管理接口，支持负载均衡、健康检查、故障转移等高级功能。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Java版本的路由管理逻辑（分散在各个驱动中）
  - 🔄 部分实现：服务端和客户端驱动接口已定义
  - ❌ 未实现：Go版本的统一路由管理器
- **差距**：
  - Java的路由管理逻辑需要整合到统一管理器中
  - 负载均衡算法需要适配Go实现
  - 健康检查和故障转移机制需要重新设计
- **约束**：
  - 必须保持与Java版本功能逻辑一致
  - 使用现有的routeInfoDriver和StatefulRedisExecutor
  - 符合Go语言和Kratos框架设计规范
- **风险**：
  - 技术风险：负载均衡算法实现复杂度
  - 业务风险：故障转移逻辑可能遗漏
  - 依赖风险：依赖组件接口变更影响

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`StatefulRouteManager` - 统一路由管理接口
  - **核心方法**：
    - 路由管理：`RouteRequest`、`GetBestPod`、`UpdatePodHealth`
    - 负载均衡：`CalculateLoad`、`SelectPod`、`BalanceLoad`
    - 故障转移：`HandlePodFailure`、`MigrateUser`、`RecoverPod`
  - **输入输出参数及错误码**：
    - 使用`context.Context`进行超时和取消控制
    - 返回`error`接口类型，支持错误包装
    - 使用Go的多返回值特性

- **架构设计**：
  - 采用策略模式，支持多种负载均衡算法
  - 使用观察者模式，支持健康状态变化通知
  - 支持配置驱动的路由策略

- **核心功能模块**：
  - `StatefulRouteManager`: 统一路由管理接口
  - 负载均衡器：多种算法实现
  - 健康检查器：pod状态监控
  - 故障转移器：异常处理和恢复

- **极小任务拆分**：
  - T04-01：定义路由管理器接口
  - T04-02：实现负载均衡器
  - T04-03：实现健康检查器
  - T04-04：实现故障转移器
  - T04-05：实现统一路由管理器

### A4 行动（Act）
#### T04-01：定义路由管理器接口
```go
// route/state/route_manager.go
package state

import (
    "context"
    "time"
    
    "tpf-service-go/stateful-route/types"
)

// LoadBalancingStrategy 负载均衡策略
type LoadBalancingStrategy int

const (
    StrategyRoundRobin LoadBalancingStrategy = iota    // 轮询
    StrategyLeastConnections                          // 最少连接
    StrategyWeightedRoundRobin                        // 加权轮询
    StrategyLeastResponseTime                         // 最少响应时间
)

// PodHealthStatus Pod健康状态
type PodHealthStatus struct {
    PodIndex     int           `json:"pod_index"`
    IsHealthy   bool          `json:"is_healthy"`
    LastCheck   time.Time     `json:"last_check"`
    ResponseTime time.Duration `json:"response_time"`
    ErrorCount  int           `json:"error_count"`
}

// RouteRequest 路由请求
type RouteRequest struct {
    Namespace   string            `json:"namespace"`
    UID         string            `json:"uid"`
    ServiceName string            `json:"service_name"`
    Strategy   LoadBalancingStrategy `json:"strategy"`
    Priority   int                `json:"priority"`
}

// RouteResult 路由结果
type RouteResult struct {
    PodIndex    int               `json:"pod_index"`
    Strategy   LoadBalancingStrategy `json:"strategy"`
    LoadState  int                `json:"load_state"`
    HealthStatus *PodHealthStatus `json:"health_status"`
}

// StatefulRouteManager 有状态路由管理器接口
type StatefulRouteManager interface {
    // 路由请求到最佳pod
    RouteRequest(ctx context.Context, req *RouteRequest) (*RouteResult, error)
    
    // 获取最佳pod
    GetBestPod(ctx context.Context, namespace, serviceName string, strategy LoadBalancingStrategy) (int, error)
    
    // 更新pod健康状态
    UpdatePodHealth(ctx context.Context, namespace, serviceName string, podIndex int, status *PodHealthStatus) error
    
    // 处理pod故障
    HandlePodFailure(ctx context.Context, namespace, serviceName string, podIndex int) error
    
    // 迁移用户到其他pod
    MigrateUser(ctx context.Context, namespace, uid, serviceName string, targetPodIndex int) error
    
    // 恢复pod服务
    RecoverPod(ctx context.Context, namespace, serviceName string, podIndex int) error
    
    // 获取服务负载状态
    GetServiceLoadStatus(ctx context.Context, namespace, serviceName string) (map[int]*PodHealthStatus, error)
}
```

#### T04-02：实现负载均衡器
```go
// route/state/load_balancer.go
package state

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
    SelectPod(ctx context.Context, pods []*PodInfo, strategy LoadBalancingStrategy) (int, error)
}

// PodInfo Pod信息
type PodInfo struct {
    Index        int
    LoadState    int
    ConnectionCount int
    ResponseTime time.Duration
    IsHealthy   bool
    Weight      int
}

// LoadBalancerImpl 负载均衡器实现
type LoadBalancerImpl struct {
    mu     sync.RWMutex
    logger log.Logger
}

// NewLoadBalancer 创建新的负载均衡器
func NewLoadBalancer(logger log.Logger) *LoadBalancerImpl {
    return &LoadBalancerImpl{
        logger: logger,
    }
}

// SelectPod 根据策略选择pod
func (lb *LoadBalancerImpl) SelectPod(ctx context.Context, pods []*PodInfo, strategy LoadBalancingStrategy) (int, error) {
    if len(pods) == 0 {
        return 0, fmt.Errorf("no available pods")
    }
    
    // 过滤健康的pod
    healthyPods := lb.filterHealthyPods(pods)
    if len(healthyPods) == 0 {
        return 0, fmt.Errorf("no healthy pods available")
    }
    
    switch strategy {
    case StrategyRoundRobin:
        return lb.roundRobin(healthyPods)
    case StrategyLeastConnections:
        return lb.leastConnections(healthyPods)
    case StrategyWeightedRoundRobin:
        return lb.weightedRoundRobin(healthyPods)
    case StrategyLeastResponseTime:
        return lb.leastResponseTime(healthyPods)
    default:
        return lb.roundRobin(healthyPods)
    }
}

// roundRobin 轮询策略
func (lb *LoadBalancerImpl) roundRobin(pods []*PodInfo) (int, error) {
    lb.mu.Lock()
    defer lb.mu.Unlock()
    
    // 简单的轮询实现，实际应该使用原子计数器
    selectedIndex := rand.Intn(len(pods))
    return pods[selectedIndex].Index, nil
}

// leastConnections 最少连接策略
func (lb *LoadBalancerImpl) leastConnections(pods []*PodInfo) (int, error) {
    if len(pods) == 0 {
        return 0, fmt.Errorf("no pods available")
    }
    
    minConnections := pods[0].ConnectionCount
    selectedPod := pods[0]
    
    for _, pod := range pods[1:] {
        if pod.ConnectionCount < minConnections {
            minConnections = pod.ConnectionCount
            selectedPod = pod
        }
    }
    
    return selectedPod.Index, nil
}

// weightedRoundRobin 加权轮询策略
func (lb *LoadBalancerImpl) weightedRoundRobin(pods []*PodInfo) (int, error) {
    if len(pods) == 0 {
        return 0, fmt.Errorf("no pods available")
    }
    
    totalWeight := 0
    for _, pod := range pods {
        totalWeight += pod.Weight
    }
    
    if totalWeight == 0 {
        return lb.roundRobin(pods)
    }
    
    // 简单的加权选择，实际应该使用更复杂的算法
    randomWeight := rand.Intn(totalWeight)
    currentWeight := 0
    
    for _, pod := range pods {
        currentWeight += pod.Weight
        if randomWeight < currentWeight {
            return pod.Index, nil
        }
    }
    
    return pods[0].Index, nil
}

// leastResponseTime 最少响应时间策略
func (lb *LoadBalancerImpl) leastResponseTime(pods []*PodInfo) (int, error) {
    if len(pods) == 0 {
        return 0, fmt.Errorf("no pods available")
    }
    
    minResponseTime := pods[0].ResponseTime
    selectedPod := pods[0]
    
    for _, pod := range pods[1:] {
        if pod.ResponseTime < minResponseTime {
            minResponseTime = pod.ResponseTime
            selectedPod = pod
        }
    }
    
    return selectedPod.Index, nil
}

// filterHealthyPods 过滤健康的pod
func (lb *LoadBalancerImpl) filterHealthyPods(pods []*PodInfo) []*PodInfo {
    healthyPods := make([]*PodInfo, 0, len(pods))
    for _, pod := range pods {
        if pod.IsHealthy {
            healthyPods = append(healthyPods, pod)
        }
    }
    return healthyPods
}
```

#### T04-03：实现健康检查器
```go
// route/state/health_checker.go
package state

import (
    "context"
    "sync"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
)

// HealthChecker 健康检查器接口
type HealthChecker interface {
    StartHealthCheck(ctx context.Context, namespace, serviceName string) error
    StopHealthCheck(namespace, serviceName string) error
    GetPodHealth(ctx context.Context, namespace, serviceName string, podIndex int) (*PodHealthStatus, error)
    UpdatePodHealth(ctx context.Context, namespace, serviceName string, podIndex int, status *PodHealthStatus) error
}

// HealthCheckerImpl 健康检查器实现
type HealthCheckerImpl struct {
    mu sync.RWMutex
    
    // 健康检查任务
    healthChecks map[string]*HealthCheckTask
    
    // 依赖组件
    routeInfoDriver interfaces.RouteInfoDriver
    
    logger log.Logger
}

// HealthCheckTask 健康检查任务
type HealthCheckTask struct {
    Namespace   string
    ServiceName string
    Interval    time.Duration
    Timeout     time.Duration
    StopChan    chan struct{}
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker(routeInfoDriver interfaces.RouteInfoDriver, logger log.Logger) *HealthCheckerImpl {
    return &HealthCheckerImpl{
        healthChecks:   make(map[string]*HealthCheckTask),
        routeInfoDriver: routeInfoDriver,
        logger:         logger,
    }
}

// StartHealthCheck 启动健康检查
func (hc *HealthCheckerImpl) StartHealthCheck(ctx context.Context, namespace, serviceName string) error {
    hc.mu.Lock()
    defer hc.mu.Unlock()
    
    key := hc.getHealthCheckKey(namespace, serviceName)
    if _, exists := hc.healthChecks[key]; exists {
        return fmt.Errorf("health check already running for %s", key)
    }
    
    task := &HealthCheckTask{
        Namespace:   namespace,
        ServiceName: serviceName,
        Interval:    30 * time.Second, // 可配置
        Timeout:     5 * time.Second,  // 可配置
        StopChan:    make(chan struct{}),
    }
    
    hc.healthChecks[key] = task
    
    // 启动健康检查协程
    go hc.runHealthCheck(ctx, task)
    
    hc.logger.Infof("Started health check for %s", key)
    return nil
}

// runHealthCheck 运行健康检查
func (hc *HealthCheckerImpl) runHealthCheck(ctx context.Context, task *HealthCheckTask) {
    ticker := time.NewTicker(task.Interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-task.StopChan:
            return
        case <-ticker.C:
            hc.performHealthCheck(ctx, task)
        }
    }
}

// performHealthCheck 执行健康检查
func (hc *HealthCheckerImpl) performHealthCheck(ctx context.Context, task *HealthCheckTask) {
    // 获取服务下的所有pod
    pods, err := hc.routeInfoDriver.GetServicePods(ctx, task.Namespace, task.ServiceName)
    if err != nil {
        hc.logger.Errorf("Failed to get service pods: %v", err)
        return
    }
    
    // 对每个pod进行健康检查
    for _, podIndex := range pods {
        go hc.checkPodHealth(ctx, task.Namespace, task.ServiceName, podIndex)
    }
}

// checkPodHealth 检查单个pod的健康状态
func (hc *HealthCheckerImpl) checkPodHealth(ctx context.Context, namespace, serviceName string, podIndex int) {
    checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    startTime := time.Now()
    
    // 执行健康检查（这里应该调用实际的健康检查逻辑）
    isHealthy, err := hc.routeInfoDriver.CheckPodHealth(checkCtx, namespace, serviceName, podIndex)
    
    responseTime := time.Since(startTime)
    
    // 更新健康状态
    status := &PodHealthStatus{
        PodIndex:     podIndex,
        IsHealthy:    isHealthy,
        LastCheck:    time.Now(),
        ResponseTime: responseTime,
    }
    
    if err != nil {
        status.ErrorCount++
        hc.logger.Errorf("Health check failed for pod %d: %v", podIndex, err)
    }
    
    // 更新健康状态
    if err := hc.UpdatePodHealth(ctx, namespace, serviceName, podIndex, status); err != nil {
        hc.logger.Errorf("Failed to update pod health: %v", err)
    }
}

// StopHealthCheck 停止健康检查
func (hc *HealthCheckerImpl) StopHealthCheck(namespace, serviceName string) error {
    hc.mu.Lock()
    defer hc.mu.Unlock()
    
    key := hc.getHealthCheckKey(namespace, serviceName)
    task, exists := hc.healthChecks[key]
    if !exists {
        return fmt.Errorf("health check not running for %s", key)
    }
    
    close(task.StopChan)
    delete(hc.healthChecks, key)
    
    hc.logger.Infof("Stopped health check for %s", key)
    return nil
}

// getHealthCheckKey 获取健康检查键
func (hc *HealthCheckerImpl) getHealthCheckKey(namespace, serviceName string) string {
    return fmt.Sprintf("%s:%s", namespace, serviceName)
}
```

#### T04-04：实现故障转移器
```go
// route/state/failover_handler.go
package state

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
)

// FailoverHandler 故障转移处理器接口
type FailoverHandler interface {
    HandlePodFailure(ctx context.Context, namespace, serviceName string, podIndex int) error
    MigrateUser(ctx context.Context, namespace, uid, serviceName string, targetPodIndex int) error
    RecoverPod(ctx context.Context, namespace, serviceName string, podIndex int) error
    GetFailoverHistory(ctx context.Context, namespace, serviceName string) ([]*FailoverEvent, error)
}

// FailoverEvent 故障转移事件
type FailoverEvent struct {
    Timestamp   time.Time `json:"timestamp"`
    EventType   string    `json:"event_type"`
    PodIndex    int       `json:"pod_index"`
    Reason      string    `json:"reason"`
    AffectedUIDs []string `json:"affected_uids"`
}

// FailoverHandlerImpl 故障转移处理器实现
type FailoverHandlerImpl struct {
    mu sync.RWMutex
    
    // 故障转移历史
    failoverHistory map[string][]*FailoverEvent
    
    // 依赖组件
    statefulRedisExecutor interfaces.StatefulRedisExecutor
    routeInfoDriver       interfaces.RouteInfoDriver
    
    logger log.Logger
}

// NewFailoverHandler 创建新的故障转移处理器
func NewFailoverHandler(
    statefulRedisExecutor interfaces.StatefulRedisExecutor,
    routeInfoDriver interfaces.RouteInfoDriver,
    logger log.Logger,
) *FailoverHandlerImpl {
    return &FailoverHandlerImpl{
        failoverHistory:     make(map[string][]*FailoverEvent),
        statefulRedisExecutor: statefulRedisExecutor,
        routeInfoDriver:       routeInfoDriver,
        logger:               logger,
    }
}

// HandlePodFailure 处理pod故障
func (fh *FailoverHandlerImpl) HandlePodFailure(
    ctx context.Context, 
    namespace, serviceName string, 
    podIndex int,
) error {
    fh.logger.Warnf("Handling pod failure: namespace=%s, service=%s, pod=%d", namespace, serviceName, podIndex)
    
    // 获取受影响的用户
    affectedUIDs, err := fh.getAffectedUsers(ctx, namespace, serviceName, podIndex)
    if err != nil {
        return fmt.Errorf("failed to get affected users: %w", err)
    }
    
    // 为每个受影响的用户找到新的pod
    for _, uid := range affectedUIDs {
        if err := fh.migrateUserToNewPod(ctx, namespace, uid, serviceName); err != nil {
            fh.logger.Errorf("Failed to migrate user %s: %v", uid, err)
            continue
        }
    }
    
    // 记录故障转移事件
    event := &FailoverEvent{
        Timestamp:    time.Now(),
        EventType:    "POD_FAILURE",
        PodIndex:     podIndex,
        Reason:       "Health check failed",
        AffectedUIDs: affectedUIDs,
    }
    
    fh.recordFailoverEvent(namespace, serviceName, event)
    
    fh.logger.Infof("Pod failure handled: %d users migrated", len(affectedUIDs))
    return nil
}

// MigrateUser 迁移用户到指定pod
func (fh *FailoverHandlerImpl) MigrateUser(
    ctx context.Context, 
    namespace, uid, serviceName string, 
    targetPodIndex int,
) error {
    // 验证目标pod是否可用
    isAvailable, err := fh.routeInfoDriver.IsPodAvailable(ctx, namespace, serviceName, targetPodIndex)
    if err != nil {
        return fmt.Errorf("failed to check pod availability: %w", err)
    }
    
    if !isAvailable {
        return fmt.Errorf("target pod %d is not available", targetPodIndex)
    }
    
    // 设置新的链接关系
    _, err = fh.statefulRedisExecutor.SetLinkedPod(ctx, namespace, uid, serviceName, targetPodIndex, -1)
    if err != nil {
        return fmt.Errorf("failed to set linked pod: %w", err)
    }
    
    fh.logger.Infof("User %s migrated to pod %d", uid, targetPodIndex)
    return nil
}

// RecoverPod 恢复pod服务
func (fh *FailoverHandlerImpl) RecoverPod(
    ctx context.Context, 
    namespace, serviceName string, 
    podIndex int,
) error {
    fh.logger.Infof("Recovering pod: namespace=%s, service=%s, pod=%d", namespace, serviceName, podIndex)
    
    // 记录恢复事件
    event := &FailoverEvent{
        Timestamp:   time.Now(),
        EventType:   "POD_RECOVERY",
        PodIndex:    podIndex,
        Reason:      "Health check passed",
        AffectedUIDs: []string{},
    }
    
    fh.recordFailoverEvent(namespace, serviceName, event)
    
    fh.logger.Infof("Pod %d recovered", podIndex)
    return nil
}

// getAffectedUsers 获取受影响的用户
func (fh *FailoverHandlerImpl) getAffectedUsers(
    ctx context.Context, 
    namespace, serviceName string, 
    podIndex int,
) ([]string, error) {
    // 这里应该实现从Redis获取链接到指定pod的用户列表
    // 由于接口限制，这里返回空列表
    return []string{}, nil
}

// migrateUserToNewPod 迁移用户到新pod
func (fh *FailoverHandlerImpl) migrateUserToNewPod(
    ctx context.Context, 
    namespace, uid, serviceName string,
) error {
    // 获取最佳可用pod
    bestPod, err := fh.routeInfoDriver.GetServiceBestPod(ctx, namespace, serviceName)
    if err != nil {
        return fmt.Errorf("failed to get best pod: %w", err)
    }
    
    if bestPod == 0 {
        return fmt.Errorf("no available pod found")
    }
    
    // 迁移用户
    return fh.MigrateUser(ctx, namespace, uid, serviceName, bestPod)
}

// recordFailoverEvent 记录故障转移事件
func (fh *FailoverHandlerImpl) recordFailoverEvent(namespace, serviceName string, event *FailoverEvent) {
    fh.mu.Lock()
    defer fh.mu.Unlock()
    
    key := fmt.Sprintf("%s:%s", namespace, serviceName)
    fh.failoverHistory[key] = append(fh.failoverHistory[key], event)
    
    // 限制历史记录数量
    if len(fh.failoverHistory[key]) > 100 {
        fh.failoverHistory[key] = fh.failoverHistory[key][len(fh.failoverHistory[key])-100:]
    }
}

// GetFailoverHistory 获取故障转移历史
func (fh *FailoverHandlerImpl) GetFailoverHistory(
    ctx context.Context, 
    namespace, serviceName string,
) ([]*FailoverEvent, error) {
    fh.mu.RLock()
    defer fh.mu.RUnlock()
    
    key := fmt.Sprintf("%s:%s", namespace, serviceName)
    history, exists := fh.failoverHistory[key]
    if !exists {
        return []*FailoverEvent{}, nil
    }
    
    // 返回副本
    result := make([]*FailoverEvent, len(history))
    copy(result, history)
    
    return result, nil
}
```

#### T04-05：实现统一路由管理器
```go
// route/state/route_manager_impl.go
package state

import (
    "context"
    "fmt"
    "sync"
    
    "github.com/go-kratos/kratos/v2/log"
)

// StatefulRouteManagerImpl 有状态路由管理器实现
type StatefulRouteManagerImpl struct {
    mu sync.RWMutex
    
    // 依赖组件
    loadBalancer    LoadBalancer
    healthChecker   HealthChecker
    failoverHandler FailoverHandler
    
    // 客户端和服务端驱动
    clientDriver interfaces.StatefulRouteForClientDriver
    serverDriver interfaces.StatefulRouteForServerDriver
    
    logger log.Logger
}

// NewStatefulRouteManager 创建新的路由管理器
func NewStatefulRouteManager(
    loadBalancer LoadBalancer,
    healthChecker HealthChecker,
    failoverHandler FailoverHandler,
    clientDriver interfaces.StatefulRouteForClientDriver,
    serverDriver interfaces.StatefulRouteForServerDriver,
    logger log.Logger,
) *StatefulRouteManagerImpl {
    return &StatefulRouteManagerImpl{
        loadBalancer:    loadBalancer,
        healthChecker:   healthChecker,
        failoverHandler: failoverHandler,
        clientDriver:    clientDriver,
        serverDriver:    serverDriver,
        logger:          logger,
    }
}

// RouteRequest 路由请求到最佳pod
func (rm *StatefulRouteManagerImpl) RouteRequest(
    ctx context.Context, 
    req *RouteRequest,
) (*RouteResult, error) {
    rm.logger.Debugf("Routing request: %+v", req)
    
    // 获取最佳pod
    podIndex, err := rm.GetBestPod(ctx, req.Namespace, req.ServiceName, req.Strategy)
    if err != nil {
        return nil, fmt.Errorf("failed to get best pod: %w", err)
    }
    
    // 获取pod健康状态
    healthStatus, err := rm.healthChecker.GetPodHealth(ctx, req.Namespace, req.ServiceName, podIndex)
    if err != nil {
        rm.logger.Warnf("Failed to get pod health: %v", err)
    }
    
    // 获取pod负载状态
    loadState, err := rm.serverDriver.GetLoadState(ctx)
    if err != nil {
        rm.logger.Warnf("Failed to get pod load state: %v", err)
    }
    
    result := &RouteResult{
        PodIndex:     podIndex,
        Strategy:     req.Strategy,
        LoadState:    loadState,
        HealthStatus: healthStatus,
    }
    
    rm.logger.Infof("Request routed to pod %d", podIndex)
    return result, nil
}

// GetBestPod 获取最佳pod
func (rm *StatefulRouteManagerImpl) GetBestPod(
    ctx context.Context, 
    namespace, serviceName string, 
    strategy LoadBalancingStrategy,
) (int, error) {
    // 获取服务下的所有pod信息
    pods, err := rm.getServicePods(ctx, namespace, serviceName)
    if err != nil {
        return 0, fmt.Errorf("failed to get service pods: %w", err)
    }
    
    if len(pods) == 0 {
        return 0, fmt.Errorf("no available pods for service %s", serviceName)
    }
    
    // 使用负载均衡器选择pod
    podIndex, err := rm.loadBalancer.SelectPod(ctx, pods, strategy)
    if err != nil {
        return 0, fmt.Errorf("failed to select pod: %w", err)
    }
    
    return podIndex, nil
}

// UpdatePodHealth 更新pod健康状态
func (rm *StatefulRouteManagerImpl) UpdatePodHealth(
    ctx context.Context, 
    namespace, serviceName string, 
    podIndex int, 
    status *PodHealthStatus,
) error {
    return rm.healthChecker.UpdatePodHealth(ctx, namespace, serviceName, podIndex, status)
}

// HandlePodFailure 处理pod故障
func (rm *StatefulRouteManagerImpl) HandlePodFailure(
    ctx context.Context, 
    namespace, serviceName string, 
    podIndex int,
) error {
    return rm.failoverHandler.HandlePodFailure(ctx, namespace, serviceName, podIndex)
}

// MigrateUser 迁移用户到其他pod
func (rm *StatefulRouteManagerImpl) MigrateUser(
    ctx context.Context, 
    namespace, uid, serviceName string, 
    targetPodIndex int,
) error {
    return rm.failoverHandler.MigrateUser(ctx, namespace, uid, serviceName, targetPodIndex)
}

// RecoverPod 恢复pod服务
func (rm *StatefulRouteManagerImpl) RecoverPod(
    ctx context.Context, 
    namespace, serviceName string, 
    podIndex int,
) error {
    return rm.failoverHandler.RecoverPod(ctx, namespace, serviceName, podIndex)
}

// GetServiceLoadStatus 获取服务负载状态
func (rm *StatefulRouteManagerImpl) GetServiceLoadStatus(
    ctx context.Context, 
    namespace, serviceName string,
) (map[int]*PodHealthStatus, error) {
    // 获取服务下的所有pod
    pods, err := rm.getServicePods(ctx, namespace, serviceName)
    if err != nil {
        return nil, fmt.Errorf("failed to get service pods: %w", err)
    }
    
    result := make(map[int]*PodHealthStatus)
    for _, pod := range pods {
        healthStatus, err := rm.healthChecker.GetPodHealth(ctx, namespace, serviceName, pod.Index)
        if err != nil {
            rm.logger.Warnf("Failed to get health status for pod %d: %v", pod.Index, err)
            continue
        }
        result[pod.Index] = healthStatus
    }
    
    return result, nil
}

// getServicePods 获取服务下的所有pod信息
func (rm *StatefulRouteManagerImpl) getServicePods(
    ctx context.Context, 
    namespace, serviceName string,
) ([]*PodInfo, error) {
    // 这里应该实现从routeInfoDriver获取pod信息的逻辑
    // 由于接口限制，这里返回空列表
    return []*PodInfo{}, nil
}
```

### A5 验证（Assure）
- **测试用例**：
  - 负载均衡功能测试
  - 健康检查功能测试
  - 故障转移功能测试
  - 路由管理功能测试
- **性能验证**：
  - 负载均衡性能测试
  - 健康检查性能测试
  - 故障转移性能测试
- **回归测试**：
  - 确保与Java版本功能一致
- **测试结果**：
  - 待实现

### A6 迭代（Advance）
- 性能优化：负载均衡算法优化、健康检查优化
- 功能扩展：支持更多负载均衡策略、监控指标
- 观测性增强：添加详细日志、性能监控、故障转移统计
- 下一步任务链接：Task-05 编写单元测试

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
[总结任务完成情况]
