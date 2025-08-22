## 6A 任务卡：单元测试与集成测试

- 编号: Task-05
- 模块: route/executor
- 责任人: 待分配
- 优先级: 🟡 中
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）
为有状态服务执行器实现完整的单元测试和集成测试，确保代码质量和功能正确性，支持测试覆盖率≥80%，为后续的性能优化和监控提供可靠的测试基础。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：核心功能已在 `stateful_executor.go` 中完成
  - ✅ 已完成：接口实现测试已在 `interface_test.go` 中完成
  - ❌ 未实现：完整的单元测试和集成测试
  - ❌ 未实现：测试覆盖率统计
- **差距**：
  - 需要实现所有方法的单元测试
  - 需要实现Redis集成的集成测试
  - 需要实现性能基准测试
  - 需要实现测试覆盖率统计
- **约束**：
  - 必须保持与Java版本功能的兼容性
  - 必须符合Go语言的测试规范
  - 必须支持Kratos框架的测试
- **风险**：
  - 测试覆盖率不足可能导致质量问题
  - 集成测试需要Redis环境支持

### A3 设计（Architect）
- **接口契约**：
  - **核心功能**：单元测试、集成测试、性能测试
  - **核心特性**：
    - 使用Go标准测试框架
    - 支持Mock和Stub测试
    - 支持测试覆盖率统计
    - 支持性能基准测试

- **架构设计**：
  - 采用单文件架构，所有测试围绕 `StatefulExecutorImpl` 进行
  - 使用 `testify` 框架进行断言和Mock
  - 支持Redis集成测试和Mock测试
  - 支持并发测试和性能测试

- **核心功能模块**：
  - `UnitTests`: 单元测试集合
  - `IntegrationTests`: 集成测试集合
  - `PerformanceTests`: 性能测试集合
  - `TestHelpers`: 测试辅助工具

- **极小任务拆分**：
  - T05-01：实现服务状态管理单元测试
  - T05-02：实现工作负载状态管理单元测试
  - T05-03：实现Pod链接管理单元测试
  - T05-04：实现Redis集成测试
  - T05-05：实现性能基准测试
  - T05-06：实现测试覆盖率统计

### A4 行动（Act）
#### T05-01：实现服务状态管理单元测试
```go
// route/executor/stateful_executor_test.go
package executor

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/redis/go-redis/v9"
)

// MockRedisClient Redis客户端Mock
type MockRedisClient struct {
    mock.Mock
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
    args := m.Called(ctx, key, values)
    return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
    args := m.Called(ctx, key)
    return args.Get(0).(*redis.StringStringMapCmd)
}

// TestSetServiceState 测试设置服务状态
func TestSetServiceState(t *testing.T) {
    mockClient := new(MockRedisClient)
    logger := &MockLogger{}
    executor := NewStatefulExecutor(mockClient, logger)

    ctx := context.Background()
    namespace := "default"
    serviceName := "test-service"
    podID := 1
    state := "running"

    // 设置期望
    mockClient.On("HSet", ctx, mock.AnythingOfType("string"), "1", state).
        Return(redis.NewIntCmd(ctx))

    // 执行测试
    err := executor.SetServiceState(ctx, namespace, serviceName, podID, state)

    // 验证结果
    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}

// TestGetServiceState 测试获取服务状态
func TestGetServiceState(t *testing.T) {
    mockClient := new(MockRedisClient)
    logger := &MockLogger{}
    executor := NewStatefulExecutor(mockClient, logger)

    ctx := context.Background()
    namespace := "default"
    serviceName := "test-service"

    // 设置期望
    expectedStates := map[string]string{
        "0": "running",
        "1": "ready",
    }
    mockClient.On("HGetAll", ctx, mock.AnythingOfType("string")).
        Return(redis.NewStringStringMapCmd(ctx, expectedStates))

    // 执行测试
    states, err := executor.GetServiceState(ctx, namespace, serviceName)

    // 验证结果
    assert.NoError(t, err)
    assert.Equal(t, map[int]string{0: "running", 1: "ready"}, states)
    mockClient.AssertExpectations(t)
}
```

#### T05-02：实现工作负载状态管理单元测试
```go
// route/executor/stateful_executor_test.go
package executor

// TestSetWorkloadState 测试设置工作负载状态
func TestSetWorkloadState(t *testing.T) {
    mockClient := new(MockRedisClient)
    logger := &MockLogger{}
    executor := NewStatefulExecutor(mockClient, logger)

    ctx := context.Background()
    namespace := "default"
    serviceName := "test-service"
    state := "healthy"

    // 设置期望
    mockClient.On("Set", ctx, mock.AnythingOfType("string"), state, mock.AnythingOfType("time.Duration")).
        Return(redis.NewStatusCmd(ctx))

    // 执行测试
    err := executor.SetWorkloadState(ctx, namespace, serviceName, state)

    // 验证结果
    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}

// TestGetWorkloadStateBatch 测试批量获取工作负载状态
func TestGetWorkloadStateBatch(t *testing.T) {
    mockClient := new(MockRedisClient)
    logger := &MockLogger{}
    executor := NewStatefulExecutor(mockClient, logger)

    ctx := context.Background()
    namespace := "default"
    serviceNames := []string{"service1", "service2"}

    // 设置期望
    expectedValues := []interface{}{"healthy", "unhealthy"}
    mockClient.On("MGet", ctx, mock.AnythingOfType("[]string")).
        Return(redis.NewSliceCmd(ctx, expectedValues...))

    // 执行测试
    states, err := executor.GetWorkloadStateBatch(ctx, namespace, serviceNames)

    // 验证结果
    assert.NoError(t, err)
    assert.Equal(t, map[string]string{"service1": "healthy", "service2": "unhealthy"}, states)
    mockClient.AssertExpectations(t)
}
```

