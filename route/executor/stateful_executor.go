package executor

import (
	"context"
	"embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

//go:embed lua_scripts/*.lua
var scriptFS embed.FS

// ==================== 数据结构定义 ====================

// ServiceLinkInfo 服务链接信息
type ServiceLinkInfo struct {
	UID         string `json:"uid"`         // 用户ID
	Namespace   string `json:"namespace"`   // 命名空间
	ServiceName string `json:"serviceName"` // 服务名称
	PodID       int    `json:"podID"`       // Pod ID
	UpdateTime  int64  `json:"updateTime"`  // 更新时间
}

// ==================== 有状态执行器实现 ====================

// StatefulExecutorImpl 有状态执行器实现
// 实现StatefulExecutor接口，提供有状态服务的核心操作
type StatefulExecutorImpl struct {
	redisClient *redis.Client
	logger      log.Logger
	scriptCache map[string]string
}

// 编译时检查：确保StatefulExecutorImpl实现了StatefulExecutor接口
// 注意：这里使用编译时检查确保接口实现完整

// NewStatefulExecutor 创建新的有状态执行器
func NewStatefulExecutor(redisClient *redis.Client, logger log.Logger) *StatefulExecutorImpl {
	executor := &StatefulExecutorImpl{
		redisClient: redisClient,
		logger:      logger,
		scriptCache: make(map[string]string),
	}

	// 预加载所有Lua脚本
	if err := executor.preloadScripts(context.Background()); err != nil {
		logger.Log(log.LevelError, "msg", "Failed to preload Lua scripts", "error", err)
	}

	return executor
}

// ==================== Lua脚本管理 ====================

// preloadScripts 预加载所有Lua脚本
func (e *StatefulExecutorImpl) preloadScripts(ctx context.Context) error {
	scriptNames := []string{
		"statefulSetLink",
		"statefulSetLinkIfAbsent",
		"statefulTrySetLink",
		"statefulRemoveLink",
		"statefulRemoveLinkWithId",
		"statefulGetLinkIfPersist",
		"statefulComputeLinkIfAbsent",
		"statefulGetServicePod",
		"statefulGetService",
		"statefulSetState",
		"statefulGetLinkService",
	}

	for _, name := range scriptNames {
		if err := e.loadScript(ctx, name); err != nil {
			return fmt.Errorf("failed to load script %s: %w", name, err)
		}
	}

	return nil
}

// loadScript 加载单个Lua脚本
func (e *StatefulExecutorImpl) loadScript(ctx context.Context, name string) error {
	scriptPath := fmt.Sprintf("lua_scripts/%s.lua", name)
	scriptBytes, err := scriptFS.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script %s: %w", name, err)
	}

	script := string(scriptBytes)
	e.scriptCache[name] = script

	// 将脚本加载到Redis
	sha, err := e.redisClient.ScriptLoad(ctx, script).Result()
	if err != nil {
		return fmt.Errorf("failed to load script %s to Redis: %w", name, err)
	}

	e.logger.Log(log.LevelDebug, "msg", "Script loaded", "name", name, "sha", sha)
	return nil
}

// executeScript 执行Lua脚本
func (e *StatefulExecutorImpl) executeScript(ctx context.Context, scriptName string, keys []string, args []interface{}) (interface{}, error) {
	script, exists := e.scriptCache[scriptName]
	if !exists {
		return nil, fmt.Errorf("script %s not found", scriptName)
	}

	result := e.redisClient.Eval(ctx, script, keys, args...)
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to execute script %s: %w", scriptName, result.Err())
	}

	return result.Val(), nil
}

// ==================== 服务状态管理 ====================

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

// ==================== 工作负载状态管理 ====================

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

// ==================== Pod链接管理 ====================

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

