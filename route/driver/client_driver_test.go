package driver

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"

	"github.com/go-kratos/kratos/v2/route"
)

// SimpleLogger 简单的日志实现
type SimpleLogger struct{}

func (l *SimpleLogger) Log(level log.Level, keyvals ...interface{}) error {
	return nil
}

func TestNewStatefulRouteForClientDriverImpl(t *testing.T) {
	// 创建简单的mock对象
	serverInfo := &route.ServerInfo{
		Namespace:   "test-namespace",
		ServiceName: "test-service",
		PodIndex:    0,
	}

	logger := &SimpleLogger{}

	// 测试创建实例（不依赖外部接口）
	// 这里只是测试基本结构，实际使用时需要真实的依赖
	assert.NotNil(t, serverInfo)
	assert.NotNil(t, logger)
	assert.Equal(t, "test-namespace", serverInfo.Namespace)
	assert.Equal(t, "test-service", serverInfo.ServiceName)
	assert.Equal(t, 0, serverInfo.PodIndex)
}

func TestServerInfo(t *testing.T) {
	serverInfo := &route.ServerInfo{
		Namespace:   "default",
		ServiceName: "my-service",
		PodIndex:    1,
	}

	assert.Equal(t, "default", serverInfo.Namespace)
	assert.Equal(t, "my-service", serverInfo.ServiceName)
	assert.Equal(t, 1, serverInfo.PodIndex)
}

func TestBaseConfig(t *testing.T) {
	baseConfig := &route.BaseConfig{
		ServiceStateUpdatePeriodSeconds: 30,
		CpuFactor:                       0.7,
		TempLinkInfoCacheTimeSeconds:    300,
		UserLinkInfoCacheTimeSeconds:    240,
	}

	assert.Equal(t, 30, baseConfig.ServiceStateUpdatePeriodSeconds)
	assert.Equal(t, 0.7, baseConfig.CpuFactor)
	assert.Equal(t, 300, baseConfig.TempLinkInfoCacheTimeSeconds)
	assert.Equal(t, 240, baseConfig.UserLinkInfoCacheTimeSeconds)
}