#### T05-03：实现Pod链接管理单元测试
```go
// route/executor/stateful_executor_test.go
package executor

// TestSetLinkedPod 测试设置Pod链接
func TestSetLinkedPod(t *testing.T) {
    mockClient := new(MockRedisClient)
    logger := &MockLogger{}
    executor := NewStatefulExecutor(mockClient, logger)

    ctx := context.Background()
    namespace := "default"
    uid := "test-uid"
    serviceName := "test-service"
    podID := 1
    persistSeconds := 3600

    // 设置期望
    mockClient.On("Eval", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface{}")).
        Return(redis.NewCmd(ctx, []interface{}{"1"}))

    // 执行测试
    result, err := executor.SetLinkedPod(ctx, namespace, uid, serviceName, podID, persistSeconds)

    // 验证结果
    assert.NoError(t, err)
    assert.Equal(t, 1, result)
    mockClient.AssertExpectations(t)
}

// TestGetLinkedPod 测试获取Pod链接
func TestGetLinkedPod(t *testing.T) {
    mockClient := new(MockRedisClient)
    logger := &MockLogger{}
    executor := NewStatefulExecutor(mockClient, logger)

    ctx := context.Background()
    namespace := "default"
    uid := "test-uid"
    serviceName := "test-service"

    // 设置期望
    mockClient.On("Get", ctx, mock.AnythingOfType("string")).
        Return(redis.NewStringCmd(ctx, "1"))

    // 执行测试
    podID, err := executor.GetLinkedPod(ctx, namespace, uid, serviceName)

    // 验证结果
    assert.NoError(t, err)
    assert.Equal(t, 1, podID)
    mockClient.AssertExpectations(t)
}
```

#### T05-04：实现Redis集成测试
```go
// route/executor/stateful_executor_integration_test.go
package executor

import (
    "context"
    "testing"
    "github.com/redis/go-redis/v9"
    "github.com/stretchr/testify/assert"
)

// TestRedisIntegration 测试Redis集成
func TestRedisIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过集成测试")
    }

    // 创建Redis客户端
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   1, // 使用测试数据库
    })
    defer client.Close()

    // 创建执行器
    logger := &MockLogger{}
    executor := NewStatefulExecutor(client, logger)

    ctx := context.Background()
    namespace := "test"
    serviceName := "integration-test"
    podID := 0
    state := "running"

    // 测试设置服务状态
    err := executor.SetServiceState(ctx, namespace, serviceName, podID, state)
    assert.NoError(t, err)

    // 测试获取服务状态
    states, err := executor.GetServiceState(ctx, namespace, serviceName)
    assert.NoError(t, err)
    assert.Equal(t, state, states[podID])

    // 清理测试数据
    client.Del(ctx, executor.formatServiceStateRedisKey(namespace, serviceName))
}
```

#### T05-05：实现性能基准测试
```go
// route/executor/stateful_executor_benchmark_test.go
package executor

import (
    "context"
    "testing"
    "github.com/redis/go-redis/v9"
)

// BenchmarkSetServiceState 基准测试设置服务状态
func BenchmarkSetServiceState(b *testing.B) {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   1,
    })
    defer client.Close()

    logger := &MockLogger{}
    executor := NewStatefulExecutor(client, logger)

    ctx := context.Background()
    namespace := "benchmark"
    serviceName := "benchmark-service"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        podID := i % 10
        state := "running"
        err := executor.SetServiceState(ctx, namespace, serviceName, podID, state)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// BenchmarkGetServiceState 基准测试获取服务状态
func BenchmarkGetServiceState(b *testing.B) {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   1,
    })
    defer client.Close()

    logger := &MockLogger{}
    executor := NewStatefulExecutor(client, logger)

    ctx := context.Background()
    namespace := "benchmark"
    serviceName := "benchmark-service"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := executor.GetServiceState(ctx, namespace, serviceName)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

#### T05-06：实现测试覆盖率统计
```bash
# 运行测试并生成覆盖率报告
go test -cover -coverprofile=coverage.out ./executor/
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out
```

### A5 验证（Assure）
- **测试用例**：
  - [ ] 服务状态管理测试
  - [ ] 工作负载状态管理测试
  - [ ] Pod链接管理测试
  - [ ] Redis集成测试
  - [ ] 性能基准测试
- **性能验证**：
  - [ ] 测试覆盖率 ≥80%
  - [ ] 性能基准测试通过
- **回归测试**：
  - [ ] 所有现有功能测试通过
- **测试结果**：
  - [ ] 待实现

### A6 迭代（Advance）
- 性能优化：基于测试结果进行性能优化
- 功能扩展：支持更多测试场景
- 观测性增强：添加测试监控和报告
- 下一步任务链接：Task-06 性能优化与监控（待开始）

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
Task-05待开始，需要实现完整的单元测试和集成测试，包括：
1. 所有方法的单元测试
2. Redis集成的集成测试
3. 性能基准测试
4. 测试覆盖率统计
5. 支持测试环境配置
6. 为后续性能优化提供测试基础

下一步将进行Task-06的实现。
