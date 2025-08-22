package driver

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/route"
)

// RouteDriverFactory 路由驱动工厂
type RouteDriverFactory struct {
	log log.Logger
}

// NewRouteDriverFactory 创建新的驱动工厂
func NewRouteDriverFactory(logger log.Logger) *RouteDriverFactory {
	return &RouteDriverFactory{
		log: logger,
	}
}

// CreateRouteInfoDriver 创建路由信息驱动
func (f *RouteDriverFactory) CreateRouteInfoDriver(
	baseConfig *route.StatefulBaseConfig,
	stateCache route.ServiceStateCache,
	statefulExecutor route.StatefulExecutor,
) route.RouteInfoDriver {
	return NewRouteInfoDriverImpl(baseConfig, stateCache, statefulExecutor, f.log)
}

// CreateRouteInfoDriverWithOptions 使用选项创建路由信息驱动
func (f *RouteDriverFactory) CreateRouteInfoDriverWithOptions(opts ...RouteDriverOption) route.RouteInfoDriver {
	options := &RouteDriverOptions{
		baseConfig:        createDefaultStatefulBaseConfig(),
		linkInfoCacheTime: 300,
		enableStateCache:  true,
		enableEventNotify: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	// 创建基础组件
	var stateCache route.ServiceStateCache
	if options.enableStateCache {
		// 这里需要创建ServiceStateCache实例
		// 暂时返回nil，需要完善组件创建逻辑
		stateCache = nil
	}

	// 创建StatefulExecutor实例
	// 这里需要根据配置创建Redis客户端和StatefulExecutor
	// 暂时返回nil，需要完善组件创建逻辑
	var statefulExecutor route.StatefulExecutor = nil

	// 创建驱动实例
	driver := NewRouteInfoDriverImpl(
		options.baseConfig,
		stateCache,
		statefulExecutor,
		f.log,
	)

	// 应用选项配置
	if options.linkInfoCacheTime > 0 {
		// 这里需要设置链接信息缓存时间
		// 暂时无法直接设置，需要添加setter方法
	}

	return driver
}

// RouteDriverOptions 驱动选项
type RouteDriverOptions struct {
	baseConfig        *route.StatefulBaseConfig
	linkInfoCacheTime int
	enableStateCache  bool
	enableEventNotify bool
}

// RouteDriverOption 驱动选项函数
type RouteDriverOption func(*RouteDriverOptions)

// WithBaseConfig 设置基础配置
func WithBaseConfig(config *route.StatefulBaseConfig) RouteDriverOption {
	return func(o *RouteDriverOptions) {
		o.baseConfig = config
	}
}

// WithLinkInfoCacheTime 设置链接信息缓存时间
func WithLinkInfoCacheTime(seconds int) RouteDriverOption {
	return func(o *RouteDriverOptions) {
		o.linkInfoCacheTime = seconds
	}
}

// WithStateCache 启用或禁用状态缓存
func WithStateCache(enable bool) RouteDriverOption {
	return func(o *RouteDriverOptions) {
		o.enableStateCache = enable
	}
}

// WithEventNotify 启用或禁用事件通知
func WithEventNotify(enable bool) RouteDriverOption {
	return func(o *RouteDriverOptions) {
		o.enableEventNotify = enable
	}
}

// DefaultRouteDriverOptions 默认驱动选项
func DefaultRouteDriverOptions() *RouteDriverOptions {
	return &RouteDriverOptions{
		baseConfig:        createDefaultStatefulBaseConfig(),
		linkInfoCacheTime: 300,
		enableStateCache:  true,
		enableEventNotify: true,
	}
}

// createDefaultStatefulBaseConfig 创建默认的有状态基础配置
func createDefaultStatefulBaseConfig() *route.StatefulBaseConfig {
	return &route.StatefulBaseConfig{
		LinkInfoCacheTimeSecs: 300,
		StateCacheExpireSecs:  300,
		MaxRetryCount:         3,
		RetryDelayMs:          1000,
		HealthCheckIntervalMs: 30000,
		LogLevel:              "INFO",
		EnableMetrics:         true,
		EnableTracing:         true,
	}
}
