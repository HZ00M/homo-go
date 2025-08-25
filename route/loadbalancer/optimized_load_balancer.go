package loadbalancer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// LoadBalancingStrategy 负载均衡策略
type LoadBalancingStrategy int

const (
	StrategyRoundRobin LoadBalancingStrategy = iota
	StrategyLeastConnections
	StrategyWeightedRoundRobin
	StrategyLeastResponseTime
)

// PodInfo Pod信息
type PodInfo struct {
	Index        int
	Weight       int
	IsHealthy    bool
	ResponseTime time.Duration
}

// OptimizedLoadBalancer 优化的负载均衡器
type OptimizedLoadBalancer struct {
	mu sync.RWMutex

	// 轮询计数器
	roundRobinCounter uint64

	// 加权轮询状态
	weightedState map[int]*WeightedState

	// 最少连接状态
	connectionState map[int]*ConnectionState

	// 性能统计
	stats *LoadBalancerStats

	logger log.Logger
}

// WeightedState 加权状态
type WeightedState struct {
	CurrentWeight   int
	EffectiveWeight int
	PodIndex        int
}

// ConnectionState 连接状态
type ConnectionState struct {
	ConnectionCount int64
	LastUpdate      time.Time
	PodIndex        int
}

// LoadBalancerStats 负载均衡器统计
type LoadBalancerStats struct {
	TotalRequests    int64
	StrategyRequests map[string]int64
	AverageLatency   time.Duration
	LastUpdate       time.Time
}

// NewOptimizedLoadBalancer 创建新的优化负载均衡器
func NewOptimizedLoadBalancer(logger log.Logger) *OptimizedLoadBalancer {
	return &OptimizedLoadBalancer{
		weightedState:   make(map[int]*WeightedState),
		connectionState: make(map[int]*ConnectionState),
		stats: &LoadBalancerStats{
			StrategyRequests: make(map[string]int64),
		},
		logger: logger,
	}
}

// SelectPod 选择pod（优化版本）
func (olb *OptimizedLoadBalancer) SelectPod(ctx context.Context, pods []*PodInfo, strategy LoadBalancingStrategy) (int, error) {
	if len(pods) == 0 {
		return 0, fmt.Errorf("no available pods")
	}

	// 过滤健康的pod
	healthyPods := olb.filterHealthyPods(pods)
	if len(healthyPods) == 0 {
		return 0, fmt.Errorf("no healthy pods available")
	}

	startTime := time.Now()
	var podIndex int
	var err error

	// 根据策略选择pod
	switch strategy {
	case StrategyRoundRobin:
		podIndex, err = olb.optimizedRoundRobin(healthyPods)
	case StrategyLeastConnections:
		podIndex, err = olb.optimizedLeastConnections(healthyPods)
	case StrategyWeightedRoundRobin:
		podIndex, err = olb.optimizedWeightedRoundRobin(healthyPods)
	case StrategyLeastResponseTime:
		podIndex, err = olb.optimizedLeastResponseTime(healthyPods)
	default:
		podIndex, err = olb.optimizedRoundRobin(healthyPods)
	}

	// 更新统计信息
	olb.updateStats(strategy, time.Since(startTime))

	return podIndex, err
}

// optimizedRoundRobin 优化的轮询策略
func (olb *OptimizedLoadBalancer) optimizedRoundRobin(pods []*PodInfo) (int, error) {
	// 使用原子操作进行轮询
	counter := atomic.AddUint64(&olb.roundRobinCounter, 1)
	selectedIndex := int(counter % uint64(len(pods)))

	return pods[selectedIndex].Index, nil
}

// optimizedLeastConnections 优化的最少连接策略
func (olb *OptimizedLoadBalancer) optimizedLeastConnections(pods []*PodInfo) (int, error) {
	olb.mu.RLock()
	defer olb.mu.RUnlock()

	var minConnections int64 = 1<<63 - 1
	var selectedPod *PodInfo

	for _, pod := range pods {
		// 获取连接状态
		state, exists := olb.connectionState[pod.Index]
		if !exists {
			// 初始化连接状态
			olb.connectionState[pod.Index] = &ConnectionState{
				ConnectionCount: 0,
				LastUpdate:      time.Now(),
				PodIndex:        pod.Index,
			}
			state = olb.connectionState[pod.Index]
		}

		if state.ConnectionCount < minConnections {
			minConnections = state.ConnectionCount
			selectedPod = pod
		}
	}

	if selectedPod == nil {
		return 0, fmt.Errorf("no pod selected")
	}

	return selectedPod.Index, nil
}

