// Package provider 提供服务器信息提供者
package provider

import (
	"os"
	"strings"
)

// K8sProvider K8s 环境信息提供者
type K8sProvider struct {
	name     string
	priority int
}

// NewK8sProvider 创建新的 K8s 提供者
func NewK8sProvider() *K8sProvider {
	return &K8sProvider{
		name:     "k8s",
		priority: 1, // 最高优先级
	}
}

// GetName 返回提供者名称
func (kp *K8sProvider) GetName() string {
	return kp.name
}

// GetPriority 返回提供者优先级
func (kp *K8sProvider) GetPriority() int {
	return kp.priority
}

// CanProvide 检查是否可以提供指定字段
func (kp *K8sProvider) CanProvide(field string) bool {
	switch field {
	case "Namespace", "PodName", "PodIndex":
		return true
	default:
		return false
	}
}

// Provide 提供指定字段的值
func (kp *K8sProvider) Provide(field string) (string, error) {
	switch field {
	case "Namespace":
		return kp.getNamespace(), nil
	case "PodName":
		return kp.getPodName(), nil
	case "PodIndex":
		return kp.getPodIndex(), nil
	default:
		return "", nil
	}
}

// getNamespace 获取命名空间
func (kp *K8sProvider) getNamespace() string {
	// 优先从环境变量获取
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}

	// 从其他环境变量推断
	if ns := os.Getenv("NAMESPACE"); ns != "" {
		return ns
	}

	// 默认值
	return "default"
}

// getPodName 获取 Pod 名称
func (kp *K8sProvider) getPodName() string {
	// 优先从环境变量获取
	if podName := os.Getenv("POD_NAME"); podName != "" {
		return podName
	}

	// 从主机名获取
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return hostname
	}

	// 默认值
	return "unknown"
}

// getPodIndex 获取 Pod 索引
func (kp *K8sProvider) getPodIndex() string {
	podName := kp.getPodName()
	return kp.parsePodIndex(podName)
}

// parsePodIndex 解析 Pod 名称中的索引
func (kp *K8sProvider) parsePodIndex(podName string) string {
	if podName == "" {
		return ""
	}

	// 查找最后一个连字符后的数字
	parts := strings.Split(podName, "-")
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		// 检查是否为数字
		if strings.TrimSpace(lastPart) != "" && isNumeric(lastPart) {
			return lastPart
		}
	}

	// 查找包含数字的部分
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if isNumeric(part) {
			return part
		}
	}

	return ""
}

// isNumeric 检查字符串是否为数字
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// IsK8sEnvironment 检查是否为 K8s 环境
func (kp *K8sProvider) IsK8sEnvironment() bool {
	// 检查 K8s 相关的环境变量
	k8sVars := []string{
		"KUBERNETES_SERVICE_HOST",
		"KUBERNETES_SERVICE_PORT",
		"POD_NAME",
		"POD_NAMESPACE",
		"POD_IP",
	}

	for _, envVar := range k8sVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// GetK8sInfo 获取 K8s 环境信息
func (kp *K8sProvider) GetK8sInfo() map[string]string {
	info := make(map[string]string)

	// 基础信息
	info["namespace"] = kp.getNamespace()
	info["pod_name"] = kp.getPodName()
	info["pod_index"] = kp.getPodIndex()

	// 环境变量信息
	envVars := []string{
		"POD_IP",
		"KUBERNETES_SERVICE_HOST",
		"KUBERNETES_SERVICE_PORT",
		"NODE_NAME",
		"SERVICE_ACCOUNT_NAME",
	}

	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			key := strings.ToLower(envVar)
			info[key] = value
		}
	}

	return info
}
