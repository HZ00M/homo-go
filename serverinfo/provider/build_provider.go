// Package provider 提供服务器信息提供者
package provider

import (
	"os"
	"runtime"
	"strings"
	"time"
)

// BuildProvider 构建信息提供者
type BuildProvider struct {
	name     string
	priority int
}

// NewBuildProvider 创建新的构建提供者
func NewBuildProvider() *BuildProvider {
	return &BuildProvider{
		name:     "build",
		priority: 4, // 较低优先级
	}
}

// GetName 返回提供者名称
func (bp *BuildProvider) GetName() string {
	return bp.name
}

// GetPriority 返回提供者优先级
func (bp *BuildProvider) GetPriority() int {
	return bp.priority
}

// CanProvide 检查是否可以提供指定字段
func (bp *BuildProvider) CanProvide(field string) bool {
	switch field {
	case "Version", "ServiceName":
		return true
	default:
		return false
	}
}

// Provide 提供指定字段的值
func (bp *BuildProvider) Provide(field string) (string, error) {
	switch field {
	case "Version":
		return bp.getVersion(), nil
	case "ServiceName":
		return bp.getServiceName(), nil
	default:
		return "", nil
	}
}

// getVersion 获取版本信息
func (bp *BuildProvider) getVersion() string {
	// 从环境变量获取
	if version := os.Getenv("VERSION"); version != "" {
		return version
	}

	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}

	if version := os.Getenv("BUILD_VERSION"); version != "" {
		return version
	}

	// 从构建信息获取（避免递归调用）
	if version := os.Getenv("BUILD_VERSION"); version != "" {
		return version
	}

	// 默认值
	return "1.0.0"
}

// getServiceName 获取服务名称
func (bp *BuildProvider) getServiceName() string {
	// 从环境变量获取
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		return serviceName
	}

	if serviceName := os.Getenv("APP_NAME"); serviceName != "" {
		return serviceName
	}

	// 从构建信息获取（避免递归调用）
	if serviceName := os.Getenv("BUILD_SERVICE_NAME"); serviceName != "" {
		return serviceName
	}

	// 默认值
	return "unknown-service"
}

// getBuildInfo 获取构建信息
func (bp *BuildProvider) getBuildInfo() map[string]string {
	info := make(map[string]string)

	// 基础构建信息
	info["version"] = bp.getVersion()
	info["service_name"] = bp.getServiceName()

	// 构建时间
	if buildTime := os.Getenv("BUILD_TIME"); buildTime != "" {
		info["build_time"] = buildTime
	} else {
		info["build_time"] = time.Now().Format(time.RFC3339)
	}

	// Git 信息
	if gitCommit := os.Getenv("GIT_COMMIT"); gitCommit != "" {
		info["git_commit"] = gitCommit
	}

	if gitBranch := os.Getenv("GIT_BRANCH"); gitBranch != "" {
		info["git_branch"] = gitBranch
	}

	if gitTag := os.Getenv("GIT_TAG"); gitTag != "" {
		info["git_tag"] = gitTag
	}

	// 构建环境信息
	if buildEnv := os.Getenv("BUILD_ENV"); buildEnv != "" {
		info["build_env"] = buildEnv
	}

	if buildNumber := os.Getenv("BUILD_NUMBER"); buildNumber != "" {
		info["build_number"] = buildNumber
	}

	// Go 运行时信息
	info["go_version"] = runtime.Version()
	info["go_os"] = runtime.GOOS
	info["go_arch"] = runtime.GOARCH

	// 编译时间
	info["compile_time"] = time.Now().Format(time.RFC3339)

	return info
}

// GetBuildTime 获取构建时间
func (bp *BuildProvider) GetBuildTime() time.Time {
	if buildTime := os.Getenv("BUILD_TIME"); buildTime != "" {
		if t, err := time.Parse(time.RFC3339, buildTime); err == nil {
			return t
		}
	}

	// 如果解析失败，返回当前时间
	return time.Now()
}

// GetGitInfo 获取 Git 信息
func (bp *BuildProvider) GetGitInfo() map[string]string {
	gitInfo := make(map[string]string)

	gitVars := []string{
		"GIT_COMMIT",
		"GIT_BRANCH",
		"GIT_TAG",
		"GIT_REPO",
		"GIT_AUTHOR",
		"GIT_COMMIT_TIME",
	}

	for _, envVar := range gitVars {
		if value := os.Getenv(envVar); value != "" {
			key := strings.ToLower(strings.Replace(envVar, "GIT_", "", 1))
			gitInfo[key] = value
		}
	}

	return gitInfo
}

// IsDevelopmentBuild 检查是否为开发构建
func (bp *BuildProvider) IsDevelopmentBuild() bool {
	buildEnv := os.Getenv("BUILD_ENV")
	return buildEnv == "dev" || buildEnv == "development" || buildEnv == "debug"
}

// IsProductionBuild 检查是否为生产构建
func (bp *BuildProvider) IsProductionBuild() bool {
	buildEnv := os.Getenv("BUILD_ENV")
	return buildEnv == "prod" || buildEnv == "production"
}

// GetBuildEnvironment 获取构建环境
func (bp *BuildProvider) GetBuildEnvironment() string {
	if buildEnv := os.Getenv("BUILD_ENV"); buildEnv != "" {
		return buildEnv
	}

	// 根据其他信息推断
	if bp.IsDevelopmentBuild() {
		return "dev"
	}

	if bp.IsProductionBuild() {
		return "prod"
	}

	return "unknown"
}

// GetBuildSummary 获取构建摘要
func (bp *BuildProvider) GetBuildSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	// 基础信息
	summary["version"] = bp.getVersion()
	summary["service_name"] = bp.getServiceName()
	summary["build_time"] = bp.GetBuildTime()
	summary["build_env"] = bp.GetBuildEnvironment()

	// 详细信息
	summary["build_info"] = bp.getBuildInfo()
	summary["git_info"] = bp.GetGitInfo()

	// 环境标识
	summary["is_development"] = bp.IsDevelopmentBuild()
	summary["is_production"] = bp.IsProductionBuild()

	return summary
}
