package driver

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/route"
	"github.com/stretchr/testify/assert"
)

// MockLogger 模拟Logger接口
type MockLogger struct{}

func (m *MockLogger) Log(level log.Level, keyvals ...any) error {
	return nil
}

// MockStateChanged 模拟StateChanged接口
type MockStateChanged struct{}

func (m *MockStateChanged) OnStateChanged(namespace, serviceName string, podID int, pre, now *route.StatefulServiceState) {
	// 模拟实现
}

// TestRouteInfoDriverImpl_InterfaceImplementation 测试RouteInfoDriverImpl是否正确实现RouteInfoDriver接口
func TestRouteInfoDriverImpl_InterfaceImplementation(t *testing.T) {
	// 创建模拟组件
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	// 创建配置
	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
		StateCacheExpireSecs:  300,
		MaxRetryCount:         3,
		RetryDelayMs:          1000,
		HealthCheckIntervalMs: 30000,
		LogLevel:              "INFO",
		EnableMetrics:         true,
		EnableTracing:         true,
	}

	// 创建驱动实例
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	// 验证接口实现
	var _ route.RouteInfoDriver = driver
	assert.NotNil(t, driver)
}

// TestRouteInfoDriverImpl_GetLinkInfoCacheTimeSecs 测试获取链接信息缓存时间
func TestRouteInfoDriverImpl_GetLinkInfoCacheTimeSecs(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}

	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	cacheTime := driver.GetLinkInfoCacheTimeSecs()
	assert.Equal(t, 300, cacheTime)
}

// TestRouteDriverFactory_CreateRouteInfoDriver 测试驱动工厂创建驱动
func TestRouteDriverFactory_CreateRouteInfoDriver(t *testing.T) {
	mockLogger := &MockLogger{}
	factory := NewRouteDriverFactory(mockLogger)

	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	baseConfig := &route.StatefulBaseConfig{}

	// 创建驱动
	driver := factory.CreateRouteInfoDriver(baseConfig, mockStateCache, mockStatefulExecutor)

	// 验证驱动已创建
	assert.NotNil(t, driver)
	assert.IsType(t, &RouteInfoDriverImpl{}, driver)
}

// MockServiceStateCache 模拟ServiceStateCache接口
type MockServiceStateCache struct {
	routablePodsFunc func(namespace, serviceName string) map[int]*route.StatefulServiceState
	alivePodsFunc    func(namespace, serviceName string) map[int]*route.StatefulServiceState
}

func (m *MockServiceStateCache) GetOrder() int       { return 0 }
func (m *MockServiceStateCache) OnInitModule() error { return nil }
func (m *MockServiceStateCache) GetServiceBestPod(ctx context.Context, namespace, serviceName string) (int, error) {
	return 0, nil
}
func (m *MockServiceStateCache) IsPodAvailable(namespace, serviceName string, podIndex int) bool {
	return false
}
func (m *MockServiceStateCache) IsPodRoutable(namespace, serviceName string, podIndex int) bool {
	return false
}
func (m *MockServiceStateCache) AlivePods(namespace, serviceName string) map[int]*route.StatefulServiceState {
	if m.alivePodsFunc != nil {
		return m.alivePodsFunc(namespace, serviceName)
	}
	return nil
}
func (m *MockServiceStateCache) RoutablePods(namespace, serviceName string) map[int]*route.StatefulServiceState {
	if m.routablePodsFunc != nil {
		return m.routablePodsFunc(namespace, serviceName)
	}
	return nil
}
func (m *MockServiceStateCache) Run() {}

// MockStatefulExecutor 模拟StatefulExecutor接口
type MockStatefulExecutor struct {
	getServiceStateFunc  func(ctx context.Context, namespace, serviceName string) (map[int]string, error)
	getWorkloadStateFunc func(ctx context.Context, namespace, serviceName string) (string, error)
}

