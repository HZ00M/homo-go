## 6A 任务卡：核心业务逻辑实现

- 编号: Task-03
- 模块: route/executor
- 责任人: AI助手
- 优先级: 🔴 高
- 状态: ✅ 已完成
- 预计完成时间: 2025-01-27
- 实际完成时间: 2025-01-27

### A1 目标（Aim）
实现有状态服务执行器的核心业务逻辑，包括服务状态管理、工作负载状态管理、Pod链接管理等核心功能，确保与Java版本的功能完全一致，支持Go语言的context和error模式。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：Java版本的核心业务逻辑完整
  - ✅ 已完成：Go版本的核心业务逻辑已在 `stateful_executor.go` 中完全实现
  - ✅ 已完成：所有接口方法都已实现
  - ✅ 已完成：与Java版本功能完全一致
- **差距**：
  - 无，核心业务逻辑已完全实现
- **约束**：
  - 已保持与Java版本功能的兼容性
  - 已符合Go语言的业务逻辑实现规范
  - 已支持Kratos框架的依赖注入
- **风险**：
  - 无，核心业务逻辑已实现

### A3 设计（Architect）
- **接口契约**：
  - **核心功能**：服务状态管理、工作负载状态管理、Pod链接管理
  - **核心方法**：
    - 服务状态管理：`SetServiceState`、`GetServiceState`
    - 工作负载状态管理：`SetWorkloadState`、`GetWorkloadState`、`GetWorkloadStateBatch`
    - Pod链接管理：`SetLinkedPod`、`TrySetLinkedPod`、`GetLinkedPod`、`RemoveLinkedPod`等

- **架构设计**：
  - 采用单文件架构，所有业务逻辑集成在 `StatefulExecutorImpl` 中
  - 直接使用 `*redis.Client` 进行Redis操作
  - 使用 `embed.FS` 嵌入Lua脚本
  - 支持完整的错误处理和日志记录

- **核心功能模块**：
  - `StatefulExecutorImpl`: 主执行器，包含所有业务逻辑
  - `LuaScripts`: 嵌入的Lua脚本集合
  - `RedisOperations`: Redis操作封装

- **极小任务拆分**：
  - ✅ T03-01：实现服务状态管理（已完成）
  - ✅ T03-02：实现工作负载状态管理（已完成）
  - ✅ T03-03：实现Pod链接管理（已完成）
  - ✅ T03-04：实现辅助方法和工具函数（已完成）

### A4 行动（Act）
#### ✅ T03-01：实现服务状态管理（已完成）
```go
// route/executor/stateful_executor.go
package executor

// SetServiceState 设置服务中特定Pod的状态
func (e *StatefulExecutorImpl) SetServiceState(ctx context.Context, namespace, serviceName string, podID int, state string) error {
    e.logger.Log(log.LevelDebug, "msg", "设置服务状态", "namespace", namespace, "serviceName", serviceName, "podID", podID, "state", state)

    // 验证参数
    if err := e.validateServiceStateParams(namespace, serviceName, podID, state); err != nil {
        return fmt.Errorf("参数验证失败: %w", err)
    }

    // 构建Redis键
    stateKey := e.formatServiceStateRedisKey(namespace, serviceName)
    podNumStr := strconv.Itoa(podID)

    // 使用HSET设置Pod状态
    result := e.redisClient.HSet(ctx, stateKey, podNumStr, state)
    if result.Err() != nil {
        return fmt.Errorf("设置服务状态到Redis失败: %w", result.Err())
    }

    e.logger.Log(log.LevelInfo, "msg", "服务状态设置成功", "key", stateKey, "podID", podID, "state", state)
    return nil
}

// GetServiceState 获取特定服务的所有Pod状态
func (e *StatefulExecutorImpl) GetServiceState(ctx context.Context, namespace, serviceName string) (map[int]string, error) {
    e.logger.Log(log.LevelDebug, "msg", "获取服务状态", "namespace", namespace, "serviceName", serviceName)

    // 验证参数
    if err := e.validateServiceParams(namespace, serviceName); err != nil {
        return nil, fmt.Errorf("参数验证失败: %w", err)
    }

    serviceKey := e.formatServiceStateRedisKey(namespace, serviceName)

    // 使用HGETALL获取所有Pod状态
    result := e.redisClient.HGetAll(ctx, serviceKey)
    if result.Err() != nil {
        return nil, fmt.Errorf("获取服务状态失败: %w", result.Err())
    }

    // 转换结果
    retMap := make(map[int]string)
    for key, value := range result.Val() {
        if podID, err := strconv.Atoi(key); err == nil {
            retMap[podID] = value
        }
    }

    e.logger.Log(log.LevelInfo, "msg", "获取服务状态成功", "namespace", namespace, "serviceName", serviceName, "count", len(retMap))
    return retMap, nil
}
```

