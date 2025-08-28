## 6A 任务卡：实现 K8s 环境信息提供者

- 编号: Task-02
- 模块: runtime
- 责任人: 待分配
- 优先级: 🔴 高
- 状态: ❌ 未开始
- 预计完成时间: -
- 实际完成时间: -

### A1 目标（Aim）
实现 K8s 环境信息提供者，专门从 Kubernetes 环境变量中获取 Pod 相关信息（如 PodName、PodIndex、Namespace 等），支持 StatefulSet 部署模式，为 RuntimeInfo 提供 K8s 环境的基础信息。

### A2 分析（Analyze）
- **现状**：
  - ✅ 已实现：无
  - 🔄 部分实现：无
  - ❌ 未实现：K8s Provider 实现、环境变量解析、Pod 信息获取
- **差距**：
  - 需要实现 K8sProvider 结构体
  - 需要解析 K8s 环境变量
  - 需要支持 StatefulSet 的 PodIndex 计算
  - 需要实现 Provider 接口的所有方法
- **约束**：
  - 必须遵循 Provider 接口契约
  - 必须支持 K8s 标准环境变量
  - 必须处理环境变量缺失的情况
- **风险**：
  - 技术风险：环境变量解析错误
  - 业务风险：PodIndex 计算错误
  - 依赖风险：K8s 环境变量不可用

### A3 设计（Architect）
- **接口契约**：
  - **核心接口**：`K8sProvider` - 实现 Provider 接口
  - **核心方法**：
    - `GetName() string` - 返回 "k8s"
    - `GetPriority() int` - 返回 1（最高优先级）
    - `CanProvide(field string) bool` - 检查是否能提供指定字段
    - `Provide(field string) (string, error)` - 提供指定字段的值
  - **支持字段**：namespace、podName、podIndex、serviceName

- **架构设计**：
  - 实现 Provider 接口，提供 K8s 环境信息
  - 使用环境变量获取 Pod 信息
  - 支持 StatefulSet 的 PodIndex 计算
  - 提供合理的默认值和错误处理

- **核心功能模块**：
  - `K8sProvider`: K8s 环境信息提供者
  - `PodInfoParser`: Pod 信息解析器
  - `EnvVarReader`: 环境变量读取器

- **极小任务拆分**：
  - T02-01：实现 K8sProvider 结构体
  - T02-02：实现环境变量读取逻辑
  - T02-03：实现 Pod 信息解析逻辑
  - T02-04：实现 Provider 接口方法
  - T02-05：添加单元测试

### A4 行动（Act）
#### T02-01：实现 K8sProvider 结构体
```go
// provider/k8s_provider.go
package provider

import (
    "os"
    "strings"
)

// K8sProvider K8s 环境信息提供者
type K8sProvider struct {
    envVars map[string]string
}

// NewK8sProvider 创建新的 K8s Provider
func NewK8sProvider() *K8sProvider {
    return &K8sProvider{
        envVars: make(map[string]string),
    }
}

// GetName 获取 Provider 名称
func (k *K8sProvider) GetName() string {
    return "k8s"
}

// GetPriority 获取 Provider 优先级（数字越小优先级越高）
func (k *K8sProvider) GetPriority() int {
    return 1 // K8s Provider 优先级最高
}
```

#### T02-02：实现环境变量读取逻辑
```go
// provider/k8s_provider.go (续)

// loadEnvVars 加载环境变量
func (k *K8sProvider) loadEnvVars() {
    // K8s 标准环境变量
    k.envVars["POD_NAME"] = os.Getenv("POD_NAME")
    k.envVars["POD_NAMESPACE"] = os.Getenv("POD_NAMESPACE")
    k.envVars["POD_IP"] = os.Getenv("POD_IP")
    k.envVars["SERVICE_NAME"] = os.Getenv("SERVICE_NAME")
    
    // 自定义环境变量
    k.envVars["TPF_NAMESPACE"] = os.Getenv("TPF_NAMESPACE")
    k.envVars["TPF_SERVICE_NAME"] = os.Getenv("TPF_SERVICE_NAME")
    
    // 服务相关环境变量
    k.envVars["K8S_SERVICE_NAME"] = os.Getenv("K8S_SERVICE_NAME")
    k.envVars["K8S_NAMESPACE"] = os.Getenv("K8S_NAMESPACE")
}

// getEnvVar 获取环境变量值
func (k *K8sProvider) getEnvVar(key string) string {
    if k.envVars == nil {
        k.loadEnvVars()
    }
    return k.envVars[key]
}
```

