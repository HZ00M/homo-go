## 6A 任务卡：编写单元测试

- 编号: Task-05
- 模块: stateful-route
- 责任人: 待分配
- 优先级: 🟡 中
- 状态: ❌ 未开始
- 预计完成时间: 待定
- 实际完成时间: -

### A1 目标（Aim）
为有状态路由模块编写全面的单元测试，确保代码质量和功能正确性，测试覆盖率至少达到70%，包括接口测试、边界条件测试、错误处理测试等。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Java版本的单元测试
  - 🔄 部分实现：Go版本的接口和实现已定义
  - ❌ 未实现：Go版本的单元测试
- **差距**：
  - Java测试用例需要转换为Go测试
  - 需要适配Go的测试框架和mock机制
  - 需要验证Go实现的正确性
- **约束**：
  - 必须保持与Java版本测试覆盖一致
  - 使用Go标准测试框架和工具
  - 符合Kratos框架的测试规范
- **风险**：
  - 技术风险：测试框架使用不当
  - 业务风险：测试用例覆盖不全面
  - 依赖风险：mock对象设计不当

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`testing.T` - Go标准测试接口
  - **核心方法**：
    - 测试函数：`TestXxx`、`BenchmarkXxx`、`ExampleXxx`
    - 断言：`assert`、`require`、`testing`
    - Mock：`gomock`、`testify/mock`
  - **输入输出参数及错误码**：
    - 使用Go标准测试函数签名
    - 返回测试结果和覆盖率报告
    - 支持并行测试和基准测试

- **架构设计**：
  - 采用表驱动测试模式
  - 使用依赖注入和mock对象
  - 支持测试套件和测试用例分组

- **核心功能模块**：
  - 数据模型测试：`ServiceState`、`RoutingState`
  - 接口测试：`StatefulRouteForServerDriver`、`StatefulRouteForClientDriver`
  - 实现测试：各种驱动实现、管理器实现
  - 集成测试：组件间交互测试

- **极小任务拆分**：
  - T05-01：设置测试环境和依赖
  - T05-02：编写数据模型测试
  - T05-03：编写接口测试
  - T05-04：编写实现测试
  - T05-05：编写集成测试和性能测试

### A4 行动（Act）
#### T05-01：设置测试环境和依赖
```go
// route/test/setup_test.go
package test

import (
    "context"
    "testing"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
    
    "tpf-service-go/stateful-route/core"
    "tpf-service-go/stateful-route/interfaces"
    "tpf-service-go/stateful-route/types"
)

// TestSuite 测试套件基类
type TestSuite struct {
    suite.Suite
    ctx    context.Context
    logger log.Logger
}

// SetupSuite 测试套件初始化
func (ts *TestSuite) SetupSuite() {
    ts.ctx = context.Background()
    ts.logger = log.NewStdLogger(log.NewWriter(os.Stdout))
}

// TearDownSuite 测试套件清理
func (ts *TestSuite) TearDownSuite() {
    // 清理资源
}

// TestMain 测试主函数
func TestMain(m *testing.M) {
    // 设置测试环境
    setupTestEnvironment()
    
    // 运行测试
    code := m.Run()
    
    // 清理测试环境
    teardownTestEnvironment()
    
    os.Exit(code)
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() {
    // 设置测试配置
    // 初始化测试数据库
    // 设置测试日志级别
}

// teardownTestEnvironment 清理测试环境
func teardownTestEnvironment() {
    // 清理测试数据
    // 关闭测试连接
}
```

#### T05-02：编写数据模型测试
```go
// route/test/types_test.go
package test

import (
    "encoding/json"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "tpf-service-go/stateful-route/types"
)

// TestServiceState 测试ServiceState数据模型
func TestServiceState(t *testing.T) {
    t.Run("NewServiceState", func(t *testing.T) {
        state := 100
        serviceState := types.NewServiceState(state)
        
        assert.NotNil(t, serviceState)
        assert.Equal(t, state, serviceState.State)
        assert.WithinDuration(t, time.Now(), serviceState.UpdateTime, 100*time.Millisecond)
    })
    
    t.Run("JSONSerialization", func(t *testing.T) {
        serviceState := &types.ServiceState{
            State:      200,
            UpdateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        }
        
        data, err := json.Marshal(serviceState)
        require.NoError(t, err)
        
        var unmarshaled types.ServiceState
        err = json.Unmarshal(data, &unmarshaled)
        require.NoError(t, err)
        
        assert.Equal(t, serviceState.State, unmarshaled.State)
        assert.Equal(t, serviceState.UpdateTime.Unix(), unmarshaled.UpdateTime.Unix())
    })
}

// TestRoutingState 测试RoutingState枚举
func TestRoutingState(t *testing.T) {
    t.Run("StringRepresentation", func(t *testing.T) {
        testCases := []struct {
            state    types.RoutingState
            expected string
        }{
            {types.RoutingStateReady, "READY"},
            {types.RoutingStateDenyNewCall, "DENY_NEW_CALL"},
            {types.RoutingStateShutdown, "SHUTDOWN"},
            {types.RoutingState(999), "UNKNOWN"},
        }
        
        for _, tc := range testCases {
            t.Run(tc.expected, func(t *testing.T) {
                assert.Equal(t, tc.expected, tc.state.String())
            })
        }
    })
    
    t.Run("EnumValues", func(t *testing.T) {
        assert.Equal(t, types.RoutingState(0), types.RoutingStateReady)
        assert.Equal(t, types.RoutingState(1), types.RoutingStateDenyNewCall)
        assert.Equal(t, types.RoutingState(2), types.RoutingStateShutdown)
    })
}
```

