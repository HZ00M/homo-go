// Package provider 提供服务器信息提供者
package provider

import (
	"os"
	"runtime"
	"strings"
)

// LocalProvider 本地环境信息提供者
type LocalProvider struct {
	name     string
	priority int
}

// NewLocalProvider 创建新的本地提供者
func NewLocalProvider() *LocalProvider {
	return &LocalProvider{
		name:     "local",
		priority: 3, // 较低优先级，作为后备
	}
}

// GetName 返回提供者名称
func (lp *LocalProvider) GetName() string {
	return lp.name
}

// GetPriority 返回提供者优先级
func (lp *LocalProvider) GetPriority() int {
	return lp.priority
}

// CanProvide 检查是否可以提供指定字段
func (lp *LocalProvider) CanProvide(field string) bool {
	switch field {
	case "Namespace", "PodName", "PodIndex":
		return true
	default:
		return false
	}
}

// Provide 提供指定字段的值
func (lp *LocalProvider) Provide(field string) (string, error) {
	switch field {
	case "Namespace":
		return lp.getNamespace(), nil
	case "PodName":
		return lp.getPodName(), nil
	case "PodIndex":
		return lp.getPodIndex(), nil
	default:
		return "", nil
	}
}

// getNamespace 获取命名空间
func (lp *LocalProvider) getNamespace() string {
	// 从环境变量获取
	if ns := os.Getenv("NAMESPACE"); ns != "" {
		return ns
	}

	if ns := os.Getenv("ENVIRONMENT"); ns != "" {
		return ns
	}

	// 默认本地环境
	return "local"
}

// getPodName 获取 Pod 名称
func (lp *LocalProvider) getPodName() string {
	// 从环境变量获取
	if podName := os.Getenv("POD_NAME"); podName != "" {
		return podName
	}

	if podName := os.Getenv("HOSTNAME"); podName != "" {
		return podName
	}

	// 从主机名获取
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return hostname
	}

	// 默认值
	return "localhost"
}

// getPodIndex 获取 Pod 索引
func (lp *LocalProvider) getPodIndex() string {
	// 本地环境通常没有 Pod 索引
	return ""
}

// IsLocalEnvironment 检查是否为本地环境
func (lp *LocalProvider) IsLocalEnvironment() bool {
	// 检查本地环境相关的环境变量
	localVars := []string{
		"LOCAL_DEV",
		"DEVELOPMENT",
		"DEBUG",
	}

	for _, envVar := range localVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	// 检查是否为本地开发环境
	namespace := lp.getNamespace()
	return namespace == "local" || namespace == "dev" || namespace == "development"
}

// GetLocalInfo 获取本地环境信息
func (lp *LocalProvider) GetLocalInfo() map[string]string {
	info := make(map[string]string)

	// 基础信息
	info["namespace"] = lp.getNamespace()
	info["pod_name"] = lp.getPodName()
	info["pod_index"] = lp.getPodIndex()

	// 系统信息
	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH
	info["go_version"] = runtime.Version()

	// 环境变量信息
	envVars := []string{
		"USER",
		"HOME",
		"PWD",
		"PATH",
		"GOPATH",
		"GOROOT",
	}

	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			key := strings.ToLower(envVar)
			info[key] = value
		}
	}

	return info
}

// GetHostname 获取主机名
func (lp *LocalProvider) GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// GetWorkingDirectory 获取工作目录
func (lp *LocalProvider) GetWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

// GetEnvironmentVariables 获取环境变量
func (lp *LocalProvider) GetEnvironmentVariables() map[string]string {
	env := make(map[string]string)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}

	return env
}