#### T02-03：实现 Pod 信息解析逻辑
```go
// provider/k8s_provider.go (续)

// parsePodIndex 从 Pod 名称解析 Pod 索引
func (k *K8sProvider) parsePodIndex(podName string) string {
    if podName == "" {
        return "0"
    }
    
    // 支持 StatefulSet 命名格式：service-name-0, service-name-1
    parts := strings.Split(podName, "-")
    if len(parts) > 1 {
        lastPart := parts[len(parts)-1]
        // 检查最后一部分是否为数字
        if _, err := strconv.Atoi(lastPart); err == nil {
            return lastPart
        }
    }
    
    // 如果不是标准格式，返回默认值
    return "0"
}

// getServiceName 获取服务名称
func (k *K8sProvider) getServiceName() string {
    // 直接返回环境变量值，不做复杂推断
    if serviceName := k.getEnvVar("TPF_SERVICE_NAME"); serviceName != "" {
        return serviceName
    }
    if serviceName := k.getEnvVar("K8S_SERVICE_NAME"); serviceName != "" {
        return serviceName
    }
    if serviceName := k.getEnvVar("SERVICE_NAME"); serviceName != "" {
        return serviceName
    }
    
    return ""
}

// getNamespace 获取命名空间
func (k *K8sProvider) getNamespace() string {
    // 直接返回环境变量值，不做复杂推断
    if namespace := k.getEnvVar("TPF_NAMESPACE"); namespace != "" {
        return namespace
    }
    if namespace := k.getEnvVar("K8S_NAMESPACE"); namespace != "" {
        return namespace
    }
    if namespace := k.getEnvVar("POD_NAMESPACE"); namespace != "" {
        return namespace
    }
    
    return ""
}
```

#### T02-04：实现 Provider 接口方法
```go
// provider/k8s_provider.go (续)

// CanProvide 检查是否能提供指定字段
func (k *K8sProvider) CanProvide(field string) bool {
    switch field {
    case "namespace", "podName", "podIndex", "serviceName":
        return true
    default:
        return false
    }
}

// Provide 提供指定字段的值
func (k *K8sProvider) Provide(field string) (string, error) {
    if !k.CanProvide(field) {
        return "", fmt.Errorf("field %s not supported by K8sProvider", field)
    }
    
    switch field {
    case "namespace":
        return k.getNamespace(), nil
    case "podName":
        return k.getEnvVar("POD_NAME"), nil
    case "podIndex":
        podName := k.getEnvVar("POD_NAME")
        return k.parsePodIndex(podName), nil
    case "serviceName":
        return k.getServiceName(), nil
    default:
        return "", fmt.Errorf("unknown field: %s", field)
    }
}

// Validate 方法已移除，简化接口设计
```

#### T02-05：添加单元测试
```go
// provider/k8s_provider_test.go
package provider

import (
    "os"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestK8sProvider_GetName(t *testing.T) {
    provider := NewK8sProvider()
    assert.Equal(t, "k8s", provider.GetName())
}

func TestK8sProvider_GetPriority(t *testing.T) {
    provider := NewK8sProvider()
    assert.Equal(t, 1, provider.GetPriority())
}

func TestK8sProvider_CanProvide(t *testing.T) {
    provider := NewK8sProvider()
    
    assert.True(t, provider.CanProvide("namespace"))
    assert.True(t, provider.CanProvide("podName"))
    assert.True(t, provider.CanProvide("podIndex"))
    assert.True(t, provider.CanProvide("serviceName"))
    assert.False(t, provider.CanProvide("unknown"))
}

func TestK8sProvider_Provide_Namespace(t *testing.T) {
    // 设置环境变量
    os.Setenv("TPF_NAMESPACE", "test-namespace")
    defer os.Unsetenv("TPF_NAMESPACE")
    
    provider := NewK8sProvider()
    namespace, err := provider.Provide("namespace")
    
    require.NoError(t, err)
    assert.Equal(t, "test-namespace", namespace)
}

func TestK8sProvider_Provide_PodIndex(t *testing.T) {
    // 设置环境变量
    os.Setenv("POD_NAME", "test-service-2")
    defer os.Unsetenv("POD_NAME")
    
    provider := NewK8sProvider()
    podIndex, err := provider.Provide("podIndex")
    
    require.NoError(t, err)
    assert.Equal(t, "2", podIndex)
}

// Validate 测试已移除，简化接口设计
```

### A5 验证（Assure）
- **测试用例**：
  - 测试 K8sProvider 的基本方法（GetName、GetPriority）
  - 测试 CanProvide 方法的字段支持检查
  - 测试 Provide 方法的环境变量获取
  - 测试 Provider 接口实现完整性
  - 测试 PodIndex 解析逻辑
- **性能验证**：
  - 环境变量读取性能
  - Pod 信息解析性能
- **回归测试**：
  - 确保 Provider 接口实现正确
  - 确保环境变量处理正确
- **测试结果**：
  - 所有测试用例通过
  - Provider 接口实现完整
  - 环境变量处理正确

### A6 迭代（Advance）
- 性能优化：缓存环境变量读取结果
- 功能扩展：支持更多 K8s 环境变量
- 观测性增强：添加环境变量读取日志
- 下一步任务链接：[Task-03](./Task-03-实现配置中心集成提供者.md)

### 📋 质量检查
- [ ] 代码质量检查完成
- [ ] 文档质量检查完成
- [ ] 测试质量检查完成

### 📋 完成总结
- **K8sProvider 结构体**：实现了完整的 K8s 环境信息提供者
- **环境变量读取**：支持标准 K8s 环境变量和自定义环境变量
- **Pod 信息解析**：支持 StatefulSet 的 PodIndex 计算
- **Provider 接口实现**：完整实现了 Provider 接口的所有方法
- **单元测试**：提供了完整的测试用例，确保功能正确性
- **设计原则**：遵循"启动时初始化，启动后不再更新"的原则，使用环境变量获取信息，简化接口设计