#### ✅ T03-02：实现工作负载状态管理（已完成）
```go
// route/executor/stateful_executor.go
package executor

// SetWorkloadState 设置整个工作负载（服务）的状态
func (e *StatefulExecutorImpl) SetWorkloadState(ctx context.Context, namespace, serviceName, state string) error {
    e.logger.Log(log.LevelDebug, "msg", "设置工作负载状态", "namespace", namespace, "serviceName", serviceName, "state", state)

    // 验证参数
    if err := e.validateServiceParams(namespace, serviceName); err != nil {
        return fmt.Errorf("参数验证失败: %w", err)
    }

    stateQueryKey := e.formatWorkloadStateRedisKey(namespace, serviceName)

    result := e.redisClient.Set(ctx, stateQueryKey, state, 0)
    if result.Err() != nil {
        return fmt.Errorf("设置工作负载状态失败: %w", result.Err())
    }

    e.logger.Log(log.LevelInfo, "msg", "工作负载状态设置成功", "key", stateQueryKey, "state", state)
    return nil
}

// GetWorkloadState 获取特定工作负载的状态
func (e *StatefulExecutorImpl) GetWorkloadState(ctx context.Context, namespace, serviceName string) (string, error) {
    svcs := []string{serviceName}
    result, err := e.GetWorkloadStateBatch(ctx, namespace, svcs)
    if err != nil {
        return "", err
    }
    return result[serviceName], nil
}

// GetWorkloadStateBatch 批量获取多个工作负载的状态
func (e *StatefulExecutorImpl) GetWorkloadStateBatch(ctx context.Context, namespace string, serviceNames []string) (map[string]string, error) {
    e.logger.Log(log.LevelDebug, "msg", "批量获取工作负载状态", "namespace", namespace, "serviceNames", serviceNames)

    retMap := make(map[string]string)
    if len(serviceNames) == 0 {
        e.logger.Log(log.LevelDebug, "msg", "getWorkloadState keys empty")
        return retMap, nil
    }

    // 构建所有键
    serviceRedisKeys := make([]string, len(serviceNames))
    for i, k := range serviceNames {
        serviceRedisKeys[i] = e.formatWorkloadStateRedisKey(namespace, k)
    }

    // 批量获取
    result := e.redisClient.MGet(ctx, serviceRedisKeys...)
    if result.Err() != nil {
        return nil, fmt.Errorf("批量获取工作负载状态失败: %w", result.Err())
    }

    if len(result.Val()) != len(serviceRedisKeys) {
        return nil, fmt.Errorf("getWorkloadState 返回的数量:%d和传入:%d不相等", len(result.Val()), len(serviceRedisKeys))
    }

    // 处理结果
    for i, val := range result.Val() {
        svc := serviceNames[i]
        if val != nil && val != "" && val != "null" && val != "nil" {
            retMap[svc] = val.(string)
        }
    }

    return retMap, nil
}
```

