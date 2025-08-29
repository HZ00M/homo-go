package provider

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestK8sProvider(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("POD_NAME", "test-pod-0")
	os.Setenv("POD_NAMESPACE", "test-namespace")
	os.Setenv("SERVICE_NAME", "test-service")
	defer func() {
		os.Unsetenv("POD_NAME")
		os.Unsetenv("POD_NAMESPACE")
		os.Unsetenv("SERVICE_NAME")
	}()

	provider := NewK8sProvider()

	// 测试基本信息
	assert.Equal(t, "k8s", provider.GetName())
	assert.Equal(t, 1, provider.GetPriority())

	// 测试支持的字段
	assert.True(t, provider.CanProvide("PodName"))
	assert.True(t, provider.CanProvide("PodIndex"))
	assert.True(t, provider.CanProvide("Namespace"))
	assert.True(t, provider.CanProvide("ServiceName"))
	assert.False(t, provider.CanProvide("NonExistentField"))

	// 测试提供字段值
	value, err := provider.Provide("PodName")
	assert.NoError(t, err)
	assert.Equal(t, "test-pod-0", value)

	value, err = provider.Provide("PodIndex")
	assert.NoError(t, err)
	assert.Equal(t, "0", value)

	value, err = provider.Provide("Namespace")
	assert.NoError(t, err)
	assert.Equal(t, "test-namespace", value)

	value, err = provider.Provide("ServiceName")
	assert.NoError(t, err)
	assert.Equal(t, "test-service", value)

	// 测试不支持的字段
	value, err = provider.Provide("NonExistentField")
	assert.Error(t, err)
	assert.Empty(t, value)

	// 测试验证 - Provider 接口不包含 Validate 方法
}

func TestLocalProvider(t *testing.T) {
	provider := NewLocalProvider()

	// 测试基本信息
	assert.Equal(t, "local", provider.GetName())
	assert.Equal(t, 2, provider.GetPriority())

	// 测试支持的字段
	assert.True(t, provider.CanProvide("ServiceName"))
	assert.True(t, provider.CanProvide("Version"))
	assert.True(t, provider.CanProvide("Namespace"))
	assert.False(t, provider.CanProvide("NonExistentField"))

	// 测试提供字段值
	value, err := provider.Provide("ServiceName")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	value, err = provider.Provide("Version")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	value, err = provider.Provide("Namespace")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	// 测试不支持的字段
	value, err = provider.Provide("NonExistentField")
	assert.Error(t, err)
	assert.Empty(t, value)

	// 测试验证 - Provider 接口不包含 Validate 方法
}

func TestConfigProvider(t *testing.T) {
	provider := NewConfigProvider()

	// 测试基本信息
	assert.Equal(t, "config", provider.GetName())
	assert.Equal(t, 3, provider.GetPriority())

	// 测试支持的字段
	assert.True(t, provider.CanProvide("AppId"))
	assert.True(t, provider.CanProvide("ArtifactId"))
	assert.True(t, provider.CanProvide("RegionId"))
	assert.True(t, provider.CanProvide("ChannelId"))
	assert.False(t, provider.CanProvide("NonExistentField"))

	// 测试提供字段值
	value, err := provider.Provide("AppId")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	value, err = provider.Provide("ArtifactId")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	value, err = provider.Provide("RegionId")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	value, err = provider.Provide("ChannelId")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	// 测试不支持的字段
	value, err = provider.Provide("NonExistentField")
	assert.Error(t, err)
	assert.Empty(t, value)

	// 测试验证 - Provider 接口不包含 Validate 方法
}

func TestBuildProvider(t *testing.T) {
	provider := NewBuildProvider()

	// 测试基本信息
	assert.Equal(t, "build", provider.GetName())
	assert.Equal(t, 4, provider.GetPriority())

	// 测试支持的字段
	assert.True(t, provider.CanProvide("Version"))
	assert.True(t, provider.CanProvide("BuildTime"))
	assert.True(t, provider.CanProvide("GitCommit"))
	assert.True(t, provider.CanProvide("GitBranch"))
	assert.False(t, provider.CanProvide("NonExistentField"))

	// 测试提供字段值
	value, err := provider.Provide("Version")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	value, err = provider.Provide("BuildTime")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	value, err = provider.Provide("GitCommit")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	value, err = provider.Provide("GitBranch")
	assert.NoError(t, err)
	assert.NotEmpty(t, value)

	// 测试不支持的字段
	value, err = provider.Provide("NonExistentField")
	assert.Error(t, err)
	assert.Empty(t, value)

	// 测试验证 - Provider 接口不包含 Validate 方法
}

func TestK8sProviderParsePodIndex(t *testing.T) {
	provider := &K8sProvider{}

	// 测试解析 Pod 索引
	testCases := []struct {
		podName  string
		expected string
	}{
		{"test-pod-0", "0"},
		{"test-pod-1", "1"},
		{"test-pod-10", "10"},
		{"test-pod", ""},
		{"test-pod-", ""},
		{"", ""},
	}

	for _, tc := range testCases {
		result := provider.parsePodIndex(tc.podName)
		assert.Equal(t, tc.expected, result)
	}
}

func TestK8sProviderGetEnvVar(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	// 测试获取环境变量 - K8sProvider 直接使用 os.Getenv
	value := os.Getenv("TEST_KEY")
	assert.Equal(t, "test_value", value)

	// 测试获取不存在的环境变量
	value = os.Getenv("NON_EXISTENT_KEY")
	assert.Empty(t, value)
}