func (m *MockStatefulExecutor) SetServiceState(ctx context.Context, namespace, serviceName string, podID int, state string) error {
	return nil
}
func (m *MockStatefulExecutor) GetServiceState(ctx context.Context, namespace, serviceName string) (map[int]string, error) {
	if m.getServiceStateFunc != nil {
		return m.getServiceStateFunc(ctx, namespace, serviceName)
	}
	return nil, nil
}
func (m *MockStatefulExecutor) SetWorkloadState(ctx context.Context, namespace, serviceName, state string) error {
	return nil
}
func (m *MockStatefulExecutor) GetWorkloadState(ctx context.Context, namespace, serviceName string) (string, error) {
	if m.getWorkloadStateFunc != nil {
		return m.getWorkloadStateFunc(ctx, namespace, serviceName)
	}
	return "", nil
}
func (m *MockStatefulExecutor) GetWorkloadStateBatch(ctx context.Context, namespace string, serviceNames []string) (map[string]string, error) {
	return nil, nil
}
func (m *MockStatefulExecutor) SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (int, error) {
	return 0, nil
}
func (m *MockStatefulExecutor) TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (bool, int, error) {
	return false, 0, nil
}
func (m *MockStatefulExecutor) SetLinkedPodIfAbsent(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (int, error) {
	return 0, nil
}
func (m *MockStatefulExecutor) GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error) {
	return 0, nil
}
func (m *MockStatefulExecutor) GetLinkedPodIfPersist(ctx context.Context, namespace, uid, serviceName string) (int, error) {
	return 0, nil
}
func (m *MockStatefulExecutor) BatchGetLinkedPod(ctx context.Context, namespace string, keys []string, serviceName string) (map[int][]string, error) {
	return nil, nil
}
func (m *MockStatefulExecutor) RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string, persistSeconds int) (bool, error) {
	return false, nil
}
func (m *MockStatefulExecutor) RemoveLinkedPodWithId(ctx context.Context, namespace, uid, serviceName string, persistSeconds, podID int) (bool, error) {
	return false, nil
}
func (m *MockStatefulExecutor) GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error) {
	return nil, nil
}

// TestRouteInfoDriverImpl_GetReadyServiceState 测试获取就绪服务状态
func TestRouteInfoDriverImpl_GetReadyServiceState(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	namespace := "test-namespace"
	serviceName := "test-service"

	// 设置mock期望 - 直接返回模拟数据
	mockStateCache.routablePodsFunc = func(namespace, serviceName string) map[int]*route.StatefulServiceState {
		return map[int]*route.StatefulServiceState{
			1: {PodID: 1, State: route.ServiceStateReady},
			2: {PodID: 2, State: route.ServiceStateReady},
		}
	}

	// 调用方法
	result, err := driver.GetReadyServiceState(context.Background(), namespace, serviceName)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, 1)
	assert.Contains(t, result, 2)
}

// TestRouteInfoDriverImpl_GetAllServiceState 测试获取所有服务状态
func TestRouteInfoDriverImpl_GetAllServiceState(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	ctx := context.Background()
	namespace := "test-namespace"
	serviceName := "test-service"

	// 设置mock期望 - 直接返回模拟数据
	mockStatefulExecutor.getServiceStateFunc = func(ctx context.Context, namespace, serviceName string) (map[int]string, error) {
		return map[int]string{
			1: "READY",
			2: "NOT_READY",
		}, nil
	}

	// 调用方法
	result, err := driver.GetAllServiceState(ctx, namespace, serviceName)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, 1)
	assert.Contains(t, result, 2)

	// 验证状态转换
	assert.Equal(t, route.ServiceStateReady, result[1].State)
	assert.Equal(t, route.ServiceStateNotReady, result[2].State)
}

// TestRouteInfoDriverImpl_GetServiceBestPod 测试获取服务最佳Pod
func TestRouteInfoDriverImpl_GetServiceBestPod(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	ctx := context.Background()
	namespace := "test-namespace"
	serviceName := "test-service"

	// 设置mock期望 - 直接返回模拟数据
	mockStateCache.routablePodsFunc = func(namespace, serviceName string) map[int]*route.StatefulServiceState {
		return map[int]*route.StatefulServiceState{
			1: {PodID: 1, State: route.ServiceStateReady, LoadState: 1}, // 空闲状态
			2: {PodID: 2, State: route.ServiceStateReady, LoadState: 2}, // 低负载状态
		}
	}

	// 调用方法
	result, err := driver.GetServiceBestPod(ctx, namespace, serviceName)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 1, result) // Pod 1 负载更低，应该被选中
}

