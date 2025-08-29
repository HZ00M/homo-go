// Package serverinfo 提供服务器信息管理功能
package serverinfo

import (
	"sync"
)

// 全局 ProviderRegistry 实例
var (
	globalRegistry *ProviderRegistry
	globalOnce     sync.Once
)

// GetGlobalRegistry 获取全局 ProviderRegistry 实例
func GetGlobalRegistry() *ProviderRegistry {
	globalOnce.Do(func() {
		globalRegistry = NewProviderRegistry()
	})
	return globalRegistry
}

// RegisterGlobalProvider 全局注册 Provider 的便捷函数
func RegisterGlobalProvider(provider Provider) {
	GetGlobalRegistry().RegisterProvider(provider)
}

// Get 获取全局唯一的 ServerInfo 实例
func Get() *ServerInfo {
	// 从全局注册表构建 ServerInfo
	if globalInfo, err := GetGlobalRegistry().BuildServerInfo(); err == nil {
		return globalInfo
	}
	// 如果构建失败，返回空的 ServerInfo 实例
	return NewServerInfo()
}

// BuildGlobalServerInfo 从全局注册表构建 ServerInfo 的便捷函数（保持兼容性）
func BuildGlobalServerInfo() (*ServerInfo, error) {
	return GetGlobalRegistry().BuildServerInfo()
}

// BuildGlobalServerInfoWithFields 从全局注册表构建指定字段的 ServerInfo 的便捷函数（保持兼容性）
func BuildGlobalServerInfoWithFields(fields map[string]string) (*ServerInfo, error) {
	return GetGlobalRegistry().BuildServerInfoWithFields(fields)
}

// ProviderRegistry Provider 注册表
type ProviderRegistry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewProviderRegistry 创建新的 Provider 注册表
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider 注册 Provider
func (r *ProviderRegistry) RegisterProvider(provider Provider) {
	if provider == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.GetName()] = provider
}

// GetProvider 获取指定 Provider
func (r *ProviderRegistry) GetProvider(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, exists := r.providers[name]
	return provider, exists
}

// GetAllProviders 获取所有 Provider
func (r *ProviderRegistry) GetAllProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}

// RemoveProvider 移除指定 Provider
func (r *ProviderRegistry) RemoveProvider(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
}

// ClearProviders 清空所有 Provider
func (r *ProviderRegistry) ClearProviders() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = make(map[string]Provider)
}

// GetProviderCount 获取 Provider 数量
func (r *ProviderRegistry) GetProviderCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// HasProvider 检查是否存在指定 Provider
func (r *ProviderRegistry) HasProvider(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.providers[name]
	return exists
}

// GetProvidersByPriority 按优先级获取 Provider 列表
func (r *ProviderRegistry) GetProvidersByPriority() []Provider {
	providers := r.GetAllProviders()

	// 按优先级排序（数字越小优先级越高）
	for i := 0; i < len(providers)-1; i++ {
		for j := i + 1; j < len(providers); j++ {
			if providers[i].GetPriority() > providers[j].GetPriority() {
				providers[i], providers[j] = providers[j], providers[i]
			}
		}
	}

	return providers
}

// BuildServerInfo 从注册表构建 ServerInfo
func (r *ProviderRegistry) BuildServerInfo() (*ServerInfo, error) {
	info := NewServerInfo()

	// 按优先级获取 Provider
	providers := r.GetProvidersByPriority()

	// 从每个 Provider 获取可提供的字段
	for _, provider := range providers {
		if provider == nil {
			continue
		}

		// 尝试获取所有可能的字段
		fields := []string{"ServiceName", "Version", "Namespace", "PodName", "PodIndex", "AppId", "ArtifactId", "RegionId", "ChannelId"}
		for _, field := range fields {
			if provider.CanProvide(field) {
				if value, err := provider.Provide(field); err == nil && value != "" {
					switch field {
					case "ServiceName":
						info.setServiceName(value)
					case "Version":
						info.setVersion(value)
					case "Namespace":
						info.setNamespace(value)
					case "PodName":
						info.setPodName(value)
					case "PodIndex":
						info.setPodIndex(value)
					case "AppId":
						info.setAppId(value)
					case "ArtifactId":
						info.setArtifactId(value)
					case "RegionId":
						info.setRegionId(value)
					case "ChannelId":
						info.setChannelId(value)
					default:
						info.setMetadata(field, value)
					}
				}
			}
		}
	}

	return info, nil
}

// BuildServerInfoWithFields 从注册表构建 ServerInfo 并设置额外字段
func (r *ProviderRegistry) BuildServerInfoWithFields(fields map[string]string) (*ServerInfo, error) {
	info, err := r.BuildServerInfo()
	if err != nil {
		return nil, err
	}

	// 设置额外字段（会覆盖 Provider 提供的值）
	for field, value := range fields {
		switch field {
		case "ServiceName":
			info.setServiceName(value)
		case "Version":
			info.setVersion(value)
		case "Namespace":
			info.setNamespace(value)
		case "PodName":
			info.setPodName(value)
		case "PodIndex":
			info.setPodIndex(value)
		case "AppId":
			info.setAppId(value)
		case "ArtifactId":
			info.setArtifactId(value)
		case "RegionId":
			info.setRegionId(value)
		case "ChannelId":
			info.setChannelId(value)
		default:
			info.setMetadata(field, value)
		}
	}

	return info, nil
}

// 便捷函数
func NewServerInfoFromProviders(providers ...Provider) (*ServerInfo, error) {
	registry := NewProviderRegistry()

	// 注册所有 Provider
	for _, provider := range providers {
		registry.RegisterProvider(provider)
	}

	return registry.BuildServerInfo()
}