// optimizedWeightedRoundRobin 优化的加权轮询策略
func (olb *OptimizedLoadBalancer) optimizedWeightedRoundRobin(pods []*PodInfo) (int, error) {
	olb.mu.Lock()
	defer olb.mu.Unlock()

	// 初始化加权状态
	for _, pod := range pods {
		if _, exists := olb.weightedState[pod.Index]; !exists {
			olb.weightedState[pod.Index] = &WeightedState{
				CurrentWeight:   pod.Weight,
				EffectiveWeight: pod.Weight,
				PodIndex:        pod.Index,
			}
		}
	}

	// 选择权重最高的pod
	var maxWeight int
	var selectedPod *PodInfo

	for _, pod := range pods {
		state := olb.weightedState[pod.Index]
		if state.CurrentWeight > maxWeight {
			maxWeight = state.CurrentWeight
			selectedPod = pod
		}
	}

	if selectedPod == nil {
		return 0, fmt.Errorf("no pod selected")
	}

	// 更新权重
	state := olb.weightedState[selectedPod.Index]
	state.CurrentWeight -= olb.calculateTotalWeight(pods)

	// 如果权重为负，重置为有效权重
	if state.CurrentWeight <= 0 {
		state.CurrentWeight += state.EffectiveWeight
	}

	return selectedPod.Index, nil
}

// optimizedLeastResponseTime 优化的最少响应时间策略
func (olb *OptimizedLoadBalancer) optimizedLeastResponseTime(pods []*PodInfo) (int, error) {
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
func (olb *OptimizedLoadBalancer) filterHealthyPods(pods []*PodInfo) []*PodInfo {
	healthyPods := make([]*PodInfo, 0, len(pods))
	for _, pod := range pods {
		if pod.IsHealthy {
			healthyPods = append(healthyPods, pod)
		}
	}
	return healthyPods
}

// calculateTotalWeight 计算总权重
func (olb *OptimizedLoadBalancer) calculateTotalWeight(pods []*PodInfo) int {
	total := 0
	for _, pod := range pods {
		total += pod.Weight
	}
	return total
}

// updateStats 更新统计信息
func (olb *OptimizedLoadBalancer) updateStats(strategy LoadBalancingStrategy, latency time.Duration) {
	olb.mu.Lock()
	defer olb.mu.Unlock()

	atomic.AddInt64(&olb.stats.TotalRequests, 1)

	strategyName := strategy.String()
	olb.stats.StrategyRequests[strategyName]++

	// 更新平均延迟
	if olb.stats.AverageLatency == 0 {
		olb.stats.AverageLatency = latency
	} else {
		olb.stats.AverageLatency = (olb.stats.AverageLatency + latency) / 2
	}

	olb.stats.LastUpdate = time.Now()
}

// GetStats 获取统计信息
func (olb *OptimizedLoadBalancer) GetStats(ctx context.Context) *LoadBalancerStats {
	olb.mu.RLock()
	defer olb.mu.RUnlock()

	// 返回副本
	stats := &LoadBalancerStats{
		TotalRequests:    atomic.LoadInt64(&olb.stats.TotalRequests),
		StrategyRequests: make(map[string]int64),
		AverageLatency:   olb.stats.AverageLatency,
		LastUpdate:       olb.stats.LastUpdate,
	}

	for k, v := range olb.stats.StrategyRequests {
		stats.StrategyRequests[k] = v
	}

	return stats
}

// String 策略字符串表示
func (s LoadBalancingStrategy) String() string {
	switch s {
	case StrategyRoundRobin:
		return "RoundRobin"
	case StrategyLeastConnections:
		return "LeastConnections"
	case StrategyWeightedRoundRobin:
		return "WeightedRoundRobin"
	case StrategyLeastResponseTime:
		return "LeastResponseTime"
	default:
		return "Unknown"
	}
}

// UpdateConnectionCount 更新连接数
func (olb *OptimizedLoadBalancer) UpdateConnectionCount(podIndex int, count int64) {
	olb.mu.Lock()
	defer olb.mu.Unlock()

	if state, exists := olb.connectionState[podIndex]; exists {
		state.ConnectionCount = count
		state.LastUpdate = time.Now()
	} else {
		olb.connectionState[podIndex] = &ConnectionState{
			ConnectionCount: count,
			LastUpdate:      time.Now(),
			PodIndex:        podIndex,
		}
	}
}

// GetConnectionCount 获取连接数
func (olb *OptimizedLoadBalancer) GetConnectionCount(podIndex int) int64 {
	olb.mu.RLock()
	defer olb.mu.RUnlock()

	if state, exists := olb.connectionState[podIndex]; exists {
		return state.ConnectionCount
	}
	return 0
}

// ResetStats 重置统计信息
func (olb *OptimizedLoadBalancer) ResetStats() {
	olb.mu.Lock()
	defer olb.mu.Unlock()

	olb.stats = &LoadBalancerStats{
		StrategyRequests: make(map[string]int64),
	}
	olb.logger.Log(log.LevelInfo, "Load balancer stats reset")
}