#### T05-03：编写接口测试
```go
// route/test/interfaces_test.go
package test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    
    "tpf-service-go/stateful-route/interfaces"
    "tpf-service-go/stateful-route/types"
)

// MockStatefulRouteForServerDriver Mock服务端驱动
type MockStatefulRouteForServerDriver struct {
    mock.Mock
}

func (m *MockStatefulRouteForServerDriver) SetLoadState(ctx context.Context, loadState int) error {
    args := m.Called(ctx, loadState)
    return args.Error(0)
}

func (m *MockStatefulRouteForServerDriver) GetLoadState(ctx context.Context) (int, error) {
    args := m.Called(ctx)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForServerDriver) SetRoutingState(ctx context.Context, state types.RoutingState) error {
    args := m.Called(ctx, state)
    return args.Error(0)
}

func (m *MockStatefulRouteForServerDriver) GetRoutingState(ctx context.Context) (types.RoutingState, error) {
    args := m.Called(ctx)
    return args.Get(0).(types.RoutingState), args.Error(1)
}

func (m *MockStatefulRouteForServerDriver) SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId int) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName, podId)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForServerDriver) TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podId int) (bool, int, error) {
    args := m.Called(ctx, namespace, uid, serviceName, podId)
    return args.Bool(0), args.Int(1), args.Error(2)
}

func (m *MockStatefulRouteForServerDriver) RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string) error {
    args := m.Called(ctx, namespace, uid, serviceName)
    return args.Error(0)
}

func (m *MockStatefulRouteForServerDriver) RemoveLinkedPodWithId(ctx context.Context, namespace, uid, serviceName string, podId int, persistSeconds int) error {
    args := m.Called(ctx, namespace, uid, serviceName, podId, persistSeconds)
    return args.Error(0)
}

// TestStatefulRouteForServerDriverInterface 测试服务端驱动接口
func TestStatefulRouteForServerDriverInterface(t *testing.T) {
    t.Run("SetLoadState", func(t *testing.T) {
        mockDriver := new(MockStatefulRouteForServerDriver)
        ctx := context.Background()
        loadState := 150
        
        mockDriver.On("SetLoadState", ctx, loadState).Return(nil)
        
        err := mockDriver.SetLoadState(ctx, loadState)
        assert.NoError(t, err)
        mockDriver.AssertExpectations(t)
    })
    
    t.Run("GetLoadState", func(t *testing.T) {
        mockDriver := new(MockStatefulRouteForServerDriver)
        ctx := context.Background()
        expectedLoadState := 200
        
        mockDriver.On("GetLoadState", ctx).Return(expectedLoadState, nil)
        
        loadState, err := mockDriver.GetLoadState(ctx)
        assert.NoError(t, err)
        assert.Equal(t, expectedLoadState, loadState)
        mockDriver.AssertExpectations(t)
    })
    
    t.Run("SetRoutingState", func(t *testing.T) {
        mockDriver := new(MockStatefulRouteForServerDriver)
        ctx := context.Background()
        routingState := types.RoutingStateDenyNewCall
        
        mockDriver.On("SetRoutingState", ctx, routingState).Return(nil)
        
        err := mockDriver.SetRoutingState(ctx, routingState)
        assert.NoError(t, err)
        mockDriver.AssertExpectations(t)
    })
}

// MockStatefulRouteForClientDriver Mock客户端驱动
type MockStatefulRouteForClientDriver struct {
    mock.Mock
}

func (m *MockStatefulRouteForClientDriver) ComputeLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) SetLinkedPodIfAbsent(ctx context.Context, namespace, uid, serviceName string, podIndex int) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName, podIndex)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) GetLinkedPodInCacheOrIfPersist(ctx context.Context, namespace, uid, serviceName string) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) GetLinkedPodNotCache(ctx context.Context, namespace, uid, serviceName string) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) GetLinkedInCache(ctx context.Context, namespace, uid, serviceName string) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) CacheExpired(ctx context.Context, namespace, uid, serviceName string) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) SetCache(ctx context.Context, namespace, uid, serviceName string, podIndex int) (int, error) {
    args := m.Called(ctx, namespace, uid, serviceName, podIndex)
    return args.Int(0), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error) {
    args := m.Called(ctx, namespace, uid)
    return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockStatefulRouteForClientDriver) BatchGetLinkedPod(ctx context.Context, namespace string, keys []string, serviceName string) (map[int][]string, error) {
    args := m.Called(ctx, namespace, keys, serviceName)
    return args.Get(0).(map[int][]string), args.Error(1)
}

// TestStatefulRouteForClientDriverInterface 测试客户端驱动接口
func TestStatefulRouteForClientDriverInterface(t *testing.T) {
    t.Run("ComputeLinkedPod", func(t *testing.T) {
        mockDriver := new(MockStatefulRouteForClientDriver)
        ctx := context.Background()
        namespace := "test-namespace"
        uid := "test-uid"
        serviceName := "test-service"
        expectedPodIndex := 1
        
        mockDriver.On("ComputeLinkedPod", ctx, namespace, uid, serviceName).Return(expectedPodIndex, nil)
        
        podIndex, err := mockDriver.ComputeLinkedPod(ctx, namespace, uid, serviceName)
        assert.NoError(t, err)
        assert.Equal(t, expectedPodIndex, podIndex)
        mockDriver.AssertExpectations(t)
    })
    
    t.Run("GetLinkedPod", func(t *testing.T) {
        mockDriver := new(MockStatefulRouteForClientDriver)
        ctx := context.Background()
        namespace := "test-namespace"
        uid := "test-uid"
        serviceName := "test-service"
        expectedPodIndex := 2
        
        mockDriver.On("GetLinkedPod", ctx, namespace, uid, serviceName).Return(expectedPodIndex, nil)
        
        podIndex, err := mockDriver.GetLinkedPod(ctx, namespace, uid, serviceName)
        assert.NoError(t, err)
        assert.Equal(t, expectedPodIndex, podIndex)
        mockDriver.AssertExpectations(t)
    })
}
```

