// Package serverinfo 提供服务器信息管理功能
package serverinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockProvider 模拟 Provider 实现
type MockProvider struct {
	name     string
	priority int
}

func (m *MockProvider) GetName() string {
	return m.name
}

func (m *MockProvider) GetPriority() int {
	return m.priority
}

func (m *MockProvider) CanProvide(field string) bool {
	return field == "ServiceName" || field == "Version"
}

func (m *MockProvider) Provide(field string) (string, error) {
	switch field {
	case "ServiceName":
		return "test-service", nil
	case "Version":
		return "1.0.0", nil
	default:
		return "", nil
	}
}

func TestProvideServerInfo(t *testing.T) {
	// 测试 ProvideServerInfo 函数

	// 创建注册表并注册测试 Provider
	registry := NewProviderRegistry()
	mockProvider := &MockProvider{name: "test", priority: 1}
	registry.RegisterProvider(mockProvider)

	// 测试 ProvideServerInfo
	info, err := ProvideServerInfo()

	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestProvideServerInfoEmptyRegistry(t *testing.T) {
	// 测试空注册表的情况

	// 创建空的注册表
	registry := NewProviderRegistry()

	// 测试空注册表构建 ServerInfo
	info, err := registry.BuildServerInfo()

	// 空注册表应该能构建出空的 ServerInfo
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestProvideServerInfoWithMultipleProviders(t *testing.T) {
	// 测试多个 Provider 的情况

	// 创建注册表并注册多个测试 Provider
	registry := NewProviderRegistry()
	registry.RegisterProvider(&MockProvider{name: "provider1", priority: 1})
	registry.RegisterProvider(&MockProvider{name: "provider2", priority: 2})
	registry.RegisterProvider(&MockProvider{name: "provider3", priority: 3})

	// 测试构建 ServerInfo
	info, err := registry.BuildServerInfo()

	assert.NoError(t, err)
	assert.NotNil(t, info)
}