// SetLinkedPodIfAbsent 仅在不存在链接时设置Pod链接
func (e *StatefulExecutorImpl) SetLinkedPodIfAbsent(ctx context.Context, namespace, uid, serviceName string, podID, persistSeconds int) (int, error) {
	e.logger.Log(log.LevelDebug, "msg", "设置Pod链接（如果不存在）", "namespace", namespace, "uid", uid, "serviceName", serviceName, "podID", podID, "persistSeconds", persistSeconds)

	keys, args := e.createUidKeysAndArgs(namespace, uid, serviceName, podID, persistSeconds)

	result, err := e.executeScript(ctx, "statefulSetLinkIfAbsent", keys, args)
	if err != nil {
		return -1, fmt.Errorf("执行SetLinkIfAbsent脚本失败: %w", err)
	}

	return e.parseIntResult(result)
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

// GetLinkedPodIfPersist 仅在链接仍然持久时获取链接的Pod
func (e *StatefulExecutorImpl) GetLinkedPodIfPersist(ctx context.Context, namespace, uid, serviceName string) (int, error) {
	e.logger.Log(log.LevelDebug, "msg", "获取持久链接的Pod", "namespace", namespace, "uid", uid, "serviceName", serviceName)

	uidSvcKey := e.formatUidWithSpecificSvcLinkRedisKey(namespace, serviceName, uid)

	result, err := e.executeScript(ctx, "statefulGetLinkIfPersist", []string{uidSvcKey}, []interface{}{})
	if err != nil {
		return -1, fmt.Errorf("执行GetLinkIfPersist脚本失败: %w", err)
	}

	return e.parseIntResult(result)
}

// BatchGetLinkedPod 批量获取多个UID的链接Pod
func (e *StatefulExecutorImpl) BatchGetLinkedPod(ctx context.Context, namespace string, keys []string, serviceName string) (map[int][]string, error) {
	e.logger.Log(log.LevelDebug, "msg", "批量获取链接的Pod", "namespace", namespace, "keys", keys, "serviceName", serviceName)

	retMap := make(map[int][]string)
	if len(keys) == 0 {
		e.logger.Log(log.LevelDebug, "msg", "batchGetLinkedPod keys empty")
		return retMap, nil
	}

	// 构建所有键
	uidSvcKeys := make([]string, len(keys))
	for i, k := range keys {
		uidSvcKeys[i] = e.formatUidWithSpecificSvcLinkRedisKey(namespace, serviceName, k)
	}

	// 批量获取
	result := e.redisClient.MGet(ctx, uidSvcKeys...)
	if result.Err() != nil {
		return nil, fmt.Errorf("批量获取Pod链接失败: %w", result.Err())
	}

	if len(result.Val()) != len(keys) {
		return nil, fmt.Errorf("batchGetLinkedPod 返回的数量:%d和传入:%d不相等", len(result.Val()), len(keys))
	}

	// 处理结果
	for i, val := range result.Val() {
		if val != nil && val != "" && val != "null" && val != "nil" {
			if podIndex, err := strconv.Atoi(val.(string)); err == nil {
				linkID := keys[i]
				retMap[podIndex] = append(retMap[podIndex], linkID)
			}
		}
	}

	return retMap, nil
}

// RemoveLinkedPod 移除Pod链接，同时可以选择维持临时状态
func (e *StatefulExecutorImpl) RemoveLinkedPod(ctx context.Context, namespace, uid, serviceName string, persistSeconds int) (bool, error) {
	e.logger.Log(log.LevelDebug, "msg", "移除Pod链接", "namespace", namespace, "uid", uid, "serviceName", serviceName, "persistSeconds", persistSeconds)

	uidSvcKey := e.formatUidWithSpecificSvcLinkRedisKey(namespace, serviceName, uid)
	uidKey := e.formatUidLinkRedisKey(namespace, uid)

	keys := []string{uidSvcKey, uidKey}
	args := []interface{}{serviceName, persistSeconds}

	result, err := e.executeScript(ctx, "statefulRemoveLink", keys, args)
	if err != nil {
		return false, fmt.Errorf("执行RemoveLink脚本失败: %w", err)
	}

	return e.parseBoolResult(result)
}

// RemoveLinkedPodWithId 如果匹配指定的Pod ID，则移除Pod链接
func (e *StatefulExecutorImpl) RemoveLinkedPodWithId(ctx context.Context, namespace, uid, serviceName string, persistSeconds, podID int) (bool, error) {
	e.logger.Log(log.LevelDebug, "msg", "移除指定Pod ID的链接", "namespace", namespace, "uid", uid, "serviceName", serviceName, "persistSeconds", persistSeconds, "podID", podID)

	uidSvcKey := e.formatUidWithSpecificSvcLinkRedisKey(namespace, serviceName, uid)
	uidKey := e.formatUidLinkRedisKey(namespace, uid)

	keys := []string{uidSvcKey, uidKey}
	args := []interface{}{serviceName, persistSeconds, podID}

	result, err := e.executeScript(ctx, "statefulRemoveLinkWithId", keys, args)
	if err != nil {
		return false, fmt.Errorf("执行RemoveLinkWithId脚本失败: %w", err)
	}

	return e.parseBoolResult(result)
}

// GetLinkService 获取特定UID链接的所有服务
func (e *StatefulExecutorImpl) GetLinkService(ctx context.Context, namespace, uid string) (map[string]int, error) {
	e.logger.Log(log.LevelDebug, "msg", "获取链接的服务", "namespace", namespace, "uid", uid)

	uidKey := e.formatUidLinkRedisKey(namespace, uid)
	keys := []string{uidKey}
	args := []interface{}{}

	result, err := e.executeScript(ctx, "statefulGetLinkService", keys, args)
	if err != nil {
		return nil, fmt.Errorf("执行GetLinkService脚本失败: %w", err)
	}

	return e.parseLinkServiceResult(result)
}

// ==================== 辅助方法 ====================

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

// parseTrySetLinkResult 解析TrySetLink结果
func (e *StatefulExecutorImpl) parseTrySetLinkResult(result interface{}, podID int) (bool, int, error) {
	if resultList, ok := result.([]interface{}); ok && len(resultList) >= 2 {
		if currentPodStr, ok := resultList[0].(string); ok {
			if currentPodID, err := strconv.Atoi(currentPodStr); err == nil {
				if currentPodID == podID {
					return true, currentPodID, nil
				} else {
					return false, currentPodID, nil
				}
			}
		}
	}
	return false, -1, fmt.Errorf("无效的脚本结果")
}

// parseBoolResult 解析布尔结果
func (e *StatefulExecutorImpl) parseBoolResult(result interface{}) (bool, error) {
	if resultList, ok := result.([]interface{}); ok && len(resultList) > 0 {
		if val, ok := resultList[0].(int64); ok {
			return val == 1, nil
		}
	}
	return false, fmt.Errorf("无效的脚本结果")
}

// parseLinkServiceResult 解析链接服务结果
func (e *StatefulExecutorImpl) parseLinkServiceResult(result interface{}) (map[string]int, error) {
	if list, ok := result.([]interface{}); ok {
		resultMap := make(map[string]int, 8)
		for i := 0; i < len(list); i += 2 {
			if i+1 < len(list) {
				if serviceName, ok := list[i].(string); ok {
					if podIDStr, ok := list[i+1].(string); ok {
						if podID, err := strconv.Atoi(podIDStr); err == nil {
							resultMap[serviceName] = podID
						}
					}
				}
			}
		}
		return resultMap, nil
	}
	return nil, fmt.Errorf("无效的脚本结果")
}

// ==================== Redis键格式化方法 ====================

// formatUidLinkRedisKey 格式化UID链接信息的Redis键
// 格式："sf:{<namespace>:lk:<uid>}"
func (e *StatefulExecutorImpl) formatUidLinkRedisKey(namespace, uid string) string {
	return fmt.Sprintf("sf:{%s:lk:%s}", namespace, uid)
}

// formatUidWithSpecificSvcLinkRedisKey 格式化UID与特定服务链接的Redis键
// 格式："sf:{<namespace>:lk:<uid>}<serviceName>"
func (e *StatefulExecutorImpl) formatUidWithSpecificSvcLinkRedisKey(namespace, serviceName, uid string) string {
	return fmt.Sprintf("sf:{%s:lk:%s}%s", namespace, uid, serviceName)
}

// formatServiceStateRedisKey 格式化服务状态信息的Redis键
// 格式："sf:{<namespace>:state:<serviceName>}"
func (e *StatefulExecutorImpl) formatServiceStateRedisKey(namespace, serviceName string) string {
	return fmt.Sprintf("sf:{%s:state:%s}", namespace, serviceName)
}

// formatWorkloadStateRedisKey 格式化工作负载状态信息的Redis键
// 格式："sf:{<namespace>:workload:<serviceName>}"
func (e *StatefulExecutorImpl) formatWorkloadStateRedisKey(namespace, serviceName string) string {
	return fmt.Sprintf("sf:{%s:workload:%s}", namespace, serviceName)
}

// ==================== 参数验证方法 ====================

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

// validateServiceParams 验证服务参数
func (e *StatefulExecutorImpl) validateServiceParams(namespace, serviceName string) error {
	if strings.TrimSpace(namespace) == "" {
		return fmt.Errorf("namespace不能为空")
	}
	if strings.TrimSpace(serviceName) == "" {
		return fmt.Errorf("serviceName不能为空")
	}
	return nil
}

// validateUidParams 验证UID参数
func (e *StatefulExecutorImpl) validateUidParams(uid, namespace, serviceName string, podID int) error {
	if strings.TrimSpace(uid) == "" {
		return fmt.Errorf("uid不能为空")
	}
	if strings.TrimSpace(namespace) != "" && strings.TrimSpace(serviceName) != "" && podID >= 0 {
		// 如果提供了完整的服务信息，验证它们
		if err := e.validateServiceStateParams(namespace, serviceName, podID, "dummy"); err != nil {
			return err
		}
	}
	return nil
}
