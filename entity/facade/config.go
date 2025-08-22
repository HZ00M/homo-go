package facade

// Config 运行时配置（示例：与文档配置建议对应最小集）
type Config struct {
	CacheTTLMillis   int64 // entity.cache.ttl_ms
	SavePeriodMillis int64 // entity.save.period_ms
	QueueSize        int   // entity.queue.size
	KeepAliveOnGet   bool  // entity.keepalive.on_get
	RouteLinkTTLSec  int   // route.link.cache_ttl_s
	ClientMaxRetry   int   // caller.retry.max
}

// DefaultConfig 返回默认配置（便于测试）
func DefaultConfig() Config {
	return Config{
		CacheTTLMillis:   0,
		SavePeriodMillis: 0,
		QueueSize:        1024,
		KeepAliveOnGet:   true,
		RouteLinkTTLSec:  5,
		ClientMaxRetry:   1,
	}
}
