// Package provider 提供服务器信息提供者
package provider

import (
	"os"
	"strings"
)

// ConfigProvider 配置中心集成提供者
type ConfigProvider struct {
	name     string
	priority int
}

// NewConfigProvider 创建新的配置提供者
func NewConfigProvider() *ConfigProvider {
	return &ConfigProvider{
		name:     "config",
		priority: 2, // 中等优先级
	}
}

// GetName 返回提供者名称
func (cp *ConfigProvider) GetName() string {
	return cp.name
}

// GetPriority 返回提供者优先级
func (cp *ConfigProvider) GetPriority() int {
	return cp.priority
}

// CanProvide 检查是否可以提供指定字段
func (cp *ConfigProvider) CanProvide(field string) bool {
	switch field {
	case "AppId", "ArtifactId", "RegionId", "ChannelId":
		return true
	default:
		return false
	}
}

// Provide 提供指定字段的值
func (cp *ConfigProvider) Provide(field string) (string, error) {
	switch field {
	case "AppId":
		return cp.getAppId(), nil
	case "ArtifactId":
		return cp.getArtifactId(), nil
	case "RegionId":
		return cp.getRegionId(), nil
	case "ChannelId":
		return cp.getChannelId(), nil
	default:
		return "", nil
	}
}

// getAppId 获取应用 ID
func (cp *ConfigProvider) getAppId() string {
	// 从环境变量获取
	if appId := os.Getenv("APP_ID"); appId != "" {
		return appId
	}

	if appId := os.Getenv("APPLICATION_ID"); appId != "" {
		return appId
	}

	// 从配置中心获取（这里简化实现，实际应该调用配置中心）
	if appId := cp.getConfigCenterValue("app.id"); appId != "" {
		return appId
	}

	// 默认值
	return "unknown"
}

// getArtifactId 获取制品 ID
func (cp *ConfigProvider) getArtifactId() string {
	// 从环境变量获取
	if artifactId := os.Getenv("ARTIFACT_ID"); artifactId != "" {
		return artifactId
	}

	if artifactId := os.Getenv("PROJECT_ID"); artifactId != "" {
		return artifactId
	}

	// 从配置中心获取
	if artifactId := cp.getConfigCenterValue("artifact.id"); artifactId != "" {
		return artifactId
	}

	// 默认值
	return "unknown"
}

// getRegionId 获取区域 ID
func (cp *ConfigProvider) getRegionId() string {
	// 从环境变量获取
	if regionId := os.Getenv("REGION_ID"); regionId != "" {
		return regionId
	}

	if regionId := os.Getenv("REGION"); regionId != "" {
		return regionId
	}

	if regionId := os.Getenv("ZONE"); regionId != "" {
		return regionId
	}

	// 从配置中心获取
	if regionId := cp.getConfigCenterValue("region.id"); regionId != "" {
		return regionId
	}

	// 默认值
	return "unknown"
}

// getChannelId 获取渠道 ID
func (cp *ConfigProvider) getChannelId() string {
	// 从环境变量获取
	if channelId := os.Getenv("CHANNEL_ID"); channelId != "" {
		return channelId
	}

	if channelId := os.Getenv("CHANNEL"); channelId != "" {
		return channelId
	}

	// 从配置中心获取
	if channelId := cp.getConfigCenterValue("channel.id"); channelId != "" {
		return channelId
	}

	// 默认值
	return "unknown"
}

// getConfigCenterValue 从配置中心获取值
func (cp *ConfigProvider) getConfigCenterValue(key string) string {
	// 这里简化实现，实际应该调用配置中心（如 Apollo、Nacos 等）
	// 可以通过环境变量模拟配置中心的值

	// 模拟配置中心配置
	configMap := map[string]string{
		"app.id":       os.Getenv("CONFIG_APP_ID"),
		"artifact.id":  os.Getenv("CONFIG_ARTIFACT_ID"),
		"region.id":    os.Getenv("CONFIG_REGION_ID"),
		"channel.id":   os.Getenv("CONFIG_CHANNEL_ID"),
		"service.name": os.Getenv("CONFIG_SERVICE_NAME"),
		"version":      os.Getenv("CONFIG_VERSION"),
	}

	return configMap[key]
}

// IsConfigCenterAvailable 检查配置中心是否可用
func (cp *ConfigProvider) IsConfigCenterAvailable() bool {
	// 检查配置中心相关的环境变量
	configVars := []string{
		"CONFIG_CENTER_URL",
		"APOLLO_META",
		"NACOS_SERVER_ADDR",
		"CONFIG_APP_ID",
		"CONFIG_ARTIFACT_ID",
		"CONFIG_REGION_ID",
		"CONFIG_CHANNEL_ID",
	}

	for _, envVar := range configVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// GetConfigInfo 获取配置信息
func (cp *ConfigProvider) GetConfigInfo() map[string]string {
	info := make(map[string]string)

	// 基础配置信息
	info["app_id"] = cp.getAppId()
	info["artifact_id"] = cp.getArtifactId()
	info["region_id"] = cp.getRegionId()
	info["channel_id"] = cp.getChannelId()

	// 配置中心状态
	info["config_center_available"] = "false"
	if cp.IsConfigCenterAvailable() {
		info["config_center_available"] = "true"
	}

	// 环境变量信息
	envVars := []string{
		"CONFIG_CENTER_URL",
		"APOLLO_META",
		"NACOS_SERVER_ADDR",
		"CONFIG_APP_ID",
		"CONFIG_ARTIFACT_ID",
		"CONFIG_REGION_ID",
		"CONFIG_CHANNEL_ID",
		"CONFIG_SERVICE_NAME",
		"CONFIG_VERSION",
	}

	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			key := strings.ToLower(envVar)
			info[key] = value
		}
	}

	return info
}

// GetConfigCenterType 获取配置中心类型
func (cp *ConfigProvider) GetConfigCenterType() string {
	if os.Getenv("APOLLO_META") != "" {
		return "apollo"
	}

	if os.Getenv("NACOS_SERVER_ADDR") != "" {
		return "nacos"
	}

	if os.Getenv("CONFIG_CENTER_URL") != "" {
		return "custom"
	}

	return "none"
}