#### ✅ T03-03：实现Pod链接管理（已完成）
```go
// route/executor/stateful_executor.go
package executor

// SetLinkedPod 将Pod与特定UID建立持久链接
func (e *StatefulExecutorImpl) SetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (int, error) {
    e.logger.Log(log.LevelDebug, "msg", "设置Pod链接", "namespace", namespace, "uid", uid, "serviceName", serviceName, "podID", podID, "persistSeconds", persistSeconds)

    keys, args := e.createUidKeysAndArgs(namespace, uid, serviceName, podID, persistSeconds)

    result, err := e.executeScript(ctx, "statefulSetLink", keys, args)
    if err != nil {
        return -1, fmt.Errorf("执行SetLink脚本失败: %w", err)
    }

    return e.parseIntResult(result)
}

// TrySetLinkedPod 尝试建立Pod链接，返回操作是否成功以及当前链接的Pod
func (e *StatefulExecutorImpl) TrySetLinkedPod(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (bool, int, error) {
    e.logger.Log(log.LevelDebug, "msg", "尝试设置Pod链接", "namespace", namespace, "uid", uid, "serviceName", serviceName, "podID", podID, "persistSeconds", persistSeconds)

    keys, args := e.createUidKeysAndArgs(namespace, uid, serviceName, podID, persistSeconds)

    result, err := e.executeScript(ctx, "statefulTrySetLink", keys, args)
    if err != nil {
        return false, -1, fmt.Errorf("执行TrySetLink脚本失败: %w", err)
    }

    return e.parseTrySetLinkResult(result, podID)
}

// GetLinkedPod 获取UID和服务当前链接的Pod
func (e *StatefulExecutorImpl) GetLinkedPod(ctx context.Context, namespace, uid, serviceName string) (int, error) {
    e.logger.Log(log.LevelDebug, "msg", "获取链接的Pod", "namespace", namespace, "uid", uid, "serviceName", serviceName)

    uidSvcKey := e.formatUidWithSpecificSvcLinkRedisKey(namespace, serviceName, uid)

    result := e.redisClient.Get(ctx, uidSvcKey)
    if result.Err() != nil {
        if result.Err() == redis.Nil {
            return -1, nil // 没有链接
        }
        return -1, fmt.Errorf("获取Pod链接失败: %w", result.Err())
    }

    podIndexStr := result.Val()
    if podIndexStr == "" {
        return -1, nil
    }

    podIndex, err := strconv.Atoi(podIndexStr)
    if err != nil {
        return -1, fmt.Errorf("无效的Pod索引: %s", podIndexStr)
    }

    return podIndex, nil
}
```

#### ✅ T03-04：实现辅助方法和工具函数（已完成）
```go
// route/executor/stateful_executor.go
package executor

// createUidKeysAndArgs 创建Pod链接操作所需的Redis键和参数数组
func (e *StatefulExecutorImpl) createUidKeysAndArgs(namespace, uid, serviceName string, podID, persistSeconds int) ([]string, []interface{}) {
    uidSvcKey := e.formatUidWithSpecificSvcLinkRedisKey(namespace, serviceName, uid)
    uidKey := e.formatUidLinkRedisKey(namespace, uid)

    keys := []string{uidSvcKey, uidKey}
    args := []interface{}{serviceName, podID, persistSeconds}

    return keys, args
}

// parseIntResult 解析整数结果
func (e *StatefulExecutorImpl) parseIntResult(result interface{}) (int, error) {
    if resultList, ok := result.([]interface{}); ok && len(resultList) > 0 {
        if podStr, ok := resultList[0].(string); ok {
            if podID, err := strconv.Atoi(podStr); err == nil {
                if podID == -1 {
                    return -1, nil
                }
                return podID, nil
            }
        }
    }
    return -1, fmt.Errorf("无效的脚本结果")
}

// validateServiceStateParams 验证服务状态参数
func (e *StatefulExecutorImpl) validateServiceStateParams(namespace, serviceName string, podID int, state string) error {
    if strings.TrimSpace(namespace) == "" {
        return fmt.Errorf("namespace不能为空")
    }
    if strings.TrimSpace(serviceName) == "" {
        return fmt.Errorf("serviceName不能为空")
    }
    if podID < 0 {
        return fmt.Errorf("podID不能为负数")
    }
    if strings.TrimSpace(state) == "" {
        return fmt.Errorf("state不能为空")
    }
    return nil
}
```

### A5 验证（Assure）
- **测试用例**：
  - ✅ 服务状态管理测试（已完成）
  - ✅ 工作负载状态管理测试（已完成）
  - ✅ Pod链接管理测试（已完成）
  - ✅ 参数验证测试（已完成）
- **性能验证**：
  - ✅ 业务逻辑性能测试（已完成）
  - ✅ Redis操作性能测试（已完成）
- **回归测试**：
  - ✅ 与Java版本功能兼容性验证（已完成）
- **测试结果**：
  - ✅ 所有测试通过

### A6 迭代（Advance）
- 性能优化：业务逻辑已优化，减少不必要的操作
- 功能扩展：支持更多状态类型和操作
- 观测性增强：已添加详细的日志记录和错误处理
- 下一步任务链接：Task-04 错误处理与日志系统（已完成）

### 📋 质量检查
- [x] 代码质量检查完成
- [x] 文档质量检查完成
- [x] 测试质量检查完成

### 📋 完成总结
Task-03已成功完成，核心业务逻辑已完全实现，包括：
1. 完整的服务状态管理功能
2. 完整的工作负载状态管理功能
3. 完整的Pod链接管理功能
4. 所有必要的辅助方法和工具函数
5. 与Java版本功能的完全兼容性
6. 采用单文件架构，减少模块复杂度

下一步将进行Task-04的实现。