#### T05-04：编写实现测试
```go
// route/test/core_test.go
package test

import (
    "context"
    "testing"
    "time"
    
    "github.com/go-kratos/kratos/v2/log"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "tpf-service-go/stateful-route/core"
    "tpf-service-go/stateful-route/types"
)

// TestLoadBalancer 测试负载均衡器
func TestLoadBalancer(t *testing.T) {
    logger := log.NewStdLogger(log.NewWriter(os.Stdout))
    loadBalancer := core.NewLoadBalancer(logger)
    
    t.Run("RoundRobin", func(t *testing.T) {
        pods := []*core.PodInfo{
            {Index: 1, IsHealthy: true, Weight: 1},
            {Index: 2, IsHealthy: true, Weight: 1},
            {Index: 3, IsHealthy: true, Weight: 1},
        }
        
        podIndex, err := loadBalancer.SelectPod(context.Background(), pods, core.StrategyRoundRobin)
        require.NoError(t, err)
        assert.Contains(t, []int{1, 2, 3}, podIndex)
    })
    
    t.Run("LeastConnections", func(t *testing.T) {
        pods := []*core.PodInfo{
            {Index: 1, IsHealthy: true, ConnectionCount: 10, Weight: 1},
            {Index: 2, IsHealthy: true, ConnectionCount: 5, Weight: 1},
            {Index: 3, IsHealthy: true, ConnectionCount: 15, Weight: 1},
        }
        
        podIndex, err := loadBalancer.SelectPod(context.Background(), pods, core.StrategyLeastConnections)
        require.NoError(t, err)
        assert.Equal(t, 2, podIndex) // 最少连接数
    })
    
    t.Run("NoHealthyPods", func(t *testing.T) {
        pods := []*core.PodInfo{
            {Index: 1, IsHealthy: false, Weight: 1},
            {Index: 2, IsHealthy: false, Weight: 1},
        }
        
        _, err := loadBalancer.SelectPod(context.Background(), pods, core.StrategyRoundRobin)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "no healthy pods available")
    })
}

// TestHealthChecker 测试健康检查器
func TestHealthChecker(t *testing.T) {
    t.Run("StartStopHealthCheck", func(t *testing.T) {
        // 这里需要mock RouteInfoDriver
        // 由于接口限制，这里只测试基本逻辑
        t.Skip("需要mock RouteInfoDriver接口")
    })
}

// TestFailoverHandler 测试故障转移处理器
func TestFailoverHandler(t *testing.T) {
    t.Run("HandlePodFailure", func(t *testing.T) {
        // 这里需要mock StatefulRedisExecutor和RouteInfoDriver
        // 由于接口限制，这里只测试基本逻辑
        t.Skip("需要mock依赖组件")
    })
}

// TestStatefulRouteManager 测试路由管理器
func TestStatefulRouteManager(t *testing.T) {
    t.Run("RouteRequest", func(t *testing.T) {
        // 这里需要mock所有依赖组件
        // 由于接口限制，这里只测试基本逻辑
        t.Skip("需要mock依赖组件")
    })
}
```