// TestRouteInfoDriverImpl_GetServiceBestPod_NoRoutablePods 测试没有可路由Pod的情况
func TestRouteInfoDriverImpl_GetServiceBestPod_NoRoutablePods(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	ctx := context.Background()
	namespace := "test-namespace"
	serviceName := "test-service"

	// 设置mock期望 - 没有可路由的Pod
	mockStateCache.routablePodsFunc = func(namespace, serviceName string) map[int]*route.StatefulServiceState {
		return map[int]*route.StatefulServiceState{}
	}

	// 调用方法
	result, err := driver.GetServiceBestPod(ctx, namespace, serviceName)

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not ready")
	assert.Equal(t, 0, result)
}

// TestRouteInfoDriverImpl_IsWorkloadReady 测试检查工作负载是否就绪
func TestRouteInfoDriverImpl_IsWorkloadReady(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	ctx := context.Background()
	namespace := "test-namespace"
	serviceName := "test-service"

	// 设置mock期望 - 直接返回模拟数据
	mockStatefulExecutor.getWorkloadStateFunc = func(ctx context.Context, namespace, serviceName string) (string, error) {
		return "READY", nil
	}

	// 调用方法
	result, err := driver.IsWorkloadReady(ctx, namespace, serviceName)

	// 验证结果
	assert.NoError(t, err)
	assert.True(t, result)
}

// TestRouteInfoDriverImpl_IsWorkloadReady_NotReady 测试工作负载不就绪的情况
func TestRouteInfoDriverImpl_IsWorkloadReady_NotReady(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	ctx := context.Background()
	namespace := "test-namespace"
	serviceName := "test-service"

	// 设置mock期望 - 直接返回模拟数据
	mockStatefulExecutor.getWorkloadStateFunc = func(ctx context.Context, namespace, serviceName string) (string, error) {
		return "NOT_READY", nil
	}

	// 调用方法
	result, err := driver.IsWorkloadReady(ctx, namespace, serviceName)

	// 验证结果
	assert.NoError(t, err)
	assert.False(t, result)
}

// TestRouteInfoDriverImpl_RegisterRoutingStateChangedEvent 测试注册路由状态变更事件
func TestRouteInfoDriverImpl_RegisterRoutingStateChangedEvent(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	namespace := "test-namespace"
	mockStateChanged := &MockStateChanged{}

	// 注册事件
	driver.RegisterRoutingStateChangedEvent(namespace, mockStateChanged)

	// 验证事件已注册
	// 这里我们通过调用OnRoutingStateChanged来验证
	preState := &route.StatefulServiceState{PodID: 1, State: route.ServiceStateReady}
	nowState := &route.StatefulServiceState{PodID: 1, State: route.ServiceStateNotReady}

	// 触发状态变更事件
	driver.OnRoutingStateChanged(namespace, "test-service", 1, preState, nowState)

	// 验证事件处理（这里只是验证没有panic，实际的事件处理逻辑需要更复杂的测试）
	assert.NotNil(t, driver)
}

// TestRouteInfoDriverImpl_OnRoutingStateChanged 测试路由状态变更通知
func TestRouteInfoDriverImpl_OnRoutingStateChanged(t *testing.T) {
	mockStateCache := &MockServiceStateCache{}
	mockStatefulExecutor := &MockStatefulExecutor{}
	mockLogger := &MockLogger{}

	baseConfig := &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
	}
	driver := NewRouteInfoDriverImpl(baseConfig, mockStateCache, mockStatefulExecutor, mockLogger)

	namespace := "test-namespace"
	serviceName := "test-service"
	mockStateChanged := &MockStateChanged{}

	// 注册事件
	driver.RegisterRoutingStateChangedEvent(namespace, mockStateChanged)

	// 触发状态变更
	preState := &route.StatefulServiceState{PodID: 1, State: route.ServiceStateReady}
	nowState := &route.StatefulServiceState{PodID: 1, State: route.ServiceStateNotReady}

	// 调用方法（应该不会panic）
	driver.OnRoutingStateChanged(namespace, serviceName, 1, preState, nowState)

	// 验证没有panic
	assert.NotNil(t, driver)
}