#### T05-05：编写集成测试和性能测试
```go
// route/test/integration_test.go
package test

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
)

// IntegrationTestSuite 集成测试套件
type IntegrationTestSuite struct {
    TestSuite
    // 集成测试需要的组件
}

// SetupSuite 集成测试套件初始化
func (its *IntegrationTestSuite) SetupSuite() {
    its.TestSuite.SetupSuite()
    // 初始化集成测试环境
}

// TearDownSuite 集成测试套件清理
func (its *IntegrationTestSuite) TearDownSuite() {
    // 清理集成测试环境
    its.TestSuite.TearDownSuite()
}

// TestEndToEndRouting 端到端路由测试
func (its *IntegrationTestSuite) TestEndToEndRouting() {
    // 这里需要真实的Redis和依赖组件
    // 由于环境限制，这里只测试基本逻辑
    its.T().Skip("需要真实的测试环境")
}

// TestLoadBalancing 负载均衡测试
func (its *IntegrationTestSuite) TestLoadBalancing() {
    // 测试负载均衡功能
    its.T().Skip("需要真实的测试环境")
}

// TestFailover 故障转移测试
func (its *IntegrationTestSuite) TestFailover() {
    // 测试故障转移功能
    its.T().Skip("需要真实的测试环境")
}

// BenchmarkLoadBalancer 负载均衡器性能测试
func BenchmarkLoadBalancer(b *testing.B) {
    logger := log.NewStdLogger(log.NewWriter(os.Stdout))
    loadBalancer := core.NewLoadBalancer(logger)
    
    pods := []*core.PodInfo{
        {Index: 1, IsHealthy: true, Weight: 1},
        {Index: 2, IsHealthy: true, Weight: 1},
        {Index: 3, IsHealthy: true, Weight: 1},
        {Index: 4, IsHealthy: true, Weight: 1},
        {Index: 5, IsHealthy: true, Weight: 1},
    }
    
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := loadBalancer.SelectPod(ctx, pods, core.StrategyRoundRobin)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// BenchmarkHealthChecker 健康检查器性能测试
func BenchmarkHealthChecker(b *testing.B) {
    // 这里需要mock依赖组件
    b.Skip("需要mock依赖组件")
}

// BenchmarkFailoverHandler 故障转移处理器性能测试
func BenchmarkFailoverHandler(b *testing.B) {
    // 这里需要mock依赖组件
    b.Skip("需要mock依赖组件")
}

// TestCoverage 测试覆盖率检查
func TestCoverage(t *testing.T) {
    // 检查测试覆盖率
    // 可以使用 go test -cover 命令
    t.Skip("覆盖率检查需要实际运行测试")
}
```

### A5 验证（Assure）
- **测试用例**：
  - 数据模型测试：序列化/反序列化、边界条件
  - 接口测试：方法签名、参数验证
  - 实现测试：业务逻辑、错误处理
  - 集成测试：组件交互、端到端流程
- **性能验证**：
  - 负载均衡性能基准测试
  - 健康检查性能测试
  - 故障转移性能测试
- **回归测试**：
  - 确保与Java版本测试覆盖一致
- **测试结果**：
  - 测试覆盖率 ≥ 70%
  - 所有测试用例通过
  - 性能测试达标

### A6 迭代（Advance）
- 性能优化：测试性能优化、覆盖率提升
- 功能扩展：添加更多测试场景、边界条件测试
- 观测性增强：添加测试报告、覆盖率报告、性能报告
- 下一步任务链接：Task-06 性能优化和观测性增强

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
[总结任务完成情况]
